package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/gamm/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"
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
	pool, tokenOut, err := k.estimateSwapExactAmountIn(ctx, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	// Settles balances between the tx sender and the pool to match the swap that was executed earlier.
	// Also emits swap event and updates related liquidity metrics
	if err := k.updatePoolForSwap(ctx, pool, sender, tokenIn, tokenOut); err != nil {
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
	pool, tokenIn, err := k.estimateSwapExactAmountOut(ctx, pool, tokenInDenom, tokenInMaxAmount, tokenOut, swapFee)
	if err != nil {
		return sdk.Int{}, err
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
	k.hooks.AfterSwap(ctx, sender, pool.GetId(), tokensIn, tokensOut)

	k.RecordTotalLiquidityIncrease(ctx, tokensIn)
	k.RecordTotalLiquidityDecrease(ctx, tokensOut)

	return err
}
