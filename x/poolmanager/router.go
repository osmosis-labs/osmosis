package poolmanager

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v15/app/params"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// RouteExactAmountIn defines the input denom and input amount for the first pool,
// the output of the first pool is chained as the input for the next routed pool
// transaction succeeds when final amount out is greater than tokenOutMinAmount defined.
func (k Keeper) RouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	var (
		isMultiHopRouted bool
		routeSwapFee     sdk.Dec
		sumOfSwapFees    sdk.Dec
	)

	route := types.SwapAmountInRoutes(routes)
	if err := route.Validate(); err != nil {
		return sdk.Int{}, err
	}

	// In this loop, we check if:
	// - the route is of length 2
	// - route 1 and route 2 don't trade via the same pool
	// - route 1 contains uosmo
	// - both route 1 and route 2 are incentivized pools
	//
	// If all of the above is true, then we collect the additive and max fee between the
	// two pools to later calculate the following:
	// total_swap_fee = total_swap_fee = max(swapfee1, swapfee2)
	// fee_per_pool = total_swap_fee * ((pool_fee) / (swapfee1 + swapfee2))
	if k.isOsmoRoutedMultihop(ctx, route, routes[0].TokenOutDenom, tokenIn.Denom) {
		isMultiHopRouted = true
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, route)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	for i, route := range routes {
		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := sdk.NewInt(1)
		if len(routes)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		swapModule, err := k.GetPoolModule(ctx, route.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		// check if pool is active, if not error
		if !pool.IsActive(ctx) {
			return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
		}

		swapFee := pool.GetSwapFee(ctx)

		// If we determined the route is an osmo multi-hop and both routes are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			swapFee = routeSwapFee.Mul((swapFee.Quo(sumOfSwapFees)))
		}

		tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, route.TokenOutDenom, _outMinAmount, swapFee)
		if err != nil {
			ctx.Logger().Error(err.Error())
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(route.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, nil
}

// SwapExactAmountIn is an API for swapping an exact amount of tokens
// as input to a pool to get a minimum amount of the desired token out.
// The method succeeds when tokenOutAmount is greater than tokenOutMinAmount defined.
// Errors otherwise. Also, errors if the pool id is invalid, if tokens do not belong to the pool with given
// id or if sender does not have the swapped-in tokenIn.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	pool, poolErr := swapModule.GetPool(ctx, poolId)
	if poolErr != nil {
		return sdk.Int{}, poolErr
	}

	// check if pool is active, if not error
	if !pool.IsActive(ctx) {
		return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	swapFee := pool.GetSwapFee(ctx)

	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) MultihopEstimateOutGivenExactAmountIn(
	ctx sdk.Context,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount sdk.Int, err error) {
	var (
		isMultiHopRouted bool
		routeSwapFee     sdk.Dec
		sumOfSwapFees    sdk.Dec
	)

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			tokenOutAmount = sdk.Int{}
			err = fmt.Errorf("function MultihopEstimateOutGivenExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	route := types.SwapAmountInRoutes(routes)
	if err := route.Validate(); err != nil {
		return sdk.Int{}, err
	}

	if k.isOsmoRoutedMultihop(ctx, route, routes[0].TokenOutDenom, tokenIn.Denom) {
		isMultiHopRouted = true
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, route)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	for _, route := range routes {
		swapModule, err := k.GetPoolModule(ctx, route.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		poolI, poolErr := swapModule.GetPool(ctx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		swapFee := poolI.GetSwapFee(ctx)

		// If we determined the route is an osmo multi-hop and both routes are incentivized,
		// we modify the swap fee accordingly.
		if isMultiHopRouted {
			swapFee = routeSwapFee.Mul((swapFee.Quo(sumOfSwapFees)))
		}

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, tokenIn, route.TokenOutDenom, swapFee)
		if err != nil {
			return sdk.Int{}, err
		}

		tokenOutAmount = tokenOut.Amount
		if !tokenOutAmount.IsPositive() {
			return sdk.Int{}, errors.New("token amount must be positive")
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
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	isMultiHopRouted, routeSwapFee, sumOfSwapFees := false, sdk.Dec{}, sdk.Dec{}
	route := types.SwapAmountOutRoutes(routes)
	if err := route.Validate(); err != nil {
		return sdk.Int{}, err
	}

	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = sdk.Int{}
			err = fmt.Errorf("function RouteExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	// in this loop, we check if:
	// - the route is of length 2
	// - route 1 and route 2 don't trade via the same pool
	// - route 1 contains uosmo
	// - both route 1 and route 2 are incentivized pools
	// if all of the above is true, then we collect the additive and max fee between the two pools to later calculate the following:
	// total_swap_fee = total_swap_fee = max(swapfee1, swapfee2)
	// fee_per_pool = total_swap_fee * ((pool_fee) / (swapfee1 + swapfee2))
	if k.isOsmoRoutedMultihop(ctx, route, routes[0].TokenInDenom, tokenOut.Denom) {
		isMultiHopRouted = true
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, route)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	// Determine what the estimated input would be for each pool along the multi-hop route
	// if we determined the route is an osmo multi-hop and both routes are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	var insExpected []sdk.Int
	if isMultiHopRouted {
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, routes, tokenOut, routeSwapFee, sumOfSwapFees)
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
		swapModule, err := k.GetPoolModule(ctx, route.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		_tokenOut := tokenOut

		// If there is one pool left in the route, set the expected output of the current swap
		// to the estimated input of the final pool.
		if i != len(routes)-1 {
			_tokenOut = sdk.NewCoin(routes[i+1].TokenInDenom, insExpected[i+1])
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		// check if pool is active, if not error
		if !pool.IsActive(ctx) {
			return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
		}

		swapFee := pool.GetSwapFee(ctx)
		if isMultiHopRouted {
			swapFee = routeSwapFee.Mul((swapFee.Quo(sumOfSwapFees)))
		}

		_tokenInAmount, swapErr := swapModule.SwapExactAmountOut(ctx, sender, pool, route.TokenInDenom, insExpected[i], _tokenOut, swapFee)
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

func (k Keeper) MultihopEstimateInGivenExactAmountOut(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	isMultiHopRouted, routeSwapFee, sumOfSwapFees := false, sdk.Dec{}, sdk.Dec{}
	var insExpected []sdk.Int

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			insExpected = []sdk.Int{}
			err = fmt.Errorf("function MultihopEstimateInGivenExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	route := types.SwapAmountOutRoutes(routes)
	if err := route.Validate(); err != nil {
		return sdk.Int{}, err
	}

	if k.isOsmoRoutedMultihop(ctx, route, routes[0].TokenInDenom, tokenOut.Denom) {
		isMultiHopRouted = true
		routeSwapFee, sumOfSwapFees, err = k.getOsmoRoutedMultihopTotalSwapFee(ctx, route)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	// Determine what the estimated input would be for each pool along the multi-hop route
	// if we determined the route is an osmo multi-hop and both routes are incentivized,
	// we utilize a separate function that calculates the discounted swap fees
	if isMultiHopRouted {
		insExpected, err = k.createOsmoMultihopExpectedSwapOuts(ctx, routes, tokenOut, routeSwapFee, sumOfSwapFees)
	} else {
		insExpected, err = k.createMultihopExpectedSwapOuts(ctx, routes, tokenOut)
	}
	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}

	return insExpected[0], nil
}

func (k Keeper) isOsmoRoutedMultihop(ctx sdk.Context, route types.MultihopRoute, inDenom, outDenom string) (isRouted bool) {
	if route.Length() != 2 {
		return false
	}
	intemediateDenoms := route.IntermediateDenoms()
	if len(intemediateDenoms) != 1 || intemediateDenoms[0] != appparams.BaseCoinUnit {
		return false
	}
	if inDenom == outDenom {
		return false
	}
	poolIds := route.PoolIds()
	if poolIds[0] == poolIds[1] {
		return false
	}

	route0Incentivized := k.poolIncentivesKeeper.IsPoolIncentivized(ctx, poolIds[0])
	route1Incentivized := k.poolIncentivesKeeper.IsPoolIncentivized(ctx, poolIds[1])

	return route0Incentivized && route1Incentivized
}

func (k Keeper) getOsmoRoutedMultihopTotalSwapFee(ctx sdk.Context, route types.MultihopRoute) (
	totalPathSwapFee sdk.Dec, sumOfSwapFees sdk.Dec, err error,
) {
	additiveSwapFee := sdk.ZeroDec()
	maxSwapFee := sdk.ZeroDec()

	for _, poolId := range route.PoolIds() {
		swapModule, err := k.GetPoolModule(ctx, poolId)
		if err != nil {
			return sdk.Dec{}, sdk.Dec{}, err
		}

		pool, poolErr := swapModule.GetPool(ctx, poolId)
		if poolErr != nil {
			return sdk.Dec{}, sdk.Dec{}, poolErr
		}
		swapFee := pool.GetSwapFee(ctx)
		additiveSwapFee = additiveSwapFee.Add(swapFee)
		maxSwapFee = sdk.MaxDec(maxSwapFee, swapFee)
	}
	averageSwapFee := additiveSwapFee.QuoInt64(2)
	maxSwapFee = sdk.MaxDec(maxSwapFee, averageSwapFee)
	return maxSwapFee, additiveSwapFee, nil
}

// createMultihopExpectedSwapOuts defines the output denom and output amount for the last pool in
// the route of pools the caller is intending to hop through in a fixed-output multihop tx. It estimates the input
// amount for this last pool and then chains that input as the output of the previous pool in the route, repeating
// until the first pool is reached. It returns an array of inputs, each of which correspond to a pool ID in the
// route of pools for the original multihop transaction.
// TODO: test this.
func (k Keeper) createMultihopExpectedSwapOuts(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(routes))
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]

		swapModule, err := k.GetPoolModule(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		poolI, err := swapModule.GetPool(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, route.TokenInDenom, poolI.GetSwapFee(ctx))
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
	cumulativeRouteSwapFee, sumOfSwapFees sdk.Dec,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(routes))
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]

		swapModule, err := k.GetPoolModule(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		poolI, err := swapModule.GetPool(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		swapFee := poolI.GetSwapFee(ctx)
		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, route.TokenInDenom, cumulativeRouteSwapFee.Mul((swapFee.Quo(sumOfSwapFees))))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}
