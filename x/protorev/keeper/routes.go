package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

type RouteMetaData struct {
	// The route that was built
	Route poolmanagertypes.SwapAmountInRoutes
	// The number of pool points that were consumed to build the route
	PoolPoints uint64
	// The step size that should be used in the binary search for the optimal swap amount
	StepSize osmomath.Int
}

// BuildRoutes builds all of the possible arbitrage routes given the tokenIn, tokenOut and poolId that were used in the swap.
func (k Keeper) BuildRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) []RouteMetaData {
	routes := make([]RouteMetaData, 0)

	// Append hot routes if they exist
	if tokenPairRoutes, err := k.BuildHotRoutes(ctx, tokenIn, tokenOut, poolId); err == nil {
		routes = append(routes, tokenPairRoutes...)
	}

	// Append highest liquidity routes if they exist
	if highestLiquidityRoutes, err := k.BuildHighestLiquidityRoutes(ctx, tokenIn, tokenOut, poolId); err == nil {
		routes = append(routes, highestLiquidityRoutes...)
	}

	return routes
}

// BuildHotRoutes builds all of the possible arbitrage routes using the hot routes method.
func (k Keeper) BuildHotRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]RouteMetaData, error) {
	routes := make([]RouteMetaData, 0)
	// Get all of the routes from the store that match the given tokenIn and tokenOut
	tokenPairArbRoutes, err := k.GetTokenPairArbRoutes(ctx, tokenIn, tokenOut)
	if err != nil {
		return routes, err
	}

	// Iterate through all of the routes and build hot routes
	for _, route := range tokenPairArbRoutes.ArbRoutes {
		if newRoute, err := k.BuildHotRoute(ctx, route, poolId); err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildHotRoute constructs a cyclic arbitrage route given a hot route and swap that should be placed in the hot route.
func (k Keeper) BuildHotRoute(ctx sdk.Context, route types.Route, poolId uint64) (RouteMetaData, error) {
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

	// Check that the route is valid and update the number of pool points that this route will consume when simulating and executing trades
	routePoolPoints, err := k.CalculateRoutePoolPoints(ctx, newRoute)
	if err != nil {
		return RouteMetaData{}, err
	}

	return RouteMetaData{
		Route:      newRoute,
		PoolPoints: routePoolPoints,
		StepSize:   route.StepSize,
	}, nil
}

// BuildHighestLiquidityRoutes builds cyclic arbitrage routes using the highest liquidity method. The base denoms are sorted by priority
// and routes are built in a greedy manner.
func (k Keeper) BuildHighestLiquidityRoutes(ctx sdk.Context, tokenIn, tokenOut string, poolId uint64) ([]RouteMetaData, error) {
	routes := make([]RouteMetaData, 0)
	baseDenoms, err := k.GetAllBaseDenoms(ctx)
	if err != nil {
		return routes, err
	}

	// Iterate through all denoms greedily. When simulating and executing trades, routes that are closer to the beginning of the list
	// have priority over those that are later in the list. This way we can build routes that are more likely to succeed and bring in
	// higher profits.
	for _, baseDenom := range baseDenoms {
		if newRoute, err := k.BuildHighestLiquidityRoute(ctx, baseDenom, tokenIn, tokenOut, poolId); err == nil {
			routes = append(routes, newRoute)
		}

		if newRoute, err := k.BuildTwoPoolRoute(ctx, baseDenom, tokenIn, tokenOut, poolId); err == nil {
			routes = append(routes, newRoute)
		}
	}

	return routes, nil
}

// BuildHighestLiquidityRoute constructs a cyclic arbitrage route that is starts/ends with swapDenom (ex. osmo) given the swap (tokenIn, tokenOut, poolId).
func (k Keeper) BuildHighestLiquidityRoute(ctx sdk.Context, swapDenom types.BaseDenom, tokenIn, tokenOut string, poolId uint64) (RouteMetaData, error) {
	// Create the first swap for the MultiHopSwap Route
	entryPoolId, err := k.GetPoolForDenomPair(ctx, swapDenom.Denom, tokenOut)
	if err != nil {
		return RouteMetaData{}, err
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
	exitPoolId, err := k.GetPoolForDenomPair(ctx, swapDenom.Denom, tokenIn)
	if err != nil {
		return RouteMetaData{}, err
	}
	exitHop := poolmanagertypes.SwapAmountInRoute{
		PoolId:        exitPoolId,
		TokenOutDenom: swapDenom.Denom,
	}

	newRoute := poolmanagertypes.SwapAmountInRoutes{entryHop, middleHop, exitHop}

	// Check that the route is valid and update the number of pool points that this route will consume when simulating and executing trades
	routePoolPoints, err := k.CalculateRoutePoolPoints(ctx, newRoute)
	if err != nil {
		return RouteMetaData{}, err
	}

	return RouteMetaData{
		Route:      newRoute,
		PoolPoints: routePoolPoints,
		StepSize:   swapDenom.StepSize,
	}, nil
}

// BuildTwoPoolRoute will attempt to create a two pool route that will rebalance pools that are paired
// with the base denom. This is useful for pools that contain the same assets but are imbalanced.
func (k Keeper) BuildTwoPoolRoute(
	ctx sdk.Context,
	baseDenom types.BaseDenom,
	tokenInDenom, tokenOutDenom string,
	poolId uint64,
) (RouteMetaData, error) {
	if baseDenom.Denom != tokenInDenom && baseDenom.Denom != tokenOutDenom {
		return RouteMetaData{}, fmt.Errorf("base denom (%s) must be either tokenIn (%s) or tokenOut (%s)", baseDenom.Denom, tokenInDenom, tokenOutDenom)
	}

	var (
		pool1, pool2 uint64
	)

	// In the case where the base denom is the swap out, the base denom becomes more ~expensive~ on the current pool id
	// and potentially cheaper on the highest liquidity pool. So we swap first on the current pool id and then on the
	// highest liquidity pool.
	if tokenOutDenom == baseDenom.Denom {
		highestLiquidityPool, err := k.GetPoolForDenomPair(ctx, baseDenom.Denom, tokenInDenom)
		if err != nil {
			return RouteMetaData{}, err
		}

		pool1, pool2 = poolId, highestLiquidityPool
		tokenOutDenom = tokenInDenom
	} else {
		highestLiquidityPool, err := k.GetPoolForDenomPair(ctx, baseDenom.Denom, tokenOutDenom)
		if err != nil {
			return RouteMetaData{}, err
		}

		pool1, pool2 = highestLiquidityPool, poolId
	}

	if pool1 == pool2 {
		return RouteMetaData{}, types.ErrRouteDoubleContainsPool
	}

	newRoute := poolmanagertypes.SwapAmountInRoutes{
		poolmanagertypes.SwapAmountInRoute{
			TokenOutDenom: tokenOutDenom,
			PoolId:        pool1,
		},
		poolmanagertypes.SwapAmountInRoute{
			TokenOutDenom: baseDenom.Denom,
			PoolId:        pool2,
		},
	}

	// Check that the route is valid and update the number of pool points that this route will consume when simulating and executing trades
	routePoolPoints, err := k.CalculateRoutePoolPoints(ctx, newRoute)
	if err != nil {
		return RouteMetaData{}, err
	}

	return RouteMetaData{
		Route:      newRoute,
		PoolPoints: routePoolPoints,
		StepSize:   baseDenom.StepSize,
	}, nil
}

// CalculateRoutePoolPoints calculates the number of pool points that will be consumed by a route when simulating and executing trades. This
// is only added to the global pool point counter if the route simulated is minimally profitable i.e. it will make a profit.
func (k Keeper) CalculateRoutePoolPoints(ctx sdk.Context, route poolmanagertypes.SwapAmountInRoutes) (uint64, error) {
	// Calculate the number of pool points this route will consume
	infoByPoolType := k.GetInfoByPoolType(ctx)
	totalWeight := uint64(0)

	for _, poolId := range route.PoolIds() {
		poolType, err := k.poolmanagerKeeper.GetPoolType(ctx, poolId)
		if err != nil {
			return 0, err
		}

		switch poolType {
		case poolmanagertypes.Balancer:
			totalWeight += infoByPoolType.Balancer.Weight
		case poolmanagertypes.Stableswap:
			totalWeight += infoByPoolType.Stable.Weight
		case poolmanagertypes.Concentrated:
			totalWeight += infoByPoolType.Concentrated.Weight
		case poolmanagertypes.CosmWasm:
			weight, ok := uint64(0), false
			pool, err := k.poolmanagerKeeper.GetPool(ctx, poolId)
			if err != nil {
				return 0, err
			}
			poolAddrString := pool.GetAddress().String()
			for _, weightMap := range infoByPoolType.Cosmwasm.WeightMaps {
				if weightMap.ContractAddress == poolAddrString {
					weight = weightMap.Weight
					ok = true
					break
				}
			}
			if !ok {
				return 0, fmt.Errorf("cosmwasm pool %d does not have a weight", poolId)
			}

			totalWeight += weight
		default:
			return 0, errors.New("invalid pool type")
		}
	}

	remainingPoolPoints, _, err := k.GetRemainingPoolPoints(ctx)
	if err != nil {
		return 0, err
	}

	// If the route consumes more pool points than are available, return an error
	if totalWeight > remainingPoolPoints {
		return 0, fmt.Errorf("route consumes %d pool points but only %d are available", totalWeight, remainingPoolPoints)
	}

	return totalWeight, nil
}

// IsValidPool checks if the pool is active and exists
func (k Keeper) IsValidPool(ctx sdk.Context, pool poolmanagertypes.PoolI) error {
	if !pool.IsActive(ctx) {
		return fmt.Errorf("pool %d is not active", pool.GetId())
	}

	return nil
}
