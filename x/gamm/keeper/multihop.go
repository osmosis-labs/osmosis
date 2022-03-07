package keeper

import (
	"errors"

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
	for _, route := range routes {
		pool, err := k.GetPool(ctx, route.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}
		tokenOutAmount, err = pool.CalcOutAmtGivenIn(ctx, tokenIn, route.TokenOutDenom, pool.GetPoolSwapFee())
		if err != nil {
			return sdk.Int{}, err
		}

		tokenOut := sdk.NewDecCoin(route.TokenOutDenom, tokenOutAmount)
		// Note to self, make DoTransfersForSwap signal to underlying AMM that swap happening
		err = k.DoTransfersForSwap(ctx, sender, route.PoolId, tokenIn, tokenOut)
		if err != nil {
			return sdk.Int{}, err
		}

		tokenIn = tokenOut
	}
	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, errors.New("...")
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
	panic("wave")
	// insExpected, err := k.createMultihopExpectedSwapOuts(ctx, routes, tokenOut)
	// if err != nil {
	// 	return sdk.Int{}, err
	// }

	// insExpected[0] = tokenInMaxAmount

	// for i, route := range routes {
	// 	_tokenOut := tokenOut
	// 	if i != len(routes)-1 {
	// 		_tokenOut = sdk.NewCoin(routes[i+1].TokenInDenom, insExpected[i+1])
	// 	}

	// 	_tokenInAmount, _, err := k.SwapExactAmountOut(ctx, sender, route.PoolId, route.TokenInDenom, insExpected[i], _tokenOut)
	// 	if err != nil {
	// 		return sdk.Int{}, err
	// 	}

	// 	if i == 0 {
	// 		tokenInAmount = _tokenInAmount
	// 	}
	// }

	// return
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

		tokenInAmount := pool.CalcInAmtGivenOut(
			ctx,
			tokenInDenom,
			tokenOut.Coin,
			pool.GetPoolSwapFee(),
		).TruncateInt()

		insExpected[i] = tokenInAmount

		tokenOut = sdk.NewCoin(route.TokenInDenom, tokenInAmount)
	}

	return insExpected, nil
}
