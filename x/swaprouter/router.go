package swaprouter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

var (
	swapFeeMultiplier = sdk.OneDec()
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
	for i, route := range routes {
		swapModule, err := k.GetSwapModule(ctx, route.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		if types.SwapAmountInRoutes(routes).IsOsmoRoutedMultihop() {
			swapFeeMultiplier = gammtypes.MultihopSwapFeeMultiplierForOsmoPools.Clone()
		}

		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := sdk.NewInt(1)
		if len(routes)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, route.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		swapFee := pool.GetSwapFee(ctx).Mul(swapFeeMultiplier)
		tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, route.TokenOutDenom, _outMinAmount, swapFee)
		if err != nil {
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(route.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, nil
}

// RouteExactAmountOut defines the output denom and output amount for the last pool.
// Calculation starts by providing the tokenOutAmount of the final pool to calculate the required tokenInAmount
// the calculated tokenInAmount is used as defined tokenOutAmount of the previous pool, calculating in reverse order of the swap
// Transaction succeeds if the calculated tokenInAmount of the first pool is less than the defined tokenInMaxAmount defined.
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin) (tokenInAmount sdk.Int, err error) {
	if types.SwapAmountOutRoutes(routes).IsOsmoRoutedMultihop() {
		swapFeeMultiplier = gammtypes.MultihopSwapFeeMultiplierForOsmoPools.Clone()
	}

	// Determine what the estimated input would be for each pool along the multihop route
	insExpected, err := k.createMultihopExpectedSwapOuts(ctx, routes, tokenOut, swapFeeMultiplier)
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
		swapModule, err := k.GetSwapModule(ctx, route.PoolId)
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
		swapFee := pool.GetSwapFee(ctx).Mul(swapFeeMultiplier)
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

// createMultihopExpectedSwapOuts defines the output denom and output amount for the last pool in
// the route of pools the caller is intending to hop through in a fixed-output multihop tx. It estimates the input
// amount for this last pool and then chains that input as the output of the previous pool in the route, repeating
// until the first pool is reached. It returns an array of inputs, each of which correspond to a pool ID in the
// route of pools for the original multihop transaction.
// TODO: test this.
func (k Keeper) createMultihopExpectedSwapOuts(
	ctx sdk.Context,
	routes []types.SwapAmountOutRoute,
	tokenOut sdk.Coin, swapFeeMultiplier sdk.Dec,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(routes))
	for i := len(routes) - 1; i >= 0; i-- {
		route := routes[i]

		swapModule, err := k.GetSwapModule(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		pool, err := swapModule.GetPool(ctx, route.PoolId)
		if err != nil {
			return nil, err
		}

		tokenIn, err := pool.CalcInAmtGivenOut(ctx, sdk.NewCoins(tokenOut), route.TokenInDenom, pool.GetSwapFee(ctx).Mul(swapFeeMultiplier))
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

// GetSwapModule returns the swap module for the given pool ID.
// Returns error if:
// - any database error occurs.
// - fails to find a pool with the given id.
// - the swap module of the type corresponding to the pool id is not registered
// in swaprouter's keeper constructor.
// TODO: unexport after concentrated-liqudity upgrade. Currently, it is exported
// for the upgrade handler logic and tests.
func (k Keeper) GetSwapModule(ctx sdk.Context, poolId uint64) (types.SwapI, error) {
	store := ctx.KVStore(k.storeKey)

	moduleRoute := &types.ModuleRoute{}
	found, err := osmoutils.GetIfFound(store, types.FormatModuleRouteKey(poolId), moduleRoute)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.FailedToFindRouteError{PoolId: poolId}
	}

	swapModule, routeExists := k.routes[moduleRoute.PoolType]
	if !routeExists {
		return nil, types.UndefinedRouteError{PoolType: moduleRoute.PoolType, PoolId: poolId}
	}

	return swapModule, nil
}
