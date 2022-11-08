package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func (k Keeper) CalcOutAmtGivenIn(ctx sdk.Context,
	tokenInMin sdk.Coin, tokenOutDenom string,
	swapFee sdk.Dec, priceLimit sdk.Dec,
	poolId uint64) (tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	pool, err := k.getPoolbyId(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Int{}, sdk.Dec{}, sdk.Dec{}, err
	}
	poolTickKVStore := k.GetPoolTickKVStore(ctx, poolId)
	return pool.CalcOutAmtGivenIn(ctx, poolTickKVStore, tokenInMin, tokenOutDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) SwapOutAmtGivenIn(ctx sdk.Context,
	tokenIn sdk.Coin, tokenOutDenom string,
	swapFee sdk.Dec, priceLimit sdk.Dec,
	poolId uint64) (tokenOut sdk.Coin, err error) {
	pool, err := k.getPoolbyId(ctx, poolId)
	if err != nil {
		return sdk.Coin{}, err
	}
	poolTickKVStore := k.GetPoolTickKVStore(ctx, poolId)
	return pool.SwapOutAmtGivenIn(ctx, poolTickKVStore, tokenIn, tokenOutDenom, swapFee, priceLimit, poolId)

}

// TODO: spec here and in gamm
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool gammtypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (tokenOutAmount sdk.Int, err error) {
	panic("not implemented")
	//newTokenIn, tokenOut, err := k.CalcOutAmtGivenIn(ctx, tokenIn, tokenOutDenom, swapFee, sdk.ZeroDec(), sdk.NewDec(999999999999), pool.GetId())
}

// TODO: spec here and in gamm
func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolI gammtypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	panic("not implemented")
}
