package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

// BuildRoutes builds all of the possible arbitrage routes given the tokenIn, tokenOut and poolId that were used in the swap.
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, remainingPoolPoints *uint64) []poolmanagertypes.SwapAmountInRoutes {
	routes := make([]poolmanagertypes.SwapAmountInRoutes, 0)

	// Append hot routes if they exist
	if tokenPairRoutes, err := k.BuildHotRoutes(ctx, tokenIn, tokenOut, poolId, remainingPoolPoints); err == nil {
		routes = append(routes, tokenPairRoutes...)
	}

	// Append highest liquidity routes if they exist
	if highestLiquidityRoutes, err := k.BuildHighestLiquidityRoutes(ctx, tokenIn, tokenOut, poolId, remainingPoolPoints); err == nil {
		routes = append(routes, highestLiquidityRoutes...)
	}

	return routes
}

// BuildHotRoutes builds all of the possible arbitrage routes using the hot routes method.
func (k Keeper) BuildHotRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, remainingPoolPoints *uint64) ([]poolmanagertypes.SwapAmountInRoutes, error) {
	if *remainingPoolPoints <= 0 {
		return []poolmanagertypes.SwapAmountInRoutes{}, fmt.Errorf("the number of pool points that can be consumed has been reached")
	}

	// Get all of the routes from the store that match the given tokenIn and tokenOut
	tokenPairArbRoutes, err := k.GetTokenPairArbRoutes(ctx, tokenIn, tokenOut)
	if err != nil {
		return []poolmanagertypes.SwapAmountInRoutes{}, err
	}

	// Iterate through all of the routes and build hot routes
	routes := make([]poolmanagertypes.SwapAmountInRoutes, 0)
	for index := 0; index < len(tokenPairArbRoutes.ArbRoutes) && *remainingPoolPoints > 0; index++ {
		if newRoute, err := k.BuildHotRoute(ctx, tokenPairArbRoutes.ArbRoutes[index], poolId, remainingPoolPoints); err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildHotRoute constructs a cyclic arbitrage route given a hot route and swap that should be placed in the hot route.
func (k Keeper) BuildHotRoute(ctx sdk.Context, route *types.Route, poolId uint64, remainingPoolPoints *uint64) (poolmanagertypes.SwapAmountInRoutes, error) {
	newRoute := make(poolmanagertypes.SwapAmountInRoutes, 0)

	for _, trade := range route.Trades {
		// 0 is a placeholder for pools swapped on that should be entered into the hot route
		if trade.Pool == 0 {
			newRoute = append(newRoute, poolmanagertypes.SwapAmountInRoute{
				PoolId:        poolId,
				TokenOutDenom: trade.TokenOut,
			})
		} else {
			newRoute = append(newRoute, poolmanagertypes.SwapAmountInRoute{
				PoolId:        trade.Pool,
				TokenOutDenom: trade.TokenOut,
			})
		}
	}

	// Check that the route is valid and update the number of pool points that can be consumed after the route is built
	if err := k.CheckAndUpdateRouteState(ctx, newRoute, remainingPoolPoints); err != nil {
		return poolmanagertypes.SwapAmountInRoutes{}, err
	}

	return newRoute, nil
}

// BuildHighestLiquidityRoutes builds cyclic arbitrage routes using the highest liquidity method. The base denoms are sorted by priority
// and routes are built in a greedy manner.
func (k Keeper) BuildHighestLiquidityRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64, remainingPoolPoints *uint64) ([]poolmanagertypes.SwapAmountInRoutes, error) {
	routes := make([]poolmanagertypes.SwapAmountInRoutes, 0)
	baseDenoms := k.GetAllBaseDenoms(ctx)

	// Iterate through all denoms greedily and build routes until the max number of pool points to be consumed is reached
	for index := 0; index < len(baseDenoms) && *remainingPoolPoints > 0; index++ {
		if newRoute, err := k.BuildHighestLiquidityRoute(ctx, baseDenoms[index], tokenIn, tokenOut, poolId, remainingPoolPoints); err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildHighestLiquidityRoute constructs a cyclic arbitrage route that is starts/ends with swapDenom (ex. osmo) given the swap (tokenIn, tokenOut, poolId).
func (k Keeper) BuildHighestLiquidityRoute(ctx sdk.Context, swapDenom, tokenIn, tokenOut string, poolId uint64, remainingPoolPoints *uint64) (poolmanagertypes.SwapAmountInRoutes, error) {
	// Create the first swap for the MultiHopSwap Route
	entryPoolId, err := k.GetPoolForDenomPair(ctx, swapDenom, tokenOut)
	if err != nil {
		return poolmanagertypes.SwapAmountInRoutes{}, err
	}
	entryHop := poolmanagertypes.SwapAmountInRoute{
		PoolId:        entryPoolId,
		TokenOutDenom: tokenOut,
	}

	middleHop := poolmanagertypes.SwapAmountInRoute{
		PoolId:        poolId,
		TokenOutDenom: tokenIn,
	}

	// Creating the third swap in the arb
	exitPoolId, err := k.GetPoolForDenomPair(ctx, swapDenom, tokenIn)
	if err != nil {
		return poolmanagertypes.SwapAmountInRoutes{}, err
	}
	exitHop := poolmanagertypes.SwapAmountInRoute{
		PoolId:        exitPoolId,
		TokenOutDenom: swapDenom,
	}

	newRoute := poolmanagertypes.SwapAmountInRoutes{entryHop, middleHop, exitHop}

	// Check that the route is valid and update the number of pool points that can be consumed after the route is built
	if err := k.CheckAndUpdateRouteState(ctx, newRoute, remainingPoolPoints); err != nil {
		return poolmanagertypes.SwapAmountInRoutes{}, err
	}

	return newRoute, nil
}

// CheckAndUpdateRouteState checks if the cyclic arbitrage route that was created via the highest liquidity route or hot route method is valid.
// If the route is too expensive to iterate through, has a inactive or invalid pool, or unsupported pool type, an error is returned.
func (k Keeper) CheckAndUpdateRouteState(ctx sdk.Context, route poolmanagertypes.SwapAmountInRoutes, remainingPoolPoints *uint64) error {
	if *remainingPoolPoints <= 0 {
		return fmt.Errorf("the number of routes that can be iterated through has been reached")
	}

	poolWeights := k.GetPoolWeights(ctx)

	totalWeight := uint64(0)
	poolIds := route.PoolIds()
	for index := 0; totalWeight <= *remainingPoolPoints && index < len(poolIds); index++ {
		// Ensure that all of the pools in the route exist and are active
		if err := k.IsValidPool(ctx, poolIds[index]); err != nil {
			return err
		}

		poolType, err := k.gammKeeper.GetPoolType(ctx, poolIds[index])
		if err != nil {
			return err
		}

		switch poolType {
		case poolmanagertypes.Balancer:
			totalWeight += poolWeights.BalancerWeight
		case poolmanagertypes.Stableswap:
			totalWeight += poolWeights.StableWeight
		case poolmanagertypes.Concentrated:
			totalWeight += poolWeights.ConcentratedWeight
		default:
			return fmt.Errorf("invalid pool type")
		}
	}

	// Check that the route can be iterated
	if *remainingPoolPoints < totalWeight {
		return fmt.Errorf("the total weight of the route is too expensive to iterate through: %d > %d", totalWeight, *remainingPoolPoints)
	}

	if err := k.IncrementPointCountForBlock(ctx, totalWeight); err != nil {
		return err
	}
	*remainingPoolPoints -= totalWeight
	return nil
}

// IsValidPool checks if the pool is active and exists
func (k Keeper) IsValidPool(ctx sdk.Context, poolId uint64) error {
	pool, err := k.gammKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}
	if !pool.IsActive(ctx) {
		return fmt.Errorf("pool %d is not active", poolId)
	}
	return nil
}

// RemainingPoolPointsForTx calculates the number of pool points that can be consumed in the current transaction.
// Returns a pointer that will be used throughout the lifetime of a transaction.
func (k Keeper) RemainingPoolPointsForTx(ctx sdk.Context) (*uint64, error) {
	maxRoutesPerTx, err := k.GetMaxPointsPerTx(ctx)
	if err != nil {
		return nil, err
	}

	maxRoutesPerBlock, err := k.GetMaxPointsPerBlock(ctx)
	if err != nil {
		return nil, err
	}

	currentRouteCount, err := k.GetPointCountForBlock(ctx)
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
