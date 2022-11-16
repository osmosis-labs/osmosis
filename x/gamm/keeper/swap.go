package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// SwapExactAmountIn attempts to swap one asset, tokenIn, for another asset
// denominated via tokenOutDenom through a pool denoted by poolId specifying that
// tokenOutMinAmount must be returned in the resulting asset returning an error
// upon failure. Upon success, the resulting tokens swapped for are returned. A
// swap fee is applied determined by the pool's parameters.
// The estimateSwap parameter is used to determine whether the swap is an estimate or not.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	estimateSwap bool,
) (sdk.Int, error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	swapFee := pool.GetSwapFee(ctx)

	// if the swap is an estimate swap, we want to use cache ctx to avoid mutating state
	return k.swapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee, estimateSwap, estimateSwap)
}

// swapExactAmountIn is an internal method for swapping an exact amount of tokens
// as input to a pool, using the provided swapFee. This is intended to allow
// different swap fees as determined by multi-hops, or when recovering from
// chain liveness failures.
// If useCacheCtx is true, then the swap will be performed on a cached context.
// If the estimateSwap parameter is true, then the swap will not send actual coins upon swap.
func (k Keeper) swapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
	useCacheCtx bool,
	estimateSwap bool,
) (tokenOutAmount sdk.Int, err error) {
	if useCacheCtx {
		ctx, _ = ctx.CacheContext()
	}

	tokenOut, err := k.swapExactAmountInNoTokenSend(ctx, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, tokenIn, tokenOut); err != nil {
		return sdk.Int{}, err
	}

	if !estimateSwap {
		err = k.sendCoinsAfterSwap(ctx, pool, tokenIn, tokenOut, sender)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	return tokenOut.Amount, nil
}

// swapExactAmountInNoTokenSend performs `SwapOutAmtGivenIn` for the given inputs,
// and potentially mutates state. It does not save the new pool struct, or do token transfers.
// Returns the altered pool after swap, and the token out when swapped.
func (k Keeper) swapExactAmountInNoTokenSend(
	ctx sdk.Context,
	pool types.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (tokenOut sdk.Coin, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Coin{}, errors.New("cannot trade same denomination in and out")
	}
	tokensIn := sdk.Coins{tokenIn}

	defer func() {
		if r := recover(); r != nil {
			tokenOut = sdk.Coin{}
			err = fmt.Errorf("function swapExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	// Executes the swap in the pool and stores the output. Updates pool assets but
	// does not actually transfer any tokens to or from the pool.
	tokenOutCoin, err := pool.SwapOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	tokenOutAmount := tokenOutCoin.Amount

	if !tokenOutAmount.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", tokenOutDenom)
	}

	return tokenOutCoin, nil
}

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	estimateSwap bool,
) (tokenInAmount sdk.Int, err error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}
	swapFee := pool.GetSwapFee(ctx)
	return k.swapExactAmountOut(ctx, sender, pool, tokenInDenom, tokenInMaxAmount, tokenOut, swapFee, estimateSwap, estimateSwap)
}

// swapExactAmountOut is an internal method for swapping to get an exact number of tokens out of a pool,
// using the provided swapFee.
// This is intended to allow different swap fees as determined by multi-hops,
// or when recovering from chain liveness failures.
// If useCacheCtx is true, then the swap will be performed on a cached context.
// If the estimateSwap parameter is true, then the swap will not send actual coins upon swap.
func (k Keeper) swapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool types.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
	useCacheCtx bool,
	estimate bool,
) (tokenInAmount sdk.Int, err error) {
	if useCacheCtx {
		ctx, _ = ctx.CacheContext()
	}

	tokenIn, err := k.swapExactAmountOutNoTokenSend(ctx, pool, tokenInDenom, tokenInMaxAmount, tokenOut, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.updatePoolForSwap(ctx, pool, tokenIn, tokenOut)
	if err != nil {
		return sdk.Int{}, err
	}

	if !estimate {
		err = k.sendCoinsAfterSwap(ctx, pool, tokenIn, tokenOut, sender)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	return tokenIn.Amount, nil
}

// swapExactAmountOutNoTokenSend performs `SwapInAmtGivenOut` for the given inputs,
// and potentially mutates state. It does not save the new pool struct, or do token transfers.
// Returns the altered pool after swap, and the token in when swapped.
func (k Keeper) swapExactAmountOutNoTokenSend(
	ctx sdk.Context,
	pool types.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenIn sdk.Coin, err error) {
	if tokenInDenom == tokenOut.Denom {
		return sdk.Coin{}, errors.New("cannot trade same denomination in and out")
	}

	defer func() {
		if r := recover(); r != nil {
			tokenIn = sdk.Coin{}
			err = fmt.Errorf("function swapExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	poolOutBal := pool.GetTotalPoolLiquidity(ctx).AmountOf(tokenOut.Denom)
	if tokenOut.Amount.GTE(poolOutBal) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrTooManyTokensOut,
			"can't get more tokens out than there are tokens in the pool")
	}

	tokenIn, err = pool.SwapInAmtGivenOut(ctx, sdk.Coins{tokenOut}, tokenInDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}
	tokenInAmount := tokenIn.Amount

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "Swap requires %s, which is greater than the amount %s", tokenIn, tokenInMaxAmount)
	}

	return tokenIn, nil
}

// updatePoolForSwap updates the pool balance and the pool state.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool types.PoolI,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
) error {
	tokensIn := sdk.Coins{tokenIn}
	tokensOut := sdk.Coins{tokenOut}

	err := k.setPool(ctx, pool)
	if err != nil {
		return err
	}

	k.RecordTotalLiquidityIncrease(ctx, tokensIn)
	k.RecordTotalLiquidityDecrease(ctx, tokensOut)

	return nil
}

// sendCoinsAfterSwap sends the coins after a swap has been executed.
// this method also includes hooks and emitting events.
func (k Keeper) sendCoinsAfterSwap(
	ctx sdk.Context,
	pool types.PoolI,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
	sender sdk.AccAddress,
) error {
	err := k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{
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

	tokensIn := sdk.Coins{tokenIn}
	tokensOut := sdk.Coins{tokenOut}
	events.EmitSwapEvent(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.hooks.AfterSwap(ctx, sender, pool.GetId(), tokensIn, tokensOut)

	return nil
}
