package poolmanager

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var (
	OSMO                 = "uosmo"
	superfluidMultiplier = sdk.MustNewDecFromStr("1.5")
)

// Define a structure to represent the routing graph
type RoutingGraph map[string]map[string][]Route

// Define a structure to represent a route
type Route struct {
	PoolID uint64
	Token  string
}

// Function to add an edge to the graph
func (g RoutingGraph) AddEdge(start, end, token string, poolID uint64) {
	if g[start] == nil {
		g[start] = make(map[string][]Route)
	}
	g[start][end] = append(g[start][end], Route{PoolID: poolID, Token: token})
}

// Function to find all direct routes between two tokens
func FindDirectRoute(g RoutingGraph, start, end string) []Route {
	if routes, exists := g[start][end]; exists {
		for i := range routes {
			routes[i].Token = end
		}
		return routes
	}
	return nil
}

// Function to find all two-hop routes between two tokens
func FindTwoHopRoute(g RoutingGraph, start, end string) [][]Route {
	var routeRoutes [][]Route

	for token, routes := range g[start] {
		if endRoutes, exists := g[token][end]; exists {
			for _, startRoute := range routes {
				startRoute.Token = token
				for _, endRoute := range endRoutes {
					endRoute.Token = end
					routeRoutes = append(routeRoutes, []Route{startRoute, endRoute})
				}
			}
		}
	}

	return routeRoutes
}

// Function to find all three-hop routes between two tokens
func FindThreeHopRoute(g RoutingGraph, start, end string) [][]Route {
	var routeRoutes [][]Route

	for token1, routes1 := range g[start] {
		for token2, routes2 := range g[token1] {
			if token2 == start || token2 == end {
				continue
			}
			if endRoutes, exists := g[token2][end]; exists {
				for _, startRoute := range routes1 {
					startRoute.Token = token1
					for _, middleRoute := range routes2 {
						middleRoute.Token = token2
						for _, endRoute := range endRoutes {
							endRoute.Token = end
							routeRoutes = append(routeRoutes, []Route{startRoute, middleRoute, endRoute})
						}
					}
				}
			}
		}
	}

	return routeRoutes
}

// SetDenomPairRoutes sets the route map to be used for route calculations
func (k *Keeper) SetDenomPairRoutes(ctx sdk.Context) error {
	// Get all the pools
	pools, err := k.AllPools(ctx)
	if err != nil {
		return err
	}

	// Create a routingGraph to represent possible routes between tokens
	routingGraph := make(RoutingGraph)

	// Iterate through the pools
	for _, pool := range pools {
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

	k.routeMap = routingGraph
	return nil
}

// // GetDenomPairRoute returns the route with the highest liquidity between two tokens
// func (k *Keeper) GetDenomPairRoute(ctx sdk.Context, inputDenom, outputDenom string) ([]uint64, error) {
// 	// temp, remove later
// 	err := k.SetDenomPairRoutes(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if k.routeMap == nil {
// 		return nil, fmt.Errorf("route map not set")
// 	}

// 	// Get all direct routes
// 	directPoolIDs := FindDirectRoute(k.routeMap, inputDenom, outputDenom)

// 	// Get all two-hop routes
// 	var twoHopPoolIDs [][]uint64
// 	if inputDenom != OSMO && outputDenom != OSMO {
// 		twoHopPoolIDs = FindTwoHopRoute(k.routeMap, inputDenom, outputDenom)
// 	}

// 	var threeHopPoolIDs [][]uint64
// 	if inputDenom != OSMO && outputDenom != OSMO {
// 		threeHopPoolIDs = FindThreeHopRoute(k.routeMap, inputDenom, outputDenom)
// 	}

// 	// Map to store the total liquidity of each route (using string as key)
// 	routeLiquidity := make(map[string]osmomath.Int)

// 	// Check liquidity for all direct routes
// 	for _, poolID := range directPoolIDs {
// 		pool, err := k.GetPool(ctx, poolID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		poolDenoms := pool.GetPoolDenoms(ctx)
// 		liqInOsmo := osmomath.ZeroInt()
// 		for _, denom := range poolDenoms {
// 			liquidity, err := k.GetPoolLiquidityOfDenom(ctx, poolID, denom)
// 			if err != nil {
// 				return nil, err
// 			}
// 			liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
// 			if err != nil {
// 				return nil, err
// 			}

// 			if pool.GetType() == types.Concentrated {
// 				liqInOsmoInternal = liqInOsmoInternal.ToLegacyDec().Mul(superfluidMultiplier).TruncateInt()
// 			}

// 			// Multiply the liquidity by six. This is because we are comparing single routes to a max of three-hop routes.
// 			// To make this simple and comparable, we just multiply the single route liquidity by six.
// 			liqInOsmo = liqInOsmo.Add(liqInOsmoInternal.Mul(osmomath.NewIntFromUint64(6)))
// 		}
// 		routeKey := fmt.Sprintf("%v", poolID)
// 		routeLiquidity[routeKey] = liqInOsmo
// 	}

// 	// Check liquidity for all two-hop routes
// 	for _, poolIDs := range twoHopPoolIDs {
// 		totalLiquidityInOsmo := osmomath.ZeroInt()
// 		routeKey := fmt.Sprintf("%v", poolIDs)
// 		for _, poolID := range poolIDs {
// 			pool, err := k.GetPool(ctx, poolID)
// 			if err != nil {
// 				return nil, err
// 			}
// 			poolDenoms := pool.GetPoolDenoms(ctx)
// 			liqInOsmo := osmomath.ZeroInt()
// 			for _, denom := range poolDenoms {
// 				liquidity, err := k.GetPoolLiquidityOfDenom(ctx, poolID, denom)
// 				if err != nil {
// 					return nil, err
// 				}
// 				liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
// 				if err != nil {
// 					return nil, err
// 				}

// 				if pool.GetType() == types.Concentrated {
// 					liqInOsmoInternal = liqInOsmoInternal.ToLegacyDec().Mul(superfluidMultiplier).TruncateInt()
// 				}
// 				liqInOsmo = liqInOsmo.Add(liqInOsmoInternal)
// 			}

// 			// Multiply the liquidity by three. This is because we are comparing double route to a max of three-hop routes.
// 			// To make this simple and comparable, we just multiply the single route liquidity by three.
// 			totalLiquidityInOsmo = totalLiquidityInOsmo.Add(liqInOsmo.Mul(osmomath.NewIntFromUint64(3)))
// 		}
// 		routeLiquidity[routeKey] = totalLiquidityInOsmo
// 	}

// 	// Check liquidity for all three-hop routes
// 	for _, poolIDs := range threeHopPoolIDs {
// 		totalLiquidityInOsmo := osmomath.ZeroInt()
// 		routeKey := fmt.Sprintf("%v", poolIDs)
// 		for _, poolID := range poolIDs {
// 			pool, err := k.GetPool(ctx, poolID)
// 			if err != nil {
// 				return nil, err
// 			}
// 			poolDenoms := pool.GetPoolDenoms(ctx)
// 			liqInOsmo := osmomath.ZeroInt()
// 			for _, denom := range poolDenoms {
// 				liquidity, err := k.GetPoolLiquidityOfDenom(ctx, poolID, denom)
// 				if err != nil {
// 					return nil, err
// 				}
// 				liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
// 				if err != nil {
// 					return nil, err
// 				}

// 				if pool.GetType() == types.Concentrated {
// 					liqInOsmoInternal = liqInOsmoInternal.ToLegacyDec().Mul(superfluidMultiplier).TruncateInt()
// 				}
// 				liqInOsmo = liqInOsmo.Add(liqInOsmoInternal)
// 			}
// 			totalLiquidityInOsmo = totalLiquidityInOsmo.Add(liqInOsmo)
// 		}
// 		routeLiquidity[routeKey] = totalLiquidityInOsmo
// 	}

// 	// Extract and sort the keys from the routeLiquidity map
// 	// This ensures deterministic selection of the best route
// 	var keys []string
// 	for k := range routeLiquidity {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	// Find the route (single or double hop) with the highest liquidity
// 	var bestRouteKey string
// 	maxLiquidity := osmomath.ZeroInt()
// 	for _, routeKey := range keys {
// 		liquidity := routeLiquidity[routeKey]
// 		// Update best route if a higher liquidity is found,
// 		// or if the liquidity is equal but the routeKey is encountered earlier in the sorted order
// 		if liquidity.GT(maxLiquidity) || (liquidity.Equal(maxLiquidity) && bestRouteKey == "") {
// 			bestRouteKey = routeKey
// 			maxLiquidity = liquidity
// 		}
// 	}

// 	if bestRouteKey == "" {
// 		fmt.Println("No suitable route found.")
// 		return nil, fmt.Errorf("no route found with sufficient liquidity")
// 	}

// 	// Convert the best route key back to []uint64
// 	var bestRoute []uint64
// 	cleanedRouteKey := strings.Trim(bestRouteKey, "[]")
// 	idStrs := strings.Split(cleanedRouteKey, " ")

// 	for _, idStr := range idStrs {
// 		id, err := strconv.ParseUint(idStr, 10, 64)
// 		if err != nil {
// 			return nil, fmt.Errorf("error parsing pool ID: %v", err)
// 		}
// 		bestRoute = append(bestRoute, id)
// 	}

// 	// Return the route with the highest liquidity
// 	fmt.Printf("Route Selected: %v \n", strings.Join(strings.Split(bestRouteKey, " "), " -> "))
// 	return bestRoute, nil
// }

func (k *Keeper) GetDenomPairRoute(ctx sdk.Context, inputCoin sdk.Coin, outputDenom string) ([]uint64, error) {
	if k.routeMap == nil {
		return nil, fmt.Errorf("route map not set")
	}
	inputDenom := inputCoin.Denom

	if k.routeMap == nil {
		return nil, fmt.Errorf("route map not set")
	}

	// Get all direct routes
	directPoolIDs := FindDirectRoute(k.routeMap, inputDenom, outputDenom)

	// Get all two-hop routes
	var twoHopPoolIDs [][]Route
	if inputDenom != OSMO && outputDenom != OSMO {
		twoHopPoolIDs = FindTwoHopRoute(k.routeMap, inputDenom, outputDenom)
	}

	var threeHopPoolIDs [][]Route
	if inputDenom != OSMO && outputDenom != OSMO {
		threeHopPoolIDs = FindThreeHopRoute(k.routeMap, inputDenom, outputDenom)
	}

	// Map to store the total liquidity of each route (using string as key)
	routeLiquidity := make(map[string]osmomath.Int)

	// Check liquidity for all direct routes
	for _, route := range directPoolIDs {
		pool, err := k.GetPool(ctx, route.PoolID)
		if err != nil {
			return nil, err
		}
		poolDenoms := pool.GetPoolDenoms(ctx)
		liqInOsmo := osmomath.ZeroInt()
		for _, denom := range poolDenoms {
			liquidity, err := k.GetPoolLiquidityOfDenom(ctx, route.PoolID, denom)
			if err != nil {
				return nil, err
			}
			liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
			if err != nil {
				return nil, err
			}

			if pool.GetType() == types.Concentrated {
				liqInOsmoInternal = liqInOsmoInternal.ToLegacyDec().Mul(superfluidMultiplier).TruncateInt()
			}

			// Multiply the liquidity by six. This is because we are comparing single routes to a max of three-hop routes.
			// To make this simple and comparable, we just multiply the single route liquidity by six.
			liqInOsmo = liqInOsmo.Add(liqInOsmoInternal.Mul(osmomath.NewIntFromUint64(6)))
		}
		routeKey := fmt.Sprintf("%v", route)
		routeLiquidity[routeKey] = liqInOsmo
	}

	// Check liquidity for all two-hop routes
	for _, routes := range twoHopPoolIDs {
		totalLiquidityInOsmo := osmomath.ZeroInt()
		routeKey := fmt.Sprintf("%v", routes)
		for _, route := range routes {
			pool, err := k.GetPool(ctx, route.PoolID)
			if err != nil {
				return nil, err
			}
			poolDenoms := pool.GetPoolDenoms(ctx)
			liqInOsmo := osmomath.ZeroInt()
			for _, denom := range poolDenoms {
				liquidity, err := k.GetPoolLiquidityOfDenom(ctx, route.PoolID, denom)
				if err != nil {
					return nil, err
				}
				liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
				if err != nil {
					return nil, err
				}

				if pool.GetType() == types.Concentrated {
					liqInOsmoInternal = liqInOsmoInternal.ToLegacyDec().Mul(superfluidMultiplier).TruncateInt()
				}
				liqInOsmo = liqInOsmo.Add(liqInOsmoInternal)
			}

			// Multiply the liquidity by three. This is because we are comparing double route to a max of three-hop routes.
			// To make this simple and comparable, we just multiply the single route liquidity by three.
			totalLiquidityInOsmo = totalLiquidityInOsmo.Add(liqInOsmo.Mul(osmomath.NewIntFromUint64(3)))
		}
		routeLiquidity[routeKey] = totalLiquidityInOsmo
	}

	// Check liquidity for all three-hop routes
	for _, routes := range threeHopPoolIDs {
		totalLiquidityInOsmo := osmomath.ZeroInt()
		routeKey := fmt.Sprintf("%v", routes)
		for _, route := range routes {
			pool, err := k.GetPool(ctx, route.PoolID)
			if err != nil {
				return nil, err
			}
			poolDenoms := pool.GetPoolDenoms(ctx)
			liqInOsmo := osmomath.ZeroInt()
			for _, denom := range poolDenoms {
				liquidity, err := k.GetPoolLiquidityOfDenom(ctx, route.PoolID, denom)
				if err != nil {
					return nil, err
				}
				liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
				if err != nil {
					return nil, err
				}

				if pool.GetType() == types.Concentrated {
					liqInOsmoInternal = liqInOsmoInternal.ToLegacyDec().Mul(superfluidMultiplier).TruncateInt()
				}
				liqInOsmo = liqInOsmo.Add(liqInOsmoInternal)
			}
			totalLiquidityInOsmo = totalLiquidityInOsmo.Add(liqInOsmo)
		}
		routeLiquidity[routeKey] = totalLiquidityInOsmo
	}

	// Extract and sort the keys from the routeLiquidity map
	// This ensures deterministic selection of the best route
	var keys []string
	for k := range routeLiquidity {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Variables to store the best route key for each hop count
	var bestSingleHopRouteKey, bestDoubleHopRouteKey, bestTripleHopRouteKey string
	maxSingleHopLiquidity, maxDoubleHopLiquidity, maxTripleHopLiquidity := osmomath.ZeroInt(), osmomath.ZeroInt(), osmomath.ZeroInt()

	for _, routeKey := range keys {
		liquidity := routeLiquidity[routeKey]
		hopCount := len(strings.Fields(routeKey)) / 2

		fmt.Println("hopCount: ", hopCount)
		fmt.Println("routeKey: ", routeKey)

		// Update best route based on hop count and liquidity
		switch hopCount {
		case 1: // Single hop
			if liquidity.GT(maxSingleHopLiquidity) {
				maxSingleHopLiquidity = liquidity
				bestSingleHopRouteKey = routeKey
			}
		case 2: // Double hop
			if liquidity.GT(maxDoubleHopLiquidity) {
				maxDoubleHopLiquidity = liquidity
				bestDoubleHopRouteKey = routeKey
			}
		case 3: // Triple hop
			if liquidity.GT(maxTripleHopLiquidity) {
				maxTripleHopLiquidity = liquidity
				bestTripleHopRouteKey = routeKey
			}
		}
	}

	// Construct the result map
	result := make(map[string][]Route)

	singleHopRoute, err := parseRouteKey(bestSingleHopRouteKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing single hop route key: %v", err)
	}
	result["singleHop"] = singleHopRoute

	fmt.Println("bestDoubleHopRouteKey: ", bestDoubleHopRouteKey)
	doubleHopRoute, err := parseRouteKey(bestDoubleHopRouteKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing double hop route key: %v", err)
	}
	result["doubleHop"] = doubleHopRoute

	tripleHopRoute, err := parseRouteKey(bestTripleHopRouteKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing triple hop route key: %v", err)
	}
	result["tripleHop"] = tripleHopRoute

	maxAmtOut := sdk.ZeroInt()
	var maxKey string

	for key, value := range result {
		swapRoute := []types.SwapAmountInRoute{}
		for _, route := range value {
			// Construct SwapAmountInRoute for each poolID
			swapRoute = append(swapRoute, types.SwapAmountInRoute{
				PoolId:        route.PoolID,
				TokenOutDenom: route.Token,
			})
		}

		// Call MultihopEstimateOutGivenExactAmountIn with swapRoute
		amtOut, err := k.MultihopEstimateOutGivenExactAmountIn(ctx, swapRoute, inputCoin)
		if err != nil {
			continue
		}

		// Update maxAmtOut and maxKey if the current amtOut is greater
		if amtOut.GT(maxAmtOut) {
			maxAmtOut = amtOut
			maxKey = key
		}
	}

	var poolIDs []uint64
	for _, route := range result[maxKey] {
		poolIDs = append(poolIDs, route.PoolID)
	}

	return poolIDs, nil
}

// Helper function to parse route key into a slice of Route
func parseRouteKey(routeKey string) ([]Route, error) {
	var route []Route
	if routeKey == "" {
		return route, nil
	}
	cleanedRouteKey := strings.Trim(routeKey, "[]")
	idStrs := strings.Split(cleanedRouteKey, " ")

	for i := 0; i < len(idStrs); i += 2 {
		id, err := strconv.ParseUint(strings.Trim(idStrs[i], "{}"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing pool ID: %v", err)
		}
		token := strings.Trim(idStrs[i+1], "{}")
		route = append(route, Route{PoolID: id, Token: token})
	}
	return route, nil
}

// GetDirectOSMORouteWithMostLiquidity returns the route with the highest liquidity between an input denom and uosmo
func (k Keeper) GetDirectOSMORouteWithMostLiquidity(ctx sdk.Context, inputDenom string) (uint64, error) {
	if k.routeMap == nil {
		return 0, fmt.Errorf("route map not set")
	}

	// Get all direct routes from the input denom to uosmo
	directRoutes := FindDirectRoute(k.routeMap, inputDenom, OSMO)

	// Store liquidity for all direct routes found
	routeLiquidity := make(map[string]osmomath.Int)
	for _, route := range directRoutes {
		liquidity, err := k.GetPoolLiquidityOfDenom(ctx, route.PoolID, OSMO)
		if err != nil {
			return 0, err
		}
		routeKey := fmt.Sprintf("%v", route.PoolID)
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

// Transform an input denom and its amount to uosmo
// If a route is not found, returns 0 with no error.
func (k Keeper) InputDenomToOSMO(ctx sdk.Context, inputDenom string, amount osmomath.Int) (osmomath.Int, error) {
	if inputDenom == OSMO {
		return amount, nil
	}
	// start by getting the route from the input denom to uosmo
	route, err := k.GetDirectOSMORouteWithMostLiquidity(ctx, inputDenom)
	if err != nil {
		return osmomath.ZeroInt(), nil
	}

	// spot price of uosmo to input denom
	osmoPerInputToken, err := k.RouteCalculateSpotPrice(ctx, route, OSMO, inputDenom)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	// convert the input denom to uosmo
	uosmoAmount := amount.ToLegacyDec().Mul(osmoPerInputToken.Dec())
	return uosmoAmount.TruncateInt(), nil
}

// GetPoolLiquidityOfDenom returns the liquidity of a denom in a pool.
// This calls different methods depending on the pool type.
func (k Keeper) GetPoolLiquidityOfDenom(ctx sdk.Context, poolId uint64, outputDenom string) (osmomath.Int, error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	// Check the pool type, and check the pool liquidity based on the type
	switch pool.GetType() {
	case types.Balancer:
		// transform from poolI to cfmmPool
		pool, ok := pool.(gammtypes.CFMMPoolI)
		if !ok {
			return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
		}
		totalPoolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		return totalPoolLiquidity.AmountOf(outputDenom), nil
	case types.Stableswap:
		pool, ok := pool.(gammtypes.CFMMPoolI)
		if !ok {
			return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
		}
		totalPoolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		return totalPoolLiquidity.AmountOf(outputDenom), nil
	case types.Concentrated:
		poolAddress := pool.GetAddress()
		poolAddressBalances := k.bankKeeper.GetAllBalances(ctx, poolAddress)
		return poolAddressBalances.AmountOf(outputDenom), nil
	case types.CosmWasm:
		pool, ok := pool.(cosmwasmpooltypes.CosmWasmExtension)
		if !ok {
			return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
		}
		totalPoolLiquidity := pool.GetTotalPoolLiquidity(ctx)
		return totalPoolLiquidity.AmountOf(outputDenom), nil
	default:
		return osmomath.ZeroInt(), fmt.Errorf("invalid pool type")
	}
}
