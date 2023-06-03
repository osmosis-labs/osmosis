package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v16/x/poolmanager/events"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

// swapExactAmountIn is an internal method for swapping an exact amount of tokens
// as input to a pool, using the provided spreadFactor. This is intended to allow
// different spread factors as determined by multi-hops, or when recovering from
// chain liveness failures.
// TODO: investigate if spreadFactor can be unexported
// https://github.com/osmosis-labs/osmosis/issues/3130
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	spreadFactor sdk.Dec,
) (tokenOutAmount sdk.Int, err error) {
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Int{}, errors.New("cannot trade same denomination in and out")
	}
	poolSpreadFactor := pool.GetSpreadFactor(ctx)
	if spreadFactor.LT(poolSpreadFactor.QuoInt64(2)) {
		return sdk.Int{}, fmt.Errorf("given spread factor (%s) must be greater than or equal to half of the pool's spread factor (%s)", spreadFactor, poolSpreadFactor)
	}
	tokensIn := sdk.Coins{tokenIn}

	defer func() {
		if r := recover(); r != nil {
			tokenOutAmount = sdk.Int{}
			err = fmt.Errorf("function swapExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	cfmmPool, err := asCFMMPool(pool)
	if err != nil {
		return sdk.Int{}, err
	}

	// Executes the swap in the pool and stores the output. Updates pool assets but
	// does not actually transfer any tokens to or from the pool.
	tokenOutCoin, err := cfmmPool.SwapOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}

	tokenOutAmount = tokenOutCoin.Amount

	if !tokenOutAmount.IsPositive() {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", tokenOutDenom)
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOutCoin); err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

// SwapExactAmountOut is a method for swapping to get an exact number of tokens out of a pool,
// using the provided spreadFactor.
// This is intended to allow different spread factors as determined by multi-hops,
// or when recovering from chain liveness failures.
func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool poolmanagertypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	spreadFactor sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	if tokenInDenom == tokenOut.Denom {
		return sdk.Int{}, errors.New("cannot trade same denomination in and out")
	}
	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = sdk.Int{}
			err = fmt.Errorf("function swapExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	liquidity, err := k.GetTotalPoolLiquidity(ctx, pool.GetId())
	if err != nil {
		return sdk.Int{}, err
	}

	poolOutBal := liquidity.AmountOf(tokenOut.Denom)
	if tokenOut.Amount.GTE(poolOutBal) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrTooManyTokensOut,
			"can't get more tokens out than there are tokens in the pool")
	}

	cfmmPool, err := asCFMMPool(pool)
	if err != nil {
		return sdk.Int{}, err
	}

	tokenIn, err := cfmmPool.SwapInAmtGivenOut(ctx, sdk.Coins{tokenOut}, tokenInDenom, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}
	tokenInAmount = tokenIn.Amount

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrLimitMaxAmount, "Swap requires %s, which is greater than the amount %s", tokenIn, tokenInMaxAmount)
	}

	err = k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut)
	if err != nil {
		return sdk.Int{}, err
	}
	return tokenInAmount, nil
}

// CalcOutAmtGivenIn calculates the amount of tokenOut given tokenIn and the pool's current state.
// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
func (k Keeper) CalcOutAmtGivenIn(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	spreadFactor sdk.Dec,
) (tokenOut sdk.Coin, err error) {
	cfmmPool, err := asCFMMPool(poolI)
	if err != nil {
		return sdk.Coin{}, err
	}
	return cfmmPool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(tokenIn), tokenOutDenom, spreadFactor)
}

// CalcInAmtGivenOut calculates the amount of tokenIn given tokenOut and the pool's current state.
// Returns error if the given pool is not a CFMM pool. Returns error on internal calculations.
func (k Keeper) CalcInAmtGivenOut(
	ctx sdk.Context,
	poolI poolmanagertypes.PoolI,
	tokenOut sdk.Coin,
	tokenInDenom string,
	spreadFactor sdk.Dec,
) (tokenIn sdk.Coin, err error) {
	cfmmPool, err := asCFMMPool(poolI)
	if err != nil {
		return sdk.Coin{}, err
	}
	return cfmmPool.CalcInAmtGivenOut(ctx, sdk.NewCoins(tokenOut), tokenInDenom, spreadFactor)
}

// updatePoolForSwap takes a pool, sender, and tokenIn, tokenOut amounts
// It then updates the pool's balances to the new reserve amounts, and
// sends the in tokens from the sender to the pool, and the out tokens from the pool to the sender.
func (k Keeper) updatePoolForSwap(
	ctx sdk.Context,
	pool poolmanagertypes.PoolI,
	sender sdk.AccAddress,
	tokenIn sdk.Coin,
	tokenOut sdk.Coin,
) error {
	tokensIn := sdk.Coins{tokenIn}
	tokensOut := sdk.Coins{tokenOut}

	err := k.setPool(ctx, pool)
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
	k.hooks.AfterCFMMSwap(ctx, sender, pool.GetId(), tokensIn, tokensOut)
	k.RecordTotalLiquidityIncrease(ctx, tokensIn)
	k.RecordTotalLiquidityDecrease(ctx, tokensOut)

	return err
}
