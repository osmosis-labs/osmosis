package swaprouter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// RouteExactAmountIn defines the input denom and input amount for the first pool,
// the output of the first pool is chained as the input for the next routed pool
// transaction succeeds when final amount out is greater than tokenOutMinAmount defined.
func (k Keeper) RouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error) {
	panic("not implemented")
}

func (k Keeper) MultihopEstimateOutGivenExactAmountIn(
	ctx sdk.Context,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount sdk.Int, err error) {
	panic("not implemented")
}

// MultihopSwapExactAmountOut defines the output denom and output amount for the last pool.
// Calculation starts by providing the tokenOutAmount of the final pool to calculate the required tokenInAmount
// the calculated tokenInAmount is used as defined tokenOutAmount of the previous pool, calculating in reverse order of the swap
// Transaction succeeds if the calculated tokenInAmount of the first pool is less than the defined tokenInMaxAmount defined.
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	panic("not implemented")
}

func (k Keeper) MultihopEstimateInGivenExactAmountOut(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	panic("not implemented")
}
