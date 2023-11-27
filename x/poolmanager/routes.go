package poolmanager

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var (
	directRouteCache map[string]uint64
	spotPriceCache   map[string]osmomath.BigDec
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

// SetDenomPairRoutes sets the route map to be used for route calculations
func (k Keeper) SetDenomPairRoutes(ctx sdk.Context) (types.RoutingGraph, error) {
	// Reset cache at the end of this function
	defer func() {
		directRouteCache = make(map[string]uint64)
		spotPriceCache = make(map[string]osmomath.BigDec)
	}()

	// Get all the pools
	pools, err := k.AllPools(ctx)
	if err != nil {
		return types.RoutingGraph{}, err
	}

	// Utilize the previous route map if it exists to determine which pools we can leave out of the route map
	// due to insufficient liquidity. If it doesn't exist, we will utilize all pools, and the next time this function
	// is called, we will have a previous route map to utilize.
	var previousRouteGraph types.RoutingGraph
	previousRouteMapFound, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyRouteMap, &previousRouteGraph)
	if err != nil {
		return types.RoutingGraph{}, err
	}

	// In the unlikely event that the previous route map is exists but is empty, set previousRouteMapFound to false and utilize all pools since
	// we don't have any information about routes to reason about liquidity
	previousRouteMap := convertToMap(&previousRouteGraph)
	if len(previousRouteMap.Graph) == 0 {
		previousRouteMapFound = false
	}

	// Retrieve minimum liquidity threshold from params
	minOsmoLiquidity := k.GetParams(ctx).MinOsmoValueForRoutes

	// Create a routingGraph to represent possible routes between tokens
	var routingGraph types.RoutingGraph

	// Iterate through the pools
	for _, pool := range pools {
		// If we were able to find a previous route map,
		// check if each pool meets the minimum liquidity threshold
		// If not, skip the pool
		if previousRouteMapFound {
			poolLiquidity := k.poolLiquidityToOSMO(ctx, pool, previousRouteMap)
			if poolLiquidity.LT(minOsmoLiquidity) {
				continue
			}
		}
		tokens := pool.GetPoolDenoms(ctx)
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

// getDirectOSMORouteWithMostLiquidity returns the route with the highest liquidity between an input denom and uosmo
func (k Keeper) getDirectOSMORouteWithMostLiquidity(ctx sdk.Context, inputDenom string, routeMap types.RoutingGraphMap) (uint64, error) {
	// Get all direct routes from the input denom to uosmo
	directRoutes := findRoutes(routeMap, inputDenom, k.stakingKeeper.BondDenom(ctx), 1)

	// Store liquidity for all direct routes found
	routeLiquidity := make(map[string]osmomath.Int)
	for _, route := range directRoutes {
		liquidity, err := k.getPoolLiquidityOfDenom(ctx, route[0].PoolId, k.stakingKeeper.BondDenom(ctx))
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

	// Find the route (single or double hop) with the highest liquidity
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
		return 0, fmt.Errorf("no route found with sufficient liquidity, likely no direct pairing with osmo")
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
func (k Keeper) inputAmountToOSMO(ctx sdk.Context, inputDenom string, amount osmomath.Int, routeMap types.RoutingGraphMap) (osmomath.Int, error) {
	if inputDenom == k.stakingKeeper.BondDenom(ctx) {
		return amount, nil
	}

	var route uint64
	var err error

	// Check if the route is cached
	if cachedRoute, ok := directRouteCache[inputDenom]; ok {
		route = cachedRoute
	} else {
		// If not, get the route and cache it
		route, err = k.getDirectOSMORouteWithMostLiquidity(ctx, inputDenom, routeMap)
		if err != nil {
			return osmomath.ZeroInt(), nil
		}
		directRouteCache[inputDenom] = route
	}

	var osmoPerInputToken osmomath.BigDec

	// Check if the spot price is cached
	spotPriceKey := fmt.Sprintf("%d:%s", route, inputDenom)
	if cachedSpotPrice, ok := spotPriceCache[spotPriceKey]; ok {
		osmoPerInputToken = cachedSpotPrice
	} else {
		// If not, calculate the spot price and cache it
		osmoPerInputToken, err = k.RouteCalculateSpotPrice(ctx, route, k.stakingKeeper.BondDenom(ctx), inputDenom)
		if err != nil {
			return osmomath.ZeroInt(), err
		}
		spotPriceCache[spotPriceKey] = osmoPerInputToken
	}

	// Convert the input denom to uosmo
	// Rounding is fine here
	uosmoAmount := amount.ToLegacyDec().Mul(osmoPerInputToken.Dec())
	return uosmoAmount.RoundInt(), nil
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

// poolLiquidityToOSMO returns the total liquidity of a pool in terms of uosmo
func (k Keeper) poolLiquidityToOSMO(ctx sdk.Context, pool types.PoolI, routeMap types.RoutingGraphMap) osmomath.Int {
	poolDenoms := pool.GetPoolDenoms(ctx)
	totalLiquidity := sdk.ZeroInt()
	for _, denom := range poolDenoms {
		liquidity, err := k.getPoolLiquidityOfDenom(ctx, pool.GetId(), denom)
		if err != nil {
			panic(err)
		}
		uosmoAmount, err := k.inputAmountToOSMO(ctx, denom, liquidity, routeMap)
		if err != nil {
			// no direct route found, so skip this denom
			continue
		}
		totalLiquidity = totalLiquidity.Add(uosmoAmount)
	}
	return totalLiquidity
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
