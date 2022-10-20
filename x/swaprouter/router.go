package swaprouter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// TODO: spec and tests
func (k Keeper) RouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error) {
	isGamm := true

	if isGamm {
		return k.gammKeeper.MultihopSwapExactAmountIn(ctx, sender, routes, tokenIn, tokenOutMinAmount)
	}

	return k.concentratedKeeper.MultihopSwapExactAmountIn(ctx, sender, routes, tokenIn, tokenOutMinAmount)
}

// TODO: spec and tests
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin) (tokenInAmount sdk.Int, err error) {
	isGamm := true

	if isGamm {
		return k.gammKeeper.MultihopSwapExactAmountOut(ctx, sender, routes, tokenInMaxAmount, tokenOut)
	}

	return k.concentratedKeeper.MultihopSwapExactAmountOut(ctx, sender, routes, tokenInMaxAmount, tokenOut)
}
