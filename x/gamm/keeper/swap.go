package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// SwapExactAmountIn attempts to swap one asset, tokenIn, for another asset
// denominated via tokenOutDenom through a pool denoted by poolId specifying that
// tokenOutMinAmount must be returned in the resulting asset returning an error
// upon failure. Upon success, the resulting tokens swapped for are returned. A
// swap fee is applied determined by the pool's parameters.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
) (sdk.Int, error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	swapFee := pool.GetSwapFee(ctx)
	return k.swapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
}

// swapExactAmountIn is an internal method for swapping an exact amount of tokens
// as input to a pool, using the provided swapFee. This is intended to allow
// different swap fees as determined by multi-hops, or when recovering from
// chain liveness failures.
func (k Keeper) swapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (tokenOutAmount sdk.Int, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Int{}, errors.New("cannot trade same denomination in and out")
	}
	tokensIn := sdk.Coins{tokenIn}

	// Executes the swap in the pool and stores the output. Updates pool assets but
	// does not actually transfer any tokens to or from the pool.
	tokenOutCoin, err := pool.SwapOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	tokenOutAmount = tokenOutCoin.Amount

	if !tokenOutAmount.IsPositive() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", tokenOutDenom)
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOutCoin); err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}
	swapFee := pool.GetSwapFee(ctx)
	return k.swapExactAmountOut(ctx, sender, pool, tokenInDenom, tokenInMaxAmount, tokenOut, swapFee)
}

// swapExactAmountIn is an internal method for swapping to get an exact number of tokens out of a pool,
// using the provided swapFee.
// This is intended to allow different swap fees as determined by multi-hops,
// or when recovering from chain liveness failures.
func (k Keeper) swapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	if tokenInDenom == tokenOut.Denom {
		return sdk.Int{}, errors.New("cannot trade same denomination in and out")
	}

	poolOutBal := pool.GetTotalPoolLiquidity(ctx).AmountOf(tokenOut.Denom)
	if tokenOut.Amount.GTE(poolOutBal) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrTooManyTokensOut,
			"can't get more tokens out than there are tokens in the pool")
	}
	tokenIn, err := pool.SwapInAmtGivenOut(ctx, sdk.Coins{tokenOut}, tokenInDenom, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}
	tokenInAmount = tokenIn.Amount

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "Swap requires %s, which is greater than the amount %s", tokenIn, tokenInMaxAmount)
	}

	err = k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut)
	if err != nil {
		return sdk.Int{}, err
	}
	return tokenInAmount, nil
}

// updatePoolForSwap takes a pool, sender, and tokenIn, tokenOut amounts
// It then updates the pool's balances to the new reserve amounts, and
// sends the in tokens from the sender to the pool, and the out tokens from the pool to the sender.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool types.PoolI,
	sender sdk.AccAddress,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
) error {
	tokensIn := sdk.Coins{tokenIn}
	tokensOut := sdk.Coins{tokenOut}

	err := k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{
		tokenIn,
	})
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return err
	}

	events.EmitSwapEvent(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.hooks.AfterSwap(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.RecordTotalLiquidityIncrease(ctx, tokensIn)
	k.RecordTotalLiquidityDecrease(ctx, tokensOut)

	return err
}
