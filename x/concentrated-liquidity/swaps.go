package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

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
