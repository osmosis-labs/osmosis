package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// MultihopSwapExactAmountIn defines the input denom and input amount for the first pool,
// the output of the first pool is chained as the input for the next routed pool
// transaction succeeds when final amount out is greater than tokenOutMinAmount defined.
func (k Keeper) MultihopSwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	route0Incentivized, route1Incentivized, additiveSwapFee, maxSwapFee, err := k.osmoRouteFeeCalcIn(ctx, routes)

	for i, route := range routes {
		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := sdk.NewInt(1)
		if len(routes)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := k.getPoolForSwap(ctx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		swapFee := pool.GetSwapFee(ctx)

		// if we determined the route is an osmo multi-hop and both routes are incentivized,
		// we modify the swap fee accordingly
		if route0Incentivized && route1Incentivized {
			swapFee = maxSwapFee.Mul((swapFee.Quo(additiveSwapFee)))
		}

		tokenOutAmount, err = k.swapExactAmountIn(ctx, sender, pool, tokenIn, route.TokenOutDenom, _outMinAmount, swapFee)
		if err != nil {
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(route.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, err
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
	route0Incentivized, route1Incentivized, additiveSwapFee, maxSwapFee, err := k.osmoRouteFeeCalcOut(ctx, routes)
	if err != nil {
		return sdk.Int{}, err
	}

	// Determine what the estimated input would be for each pool along the multi-hop route
	// if we determined the route is an osmo multi-hop and both routes are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	var insExpected []sdk.Int
	if route0Incentivized && route1Incentivized {
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, routes, tokenOut, additiveSwapFee, maxSwapFee)
	} else {
		insExpected, err = k.createMultihopExpectedSwapOuts(ctx, routes, tokenOut)
	}
	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}

	insExpected[0] = tokenInMaxAmount

	// Iterates through each routed pool and executes their respective swaps. Note that all of the work to get the return
	// value of this method is done when we calculate insExpected – this for loop primarily serves to execute the actual
	// swaps on each pool.
	for i, route := range routes {
		_tokenOut := tokenOut

		// If there is one pool left in the route, set the expected output of the current swap
		// to the estimated input of the final pool.
		if i != len(routes)-1 {
			_tokenOut = sdk.NewCoin(routes[i+1].TokenInDenom, insExpected[i+1])
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := k.getPoolForSwap(ctx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}
		swapFee := pool.GetSwapFee(ctx)
		if route0Incentivized && route1Incentivized {
			swapFee = maxSwapFee.Mul((swapFee.Quo(additiveSwapFee)))
		}
		_tokenInAmount, swapErr := k.swapExactAmountOut(ctx, sender, pool, route.TokenInDenom, insExpected[i], _tokenOut, swapFee)
		if swapErr != nil {
			return sdk.Int{}, swapErr
		}

		// Sets the final amount of tokens that need to be input into the first pool. Even though this is the final return value for the
		// whole method and will not change after the first iteration, we still iterate through the rest of the pools to execute their respective
		// swaps.
		if i == 0 {
			tokenInAmount = _tokenInAmount
		}
	}

	return tokenInAmount, nil
}

// createMultihopExpectedSwapOuts defines the output denom and output amount for the last pool in
// the route of pools the caller is intending to hop through in a fixed-output multihop tx. It estimates the input
// amount for this last pool and then chains that input as the output of the previous pool in the route, repeating
// until the first pool is reached. It returns an array of inputs, each of which correspond to a pool ID in the
// route of pools for the original multihop transaction.
func (k Keeper) createMultihopExpectedSwapOuts(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(routes))
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]

		pool, err := k.getPoolForSwap(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		tokenIn, err := pool.CalcInAmtGivenOut(ctx, sdk.NewCoins(tokenOut), route.TokenInDenom, pool.GetSwapFee(ctx))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

// createOsmoMultihopExpectedSwapOuts does the same as createMultihopExpectedSwapOuts, however discounts the swap fee
func (k Keeper) createOsmoMultihopExpectedSwapOuts(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
	additiveSwapFee, maxSwapFee sdk.Dec,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(routes))
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]

		pool, err := k.getPoolForSwap(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		swapFee := pool.GetSwapFee(ctx)
		tokenIn, err := pool.CalcInAmtGivenOut(ctx, sdk.NewCoins(tokenOut), route.TokenInDenom, maxSwapFee.Mul((swapFee.Quo(additiveSwapFee))))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

func (k Keeper) osmoRouteFeeCalcIn(
	ctx sdk.Context,
	routes []types.SwapAmountInRoute,
) (route0Incentivized, route1Incentivized bool, additiveSwapFee, maxSwapFee sdk.Dec, err error) {
	additiveSwapFee = sdk.ZeroDec()
	maxSwapFee = sdk.ZeroDec()

	// get list of all incentivized pools
	incentivizedPools := k.poolIncentivesKeeper.GetAllIncentivizedPools(ctx)

	// in this loop, we check if:
	// - the route is of length 2
	// - route 1 and route 2 don't trade via the same pool
	// - route 1 contains uosmo
	// - both route 1 and route 2 are incentivized pools
	// if all of the above is true, then we collect the additive and max fee between the two pools to later calculate the following:
	// total_swap_fee = total_swap_fee = max(swapfee1, swapfee2)
	// fee_per_pool = total_swap_fee * ((pool_fee) / (swapfee1 + swapfee2))
	if types.SwapAmountInRoutes(routes).IsOsmoRoutedMultihop() {
		for _, route := range routes {
			pool, poolErr := k.getPoolForSwap(ctx, route.PoolId)
			if poolErr != nil {
				return false, false, sdk.Dec{}, sdk.Dec{}, poolErr
			}
			swapFee := pool.GetSwapFee(ctx)
			additiveSwapFee = additiveSwapFee.Add(swapFee)
			if swapFee.GT(maxSwapFee) {
				maxSwapFee = swapFee
			}
			for _, pool := range incentivizedPools {
				if routes[0].PoolId == pool.PoolId {
					route0Incentivized = true
				}
				if routes[1].PoolId == pool.PoolId {
					route1Incentivized = true
				}
				if route0Incentivized && route1Incentivized {
					break
				}
			}
		}
	}

	if additiveSwapFee.Quo(sdk.NewDec(2)).GT(maxSwapFee) {
		maxSwapFee = additiveSwapFee.Quo(sdk.NewDec(2))
	}

	return route0Incentivized, route1Incentivized, additiveSwapFee, maxSwapFee, nil
}

func (k Keeper) osmoRouteFeeCalcOut(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
) (route0Incentivized, route1Incentivized bool, additiveSwapFee, maxSwapFee sdk.Dec, err error) {
	additiveSwapFee = sdk.ZeroDec()
	maxSwapFee = sdk.ZeroDec()

	// get list of all incentivized pools
	incentivizedPools := k.poolIncentivesKeeper.GetAllIncentivizedPools(ctx)

	// in this loop, we check if:
	// - the route is of length 2
	// - route 1 and route 2 don't trade via the same pool
	// - route 1 contains uosmo
	// - both route 1 and route 2 are incentivized pools
	// if all of the above is true, then we collect the additive and max fee between the two pools to later calculate the following:
	// total_swap_fee = total_swap_fee = max(swapfee1, swapfee2)
	// fee_per_pool = total_swap_fee * ((pool_fee) / (swapfee1 + swapfee2))
	if types.SwapAmountOutRoutes(routes).IsOsmoRoutedMultihop() {
		for _, route := range routes {
			pool, poolErr := k.getPoolForSwap(ctx, route.PoolId)
			if poolErr != nil {
				return false, false, sdk.Dec{}, sdk.Dec{}, poolErr
			}
			swapFee := pool.GetSwapFee(ctx)
			additiveSwapFee = additiveSwapFee.Add(swapFee)
			if swapFee.GT(maxSwapFee) {
				maxSwapFee = swapFee
			}
			for _, pool := range incentivizedPools {
				if routes[0].PoolId == pool.PoolId {
					route0Incentivized = true
				}
				if routes[1].PoolId == pool.PoolId {
					route1Incentivized = true
				}
				if route0Incentivized && route1Incentivized {
					break
				}
			}
		}
	}

	if additiveSwapFee.Quo(sdk.NewDec(2)).GT(maxSwapFee) {
		maxSwapFee = additiveSwapFee.Quo(sdk.NewDec(2))
	}

	return route0Incentivized, route1Incentivized, additiveSwapFee, maxSwapFee, nil
}
