package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// MultihopSwapExactAmountIn defines the input denom and input amount for the first pool,
// the output of the first pool is chained as the input for the next routed pool
// transaction succeeds when final amount out is greater than tokenOutMinAmount defined
func (k Keeper) MultihopSwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	for i, route := range routes {
		_outMinAmount := sdk.NewInt(1)
		if len(routes)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		tokenOutAmount, _, err = k.SwapExactAmountIn(ctx, sender, route.PoolId, tokenIn, route.TokenOutDenom, _outMinAmount)
		if err != nil {
			return sdk.Int{}, err
		}
		tokenIn = sdk.NewCoin(route.TokenOutDenom, tokenOutAmount)
	}
	return
}

// MultihopSwapExactAmountOut defines the output denom and output amount for the last pool.
// Calculation starts by providing the tokenOutAmount of the final pool to calculate the required tokenInAmount
// the calculated tokenInAmount is used as defined tokenOutAmount of the previous pool, calculating in reverse order of the swap
// Transaction succeeds if the calculated tokenInAmount of the first pool is less than the defined tokenInMaxAmount defined.
func (k Keeper) MultihopSwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	insExpected, err := k.createMultihopExpectedSwapOuts(ctx, routes, tokenOut)
	if err != nil {
		return sdk.Int{}, err
	}

	insExpected[0] = tokenInMaxAmount

	for i, route := range routes {
		_tokenOut := tokenOut
		if i != len(routes)-1 {
			_tokenOut = sdk.NewCoin(routes[i+1].TokenInDenom, insExpected[i+1])
		}

		_tokenInAmount, _, err := k.SwapExactAmountOut(ctx, sender, route.PoolId, route.TokenInDenom, insExpected[i], _tokenOut)
		if err != nil {
			return sdk.Int{}, err
		}

		if i == 0 {
			tokenInAmount = _tokenInAmount
		}
	}

	return
}

// TODO: Document this function
func (k Keeper) createMultihopExpectedSwapOuts(ctx sdk.Context, routes []types.SwapAmountOutRoute, tokenOut sdk.Coin) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(routes))
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]

		pool, inAsset, outAsset, err :=
			k.getPoolAndInOutAssets(ctx, route.PoolId, route.TokenInDenom, tokenOut.Denom)
		if err != nil {
			return nil, err
		}

		tokenInAmount := calcInGivenOut(
			inAsset.Token.Amount.ToDec(),
			inAsset.Weight.ToDec(),
			outAsset.Token.Amount.ToDec(),
			outAsset.Weight.ToDec(),
			tokenOut.Amount.ToDec(),
			pool.GetPoolSwapFee(),
		).TruncateInt()

		insExpected[i] = tokenInAmount

		tokenOut = sdk.NewCoin(route.TokenInDenom, tokenInAmount)
	}

	return insExpected, nil
}
