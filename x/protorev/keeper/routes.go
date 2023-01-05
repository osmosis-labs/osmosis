package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// BuildRoutes builds all of the possible arbitrage routes given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, maxIterableRoutes *uint64) []swaproutertypes.SwapAmountInRoutes {
	routes := make([]swaproutertypes.SwapAmountInRoutes, 0)

	// Append hot routes if they exist
	if tokenPairRoutes, err := k.BuildTokenPairRoutes(ctx, tokenIn, tokenOut, poolId, maxIterableRoutes); err == nil {
		routes = append(routes, tokenPairRoutes...)
	}

	// Append an osmo route if one exists
	if osmoRoute, err := k.BuildOsmoRoute(ctx, tokenIn, tokenOut, poolId, maxIterableRoutes); err == nil {
		routes = append(routes, osmoRoute)
	}

	// Append an atom route if one exists
	if atomRoute, err := k.BuildAtomRoute(ctx, tokenIn, tokenOut, poolId, maxIterableRoutes); err == nil {
		routes = append(routes, atomRoute)
	}

	return routes
}

// BuildTokenPairRoutes builds all of the possible arbitrage routes from the hot routes given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildTokenPairRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, maxIterableRoutes *uint64) ([]swaproutertypes.SwapAmountInRoutes, error) {
	if *maxIterableRoutes <= 0 {
		return []swaproutertypes.SwapAmountInRoutes{}, fmt.Errorf("the number of routes that can be iterated through has been exceeded")
	}

	// Get all of the routes from the store that match the given tokenIn and tokenOut
	tokenPairArbRoutes, err := k.GetTokenPairArbRoutes(ctx, tokenIn, tokenOut)
	if err != nil {
		return []swaproutertypes.SwapAmountInRoutes{}, err
	}

	// Iterate through all of the routes and build hot routes
	routes := make([]swaproutertypes.SwapAmountInRoutes, 0)
	for index := 0; index < len(tokenPairArbRoutes.ArbRoutes) && *maxIterableRoutes > 0; index++ {
		if newRoute, err := k.BuildHotRoute(ctx, tokenPairArbRoutes.ArbRoutes[index], tokenIn, tokenOut, poolId, maxIterableRoutes); err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildHotRoute constructs a cyclic arbitrage route given a hot route from the store and information about the swap that should be placed
// in the hot route.
func (k Keeper) BuildHotRoute(ctx sdk.Context, route *types.Route, tokenIn, tokenOut string, poolId uint64, maxIterableRoutes *uint64) (swaproutertypes.SwapAmountInRoutes, error) {
	newRoute := make(swaproutertypes.SwapAmountInRoutes, 0)

	for _, trade := range route.Trades {
		// 0 is a placeholder for pools swapped on that should be entered into the hot route
		if trade.Pool == 0 {
			newRoute = append(newRoute, swaproutertypes.SwapAmountInRoute{
				PoolId:        poolId,
				TokenOutDenom: trade.TokenOut,
			})
		} else {
			newRoute = append(newRoute, swaproutertypes.SwapAmountInRoute{
				PoolId:        trade.Pool,
				TokenOutDenom: trade.TokenOut,
			})
		}
	}

	// Check that the hot route is valid
	if err := k.CheckValidHotRoute(ctx, newRoute); err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}

	// Check that the route can be iterated
	if weight, err := k.GetRouteWeight(ctx, newRoute); err == nil && *maxIterableRoutes >= weight {
		err := k.IncrementRouteCountForBlock(ctx, weight)
		if err != nil {
			return swaproutertypes.SwapAmountInRoutes{}, err
		}

		*maxIterableRoutes -= weight
		return newRoute, nil
	}

	return swaproutertypes.SwapAmountInRoutes{}, fmt.Errorf("the number of routes that can be iterated through has been exceeded")
}

// CheckValidHotRoute checks if the cyclic arbitrage route that was built using the hot routes method is correct. Much of the stateless
// validation achieves the desired checks, however, we also check that the route is traversing pools that
// are active.
func (k Keeper) CheckValidHotRoute(ctx sdk.Context, route swaproutertypes.SwapAmountInRoutes) error {
	if route.Length() != 3 {
		return fmt.Errorf("invalid hot route length")
	}

	// Ensure that all of the pools in the route exist and are active
	for _, poolId := range route.PoolIds() {
		_, err := k.GetAndCheckPool(ctx, poolId)
		if err != nil {
			return err
		}
	}

	return nil
}

// BuildOsmoRoute builds a cyclic arbitrage route that starts and ends with osmo given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildOsmoRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, maxIterableRoutes *uint64) (swaproutertypes.SwapAmountInRoutes, error) {
	return k.BuildRoute(ctx, types.OsmosisDenomination, tokenIn, tokenOut, poolId, maxIterableRoutes, k.GetOsmoPool)
}

// BuildAtomRoute builds a cyclic arbitrage route that starts and ends with atom given the tokenIn, tokenOut and poolId that were used in the swap
func (k Keeper) BuildAtomRoute(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, maxIterableRoutes *uint64) (swaproutertypes.SwapAmountInRoutes, error) {
	return k.BuildRoute(ctx, types.AtomDenomination, tokenIn, tokenOut, poolId, maxIterableRoutes, k.GetAtomPool)
}

// BuildRoute constructs a cyclic arbitrage route that is starts/ends with swapDenom (atom or osmo) given the swap (tokenIn, tokenOut, poolId), and
// a function that can get the poolId from the store given a (token, swapDenom) pair.
func (k Keeper) BuildRoute(ctx sdk.Context, swapDenom, tokenIn, tokenOut string, poolId uint64, maxIterableRoutes *uint64, getPoolIDFromStore func(sdk.Context, string) (uint64, error)) (swaproutertypes.SwapAmountInRoutes, error) {
	if *maxIterableRoutes <= 0 {
		return swaproutertypes.SwapAmountInRoutes{}, fmt.Errorf("the number of routes that can be iterated through has been exceeded")
	}

	// Creating the first trade in the arb
	entryPoolId, err := getPoolIDFromStore(ctx, tokenOut)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}

	// Check that the pool exists and is active
	_, err = k.GetAndCheckPool(ctx, entryPoolId)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	// Create the first swap for the MultiHopSwap Route
	entryRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        entryPoolId,
		TokenOutDenom: tokenOut,
	}

	// Creating the second trade in the arb
	_, err = k.GetAndCheckPool(ctx, poolId)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	middleRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        poolId,
		TokenOutDenom: tokenIn,
	}

	// Creating the third trade in the arb
	exitPoolId, err := getPoolIDFromStore(ctx, tokenIn)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	_, err = k.GetAndCheckPool(ctx, exitPoolId)
	if err != nil {
		return swaproutertypes.SwapAmountInRoutes{}, err
	}
	exitRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        exitPoolId,
		TokenOutDenom: swapDenom,
	}

	newRoute := swaproutertypes.SwapAmountInRoutes{entryRoute, middleRoute, exitRoute}

	// Check that the route can be iterated
	if weight, err := k.GetRouteWeight(ctx, newRoute); err == nil && *maxIterableRoutes >= weight {
		err := k.IncrementRouteCountForBlock(ctx, weight)
		if err != nil {
			return swaproutertypes.SwapAmountInRoutes{}, err
		}

		*maxIterableRoutes -= weight
		return newRoute, nil
	}

	return swaproutertypes.SwapAmountInRoutes{}, fmt.Errorf("the number of routes that can be iterated through has been exceeded")
}

// GetAndCheckPool retrieves the pool from the x/gamm module given a poolId and ensures that the pool can be traded on
func (k Keeper) GetAndCheckPool(ctx sdk.Context, poolId uint64) (gammtypes.CFMMPoolI, error) {
	pool, err := k.gammKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return pool, err
	}
	if !pool.IsActive(ctx) {
		return pool, fmt.Errorf("pool %d is not active", poolId)
	}
	return pool, nil
}

// GetRouteWeight retrieves the weight of a route. The weight of a route is determined by the pools that are used in the route.
// Different pools will have different execution times hence the need for a weighted point system.
func (k Keeper) GetRouteWeight(ctx sdk.Context, route swaproutertypes.SwapAmountInRoutes) (uint64, error) {
	// Routes must always be of length 3
	if route.Length() != 3 {
		return 0, fmt.Errorf("invalid route length")
	}

	// The middle pool is the pool that may be a stable pool (outside pools will always be balancer pools)
	middlePool := route.PoolIds()[1]
	poolType, err := k.gammKeeper.GetPoolType(ctx, middlePool)
	if err != nil {
		return 0, err
	}

	// Get the weights of the route types
	routeWeights, err := k.GetRouteWeights(ctx)
	if err != nil {
		return 0, err
	}

	switch poolType {
	case swaproutertypes.Balancer:
		return routeWeights.BalancerWeight, nil
	case swaproutertypes.Stableswap:
		return routeWeights.StableWeight, nil
	default:
		return 0, fmt.Errorf("invalid pool type")
	}
}

// CalcNumberOfIterableRoutes calculates the number of routes that can be iterated over in the current transaction.
// Returns a pointer that will be used throughout the lifetime of a transaction.
func (k Keeper) CalcNumberOfIterableRoutes(ctx sdk.Context) (*uint64, error) {
	maxRoutesPerTx, err := k.GetMaxRoutesPerTx(ctx)
	if err != nil {
		return nil, err
	}

	maxRoutesPerBlock, err := k.GetMaxRoutesPerBlock(ctx)
	if err != nil {
		return nil, err
	}

	currentRouteCount, err := k.GetRouteCountForBlock(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate the number of routes that can be iterated over
	numberOfIterableRoutes := maxRoutesPerBlock - currentRouteCount
	if numberOfIterableRoutes > maxRoutesPerTx {
		numberOfIterableRoutes = maxRoutesPerTx
	}

	return &numberOfIterableRoutes, nil
}
