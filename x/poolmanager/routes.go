package poolmanager

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// Define a structure to represent the graph
type Graph map[string]map[string][]uint64

// Function to add an edge to the graph
func (g Graph) AddEdge(start, end string, poolID uint64) {
	if g[start] == nil {
		g[start] = make(map[string][]uint64)
	}
	g[start][end] = append(g[start][end], poolID)
}

// Function to find all direct routes between two tokens
func HasDirectRoute(g Graph, start, end string) ([]uint64, bool) {
	if pools, exists := g[start][end]; exists {
		return pools, true
	}
	return nil, false
}

// Function to find all two-hop routes between two tokens
func FindTwoHopRoute(g Graph, start, end string) ([][]string, [][]uint64) {
	var routes [][]string
	var routePoolIDs [][]uint64

	for token, pools := range g[start] {
		if endPools, exists := g[token][end]; exists {
			for _, startPoolID := range pools {
				for _, endPoolID := range endPools {
					routes = append(routes, []string{start, token, end})
					routePoolIDs = append(routePoolIDs, []uint64{startPoolID, endPoolID})
				}
			}
		}
	}

	return routes, routePoolIDs
}

func (k *Keeper) SetDenomPairRoutes(ctx sdk.Context) error {
	// Get all the pools
	pools, err := k.AllPools(ctx)
	if err != nil {
		return err
	}
	fmt.Println("pool length", len(pools))

	// Create a graph to represent possible routes between tokens
	graph := make(Graph)

	// Iterate through the pools
	for _, pool := range pools {
		// skip cosmwasmpool for now
		if pool.GetType() == types.CosmWasm {
			continue
		}
		tokens := pool.GetPoolDenoms(ctx)
		// fmt.Println("tokens", tokens)
		poolID := pool.GetId()
		// fmt.Println("poolID", poolID)
		// Create edges for all possible combinations of tokens
		for i := 0; i < len(tokens); i++ {
			for j := i + 1; j < len(tokens); j++ {
				graph.AddEdge(tokens[i], tokens[j], poolID)
				graph.AddEdge(tokens[j], tokens[i], poolID)
			}
		}
	}

	k.routeMap = graph
	// fmt.Println("routeMap", k.routeMap)
	return nil
}

func (k Keeper) GetDenomPairRoute(ctx sdk.Context, inputDenom, outputDenom string) ([]uint64, error) {
	fmt.Println("chceking route map")
	if k.routeMap == nil {
		fmt.Println("setting route map")
		err := k.SetDenomPairRoutes(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Get all direct routes
	directPoolIDs, _ := HasDirectRoute(k.routeMap, inputDenom, outputDenom)

	fmt.Println("directPoolIDs", directPoolIDs)

	// Get all two-hop routes
	_, twoHopPoolIDs := FindTwoHopRoute(k.routeMap, inputDenom, outputDenom)

	fmt.Println("twoHopPoolIDs", twoHopPoolIDs)

	// Map to store the total liquidity of each route (using string as key)
	routeLiquidity := make(map[string]osmomath.Int)

	// Check liquidity for all direct routes
	for _, poolID := range directPoolIDs {
		pool, err := k.GetPool(ctx, poolID)
		if err != nil {
			return nil, err
		}
		poolDenoms := pool.GetPoolDenoms(ctx)
		liqInOsmo := osmomath.ZeroInt()
		for _, denom := range poolDenoms {
			liquidity, err := k.GetPoolLiquidityOfDenom(ctx, poolID, denom)
			if err != nil {
				return nil, err
			}
			liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
			if err != nil {
				return nil, err
			}
			liqInOsmo = liqInOsmoInternal.Add(liquidity)
		}
		routeKey := fmt.Sprintf("%v", poolID)
		fmt.Println("routeKey", routeKey)
		routeLiquidity[routeKey] = liqInOsmo
	}

	// Check liquidity for all two-hop routes
	for _, poolIDs := range twoHopPoolIDs {
		totalLiquidityInOsmo := osmomath.ZeroInt()
		routeKey := fmt.Sprintf("%v", poolIDs)
		for _, poolID := range poolIDs {
			pool, err := k.GetPool(ctx, poolID)
			if err != nil {
				return nil, err
			}
			poolDenoms := pool.GetPoolDenoms(ctx)
			liqInOsmo := osmomath.ZeroInt()
			for _, denom := range poolDenoms {
				liquidity, err := k.GetPoolLiquidityOfDenom(ctx, poolID, denom)
				if err != nil {
					return nil, err
				}
				liqInOsmoInternal, err := k.InputDenomToOSMO(ctx, denom, liquidity)
				if err != nil {
					return nil, err
				}
				liqInOsmo = liqInOsmoInternal.Add(liquidity)
			}
			fmt.Println("liqInOsmo", liqInOsmo)
			fmt.Println("poolID", poolID)
			fmt.Println("outputDenom", outputDenom)
			totalLiquidityInOsmo = totalLiquidityInOsmo.Add(liqInOsmo)
		}
		fmt.Println("routeKey", routeKey)
		routeLiquidity[routeKey] = totalLiquidityInOsmo
	}

	// Find the route with the highest liquidity
	fmt.Println("routeLiquidity", routeLiquidity)
	var bestRouteKey string
	maxLiquidity := osmomath.ZeroInt()
	for routeKey, liquidity := range routeLiquidity {
		if liquidity.GTE(maxLiquidity) {
			bestRouteKey = routeKey
			maxLiquidity = liquidity
		}
	}

	if bestRouteKey == "" {
		fmt.Println("No suitable route found.")
		return nil, fmt.Errorf("no route found with sufficient liquidity")
	}

	// Convert the best route key back to []uint64
	var bestRoute []uint64
	cleanedRouteKey := strings.Trim(bestRouteKey, "[]")
	idStrs := strings.Split(cleanedRouteKey, " ")

	fmt.Println("idStrs", idStrs)
	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing pool ID: %v", err)
		}
		bestRoute = append(bestRoute, id)
	}

	// Return the route with the highest liquidity
	fmt.Printf("Route Selected: %v via Pool IDs: %v\n", strings.Join(strings.Split(bestRouteKey, " "), " -> "), bestRoute)
	return bestRoute, nil
}

func (k Keeper) GetDirectOSMORouteWithMostLiquidity(ctx sdk.Context, inputDenom string) (uint64, error) {
	fmt.Println("chceking route map")
	if k.routeMap == nil {
		fmt.Println("setting route map")
		err := k.SetDenomPairRoutes(ctx)
		if err != nil {
			return 0, err
		}
	}

	// Get all direct routes
	directPoolIDs, _ := HasDirectRoute(k.routeMap, inputDenom, "uosmo")

	// Check liquidity for all direct routes
	routeLiquidity := make(map[string]osmomath.Int)
	for _, poolID := range directPoolIDs {
		liquidity, err := k.GetPoolLiquidityOfDenom(ctx, poolID, "uosmo")
		if err != nil {
			return 0, err
		}
		routeKey := fmt.Sprintf("%v", poolID)
		fmt.Println("routeKey", routeKey)
		routeLiquidity[routeKey] = liquidity
	}

	// Find the route with the highest liquidity
	fmt.Println("routeLiquidity", routeLiquidity)
	var bestRouteKey string
	maxLiquidity := osmomath.ZeroInt()
	for routeKey, liquidity := range routeLiquidity {
		if liquidity.GTE(maxLiquidity) {
			bestRouteKey = routeKey
			maxLiquidity = liquidity
		}
	}

	if bestRouteKey == "" {
		fmt.Println("No suitable route found.")
		return 0, fmt.Errorf("no route found with sufficient liquidity")
	}

	// Convert the best route key back to []uint64
	var bestRoute []uint64
	cleanedRouteKey := strings.Trim(bestRouteKey, "[]")
	idStrs := strings.Split(cleanedRouteKey, " ")

	fmt.Println("idStrs", idStrs)
	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing pool ID: %v", err)
		}
		bestRoute = append(bestRoute, id)
	}

	// Return the route with the highest liquidity
	fmt.Printf("Route Selected: %v via Pool IDs: %v\n", strings.Join(strings.Split(bestRouteKey, " "), " -> "), bestRoute)
	return bestRoute[0], nil
}

func (k Keeper) InputDenomToOSMO(ctx sdk.Context, inputDenom string, amount osmomath.Int) (osmomath.Int, error) {
	if inputDenom == "uosmo" {
		return amount, nil
	}
	// start by getting the route from the input denom to uosmo
	route, err := k.GetDirectOSMORouteWithMostLiquidity(ctx, inputDenom)
	if err != nil {
		return osmomath.ZeroInt(), nil
	}

	// spot price of uosmo to input denom
	osmoPerInputToken, err := k.RouteCalculateSpotPrice(ctx, route, "uosmo", inputDenom)
	if err != nil {
		return osmomath.ZeroInt(), err
	}

	// convert the input denom to uosmo
	uosmoAmount := amount.ToLegacyDec().Mul(osmoPerInputToken.Dec())
	return uosmoAmount.TruncateInt(), nil
}

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
		fmt.Println("poolAddressBalances", poolAddressBalances)
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
