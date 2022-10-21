package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// TODO: godoc
func (k Keeper) MultihopSwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []swaproutertypes.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	return sdk.Int{}, nil
}

// TODO: godoc
func (k Keeper) MultihopSwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []swaproutertypes.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	return sdk.Int{}, nil
}
