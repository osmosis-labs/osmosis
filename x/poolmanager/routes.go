package poolmanager

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v23/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v23/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

var (
	// We cache direct routes and spot prices to avoid recalculating them.
	// It is important to note, these cache values are only used within the same query.
	// If a new query is made, the cache will be reset.
	directRouteCache map[string]uint64
	spotPriceCache   map[string]osmomath.BigDec
	shouldCache      = false
)

func init() {
	directRouteCache = make(map[string]uint64)
	spotPriceCache = make(map[string]osmomath.BigDec)
}

// findRoutes finds all routes between two tokens that match the given hop count
func findRoutes(g types.RoutingGraphMap, start, end string, hops int) [][]*types.Route {
	if hops < 1 {
		return nil
	}

	var routeRoutes [][]*types.Route

	startRoutes, startExists := g.Graph[start]
	if !startExists {
		return routeRoutes
	}

	for token, routes := range startRoutes.InnerMap {
		if hops == 1 {
			if token == end {
				for _, route := range routes.Routes {
					route.Token = end
					routeRoutes = append(routeRoutes, []*types.Route{route})
				}
			}
		} else {
			subRoutes := findRoutes(g, token, end, hops-1)
			for _, subRoute := range subRoutes {
				for _, route := range routes.Routes {
					route.Token = token
					fullRoute := append([]*types.Route{route}, subRoute...)
					routeRoutes = append(routeRoutes, fullRoute)
				}
			}
		}
	}

	return routeRoutes
}

// SetDenomPairRoutes sets the route map to be used for route calculations.
func (k Keeper) SetDenomPairRoutes(ctx sdk.Context) (types.RoutingGraph, error) {
	// Reset cache at the end of this function
	// We only ever want to use cache in the helper functions if it is being called via this function,
	// since this function is called at upgrade and epoch at determinstic times. If we used cache outside of this function,
	// it's possible the cache will mess with the determinism of the route map.
	shouldCache = true
	defer func() {
		directRouteCache = make(map[string]uint64)
		spotPriceCache = make(map[string]osmomath.BigDec)
		shouldCache = false
	}()

	// We generate all denom pair routes here, and use them to filter out pools that do not meet the minimum liquidity threshold.
	// While generating routes twice is not ideal, we are optimizing to reduce writes to state.
	pools, routeMap, err := k.generateAllDenomPairRoutes(ctx)
	if err != nil {
		return types.RoutingGraph{}, err
	}

	// Retrieve minimum liquidity threshold from params
	minValueForRoute := k.GetParams(ctx).MinValueForRoute

	// Create a routingGraph to represent possible routes between tokens
	var routingGraph types.RoutingGraph

PoolLoop:
	// Iterate through the pools
	for _, pool := range pools {
		// Some of the first cw pools created have a malformed response and are no longer in use. Remove these pools to prevent issues.
		tokens := pool.GetPoolDenoms(ctx)
		for _, token := range tokens {
			if strings.Contains(token, "pool_asset_denoms") {
				continue PoolLoop
			}
		}

		poolLiquidityInTargetDenom, err := k.poolLiquidityFromOSMOToTargetDenom(ctx, pool, routeMap, minValueForRoute.Denom)
		if err != nil {
			return types.RoutingGraph{}, err
		}
		if poolLiquidityInTargetDenom.LT(minValueForRoute.Amount) {
			continue
		}

		poolID := pool.GetId()
		// Create edges for all possible combinations of tokens
		for i := 0; i < len(tokens); i++ {
			for j := i + 1; j < len(tokens); j++ {
				// Add edges with the associated token
				routingGraph.AddEdge(tokens[i], tokens[j], tokens[i], poolID)
				routingGraph.AddEdge(tokens[j], tokens[i], tokens[j], poolID)
			}
		}
	}

	// Set the route map in state
	// NOTE: This is done with the non map version of the route graph
	// If we used maps here, the serialization would be non-deterministic
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyRouteMap, &routingGraph)
	return routingGraph, nil
}

// generateAllDenomPairRoutes generates all possible routes between tokens, without filtering out pools based on liquidity.
// This returns the pools and route map for use in other functions and does not set the route graph in state.
func (k Keeper) generateAllDenomPairRoutes(ctx sdk.Context) ([]types.PoolI, types.RoutingGraphMap, error) {
	// Get all the pools
	pools, err := k.AllPools(ctx)
	if err != nil {
		return []types.PoolI{}, types.RoutingGraphMap{}, err
	}

	// Create a routingGraph to represent possible routes between tokens
	var routingGraph types.RoutingGraph

PoolLoop:
	// Iterate through the pools
	for _, pool := range pools {
		// Some of the first cw pools created have a malformed response and are no longer in use. Remove these pools to prevent issues.
		tokens := pool.GetPoolDenoms(ctx)
		for _, token := range tokens {
			if strings.Contains(token, "pool_asset_denoms") {
				continue PoolLoop
			}
		}

		poolID := pool.GetId()
		// Create edges for all possible combinations of tokens
		for i := 0; i < len(tokens); i++ {
			for j := i + 1; j < len(tokens); j++ {
				// Add edges with the associated token
				routingGraph.AddEdge(tokens[i], tokens[j], tokens[i], poolID)
				routingGraph.AddEdge(tokens[j], tokens[i], tokens[j], poolID)
			}
		}
	}

	// Convert the route graph to a map for easier access
	routeMap := convertToMap(&routingGraph)

	return pools, routeMap, nil
}

// GetRouteMap returns the route map that is stored in state.
// It converts the route graph stored in the KVStore to a map for easier access.
func (k Keeper) GetRouteMap(ctx sdk.Context) (types.RoutingGraphMap, error) {
	var routeGraph types.RoutingGraph

	found, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyRouteMap, &routeGraph)
	if err != nil {
		return types.RoutingGraphMap{}, err
	}
	if !found {
		return types.RoutingGraphMap{}, fmt.Errorf("route map not found")
	}

	routeMap := convertToMap(&routeGraph)

	return routeMap, nil
}

// getDirectRouteWithMostLiquidity returns the single hop route with the highest liquidity between an input denom and output denom.
func (k Keeper) getDirectRouteWithMostLiquidity(ctx sdk.Context, inputDenom, outputDenom string, routeMap types.RoutingGraphMap) (uint64, error) {
	// Get all direct routes from the input denom to uosmo
	directRoutes := findRoutes(routeMap, inputDenom, outputDenom, 1)

	// Store liquidity for all direct routes found
	routeLiquidity := make(map[string]osmomath.Int)
	for _, route := range directRoutes {
		liquidity, err := k.getPoolLiquidityOfDenom(ctx, route[0].PoolId, outputDenom)
		if err != nil {
			return 0, err
		}
		routeKey := fmt.Sprintf("%v", route[0].PoolId)
		routeLiquidity[routeKey] = liquidity
	}

	// Extract and sort the keys from the routeLiquidity map
	// This ensures deterministic selection of the best route
	var keys []string
	for k := range routeLiquidity {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Find the single hop route with the highest liquidity
	var bestRouteKey string
	maxLiquidity := osmomath.ZeroInt()
	for _, routeKey := range keys {
		liquidity := routeLiquidity[routeKey]
		// Update best route if a higher liquidity is found,
		// or if the liquidity is equal but the routeKey is encountered earlier in the sorted order
		if liquidity.GT(maxLiquidity) || (liquidity.Equal(maxLiquidity) && bestRouteKey == "") {
			bestRouteKey = routeKey
			maxLiquidity = liquidity
		}
	}
	if bestRouteKey == "" {
		return 0, nil
	}

	// Convert the best route key back to []uint64
	var bestRoute []uint64
	cleanedRouteKey := strings.Trim(bestRouteKey, "[]")
	idStrs := strings.Split(cleanedRouteKey, " ")

	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing pool ID: %v", err)
		}
		bestRoute = append(bestRoute, id)
	}

	// Return the route with the highest liquidity
	return bestRoute[0], nil
}

// inputAmountToOSMO transforms an input denom and its amount to uosmo
// If a route is not found, returns 0 with no error.
// Note, this method only utilizes cache in the getter and setter of denom pair routes.
// TODO: When implementing getter in follow on PR, ensure we turn cache on for more optimal performance.
func (k Keeper) inputAmountToTargetDenom(ctx sdk.Context, inputDenom, targetDenom string, amount osmomath.Int, routeMap types.RoutingGraphMap) (osmomath.Int, error) {
	if inputDenom == targetDenom {
		return amount, nil
	}

	var route uint64
	var err error

	if shouldCache {
		// Check if the route is cached
		if cachedRoute, ok := directRouteCache[inputDenom]; ok {
			route = cachedRoute
		} else {
			// If not, get the route and cache it
			route, err = k.getDirectRouteWithMostLiquidity(ctx, inputDenom, targetDenom, routeMap)
			if err != nil {
				return osmomath.ZeroInt(), nil
			}
			if route == 0 {
				return osmomath.ZeroInt(), nil
			}
			directRouteCache[inputDenom] = route
		}
	} else {
		route, err = k.getDirectRouteWithMostLiquidity(ctx, inputDenom, targetDenom, routeMap)
		if err != nil {
			return osmomath.ZeroInt(), nil
		}
		if route == 0 {
			return osmomath.ZeroInt(), nil
		}
	}

	var taretDenomPerInputToken osmomath.BigDec

	if shouldCache {
		// Check if the spot price is cached
		spotPriceKey := fmt.Sprintf("%d:%s", route, inputDenom)
		if cachedSpotPrice, ok := spotPriceCache[spotPriceKey]; ok {
			taretDenomPerInputToken = cachedSpotPrice
		} else {
			// If not, calculate the spot price and cache it
			taretDenomPerInputToken, err = k.RouteCalculateSpotPrice(ctx, route, targetDenom, inputDenom)
			if err != nil {
				return osmomath.ZeroInt(), err
			}
			spotPriceCache[spotPriceKey] = taretDenomPerInputToken
		}
	} else {
		taretDenomPerInputToken, err = k.RouteCalculateSpotPrice(ctx, route, targetDenom, inputDenom)
		if err != nil {
			return osmomath.ZeroInt(), err
		}
	}

	// Convert the input denom to target denom
	// Rounding is fine here
	targetDenomAmount := amount.ToLegacyDec().Mul(taretDenomPerInputToken.Dec())
	return targetDenomAmount.RoundInt(), nil
}

// getPoolLiquidityOfDenom returns the liquidity of a denom in a pool.
// This calls different methods depending on the pool type.
func (k Keeper) getPoolLiquidityOfDenom(ctx sdk.Context, poolId uint64, denom string) (osmomath.Int, error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	// Check the pool type, and check the pool liquidity based on the type
	switch pool.GetType() {
	case types.Balancer:
		pool, ok := pool.(gammtypes.CFMMPoolI)
		if !ok {
			return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
		}
		totalPoolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		return totalPoolLiquidity.AmountOf(denom), nil
	case types.Stableswap:
		pool, ok := pool.(gammtypes.CFMMPoolI)
		if !ok {
			return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
		}
		totalPoolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		return totalPoolLiquidity.AmountOf(denom), nil
	case types.Concentrated:
		poolAddress := pool.GetAddress()
		poolAddressBalances := k.bankKeeper.GetAllBalances(ctx, poolAddress)
		return poolAddressBalances.AmountOf(denom), nil
	case types.CosmWasm:
		pool, ok := pool.(cosmwasmpooltypes.CosmWasmExtension)
		if !ok {
			return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
		}
		totalPoolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		return totalPoolLiquidity.AmountOf(denom), nil
	default:
		return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
	}
}

// poolLiquidityToTargetDenom returns the total liquidity of a pool in terms of the target denom.
func (k Keeper) poolLiquidityToTargetDenom(ctx sdk.Context, pool types.PoolI, routeMap types.RoutingGraphMap, targetDenom string) (osmomath.Int, error) {
	poolDenoms := pool.GetPoolDenoms(ctx)
	totalLiquidity := sdk.ZeroInt()
	for _, denom := range poolDenoms {
		liquidity, err := k.getPoolLiquidityOfDenom(ctx, pool.GetId(), denom)
		if err != nil {
			return osmomath.ZeroInt(), err
		}
		targetDenomAmount, err := k.inputAmountToTargetDenom(ctx, denom, targetDenom, liquidity, routeMap)
		if err != nil {
			// no direct route found, so skip this denom
			continue
		}
		totalLiquidity = totalLiquidity.Add(targetDenomAmount)
	}
	return totalLiquidity, nil
}

// convertToMap converts a RoutingGraph to a RoutingGraphMap
// This is done to take advantage of the map data structure for easier access.
func convertToMap(routingGraph *types.RoutingGraph) types.RoutingGraphMap {
	result := types.RoutingGraphMap{
		Graph: make(map[string]*types.InnerMap),
	}

	for _, graphEntry := range routingGraph.Entries {
		innerMap := &types.InnerMap{
			InnerMap: make(map[string]*types.Routes),
		}
		for _, innerMapEntry := range graphEntry.Value.Entries {
			routes := make([]*types.Route, len(innerMapEntry.Value.Routes))
			for i, route := range innerMapEntry.Value.Routes {
				routes[i] = &types.Route{PoolId: route.PoolId, Token: route.Token}
			}
			innerMap.InnerMap[innerMapEntry.Key] = &types.Routes{Routes: routes}
		}
		result.Graph[graphEntry.Key] = innerMap
	}

	return result
}

// poolLiquidityFromOSMOToTargetDenom starts by calculating the total liquidity of a pool in terms of uosmo.
// It then calculates the spot price of uosmo to the target denom as defined via the params.
// This two step process is done because most pools are denominated in uosmo, so it is easier to find a direct route if the liquidity is in uosmo.
func (k Keeper) poolLiquidityFromOSMOToTargetDenom(ctx sdk.Context, pool types.PoolI, routeMap types.RoutingGraphMap, targetDenom string) (osmomath.Int, error) {
	totalLiquidityInOSMO, err := k.poolLiquidityToTargetDenom(ctx, pool, routeMap, k.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return osmomath.ZeroInt(), err
	}
	if totalLiquidityInOSMO.IsZero() {
		return osmomath.ZeroInt(), nil
	}

	osmoUsdPoolId, err := k.getDirectRouteWithMostLiquidity(ctx, k.stakingKeeper.BondDenom(ctx), targetDenom, routeMap)
	if err != nil {
		return osmomath.ZeroInt(), err
	}
	if osmoUsdPoolId == 0 {
		return osmomath.ZeroInt(), nil
	}

	osmoUsdcPool, err := k.GetPool(ctx, osmoUsdPoolId)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	spotPrice, err := osmoUsdcPool.SpotPrice(ctx, k.stakingKeeper.BondDenom(ctx), targetDenom)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	totalLiquidityInUSD := spotPrice.Mul(osmomath.NewBigDec(totalLiquidityInOSMO.Int64())).Dec().TruncateInt()
	return totalLiquidityInUSD, nil
}
