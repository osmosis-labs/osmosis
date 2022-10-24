package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
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
