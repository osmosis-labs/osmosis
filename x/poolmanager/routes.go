package poolmanager

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

var (
	OSMO = "uosmo"
)

// SetDenomPairRoutes sets the route map to be used for route calculations
func (k Keeper) SetDenomPairRoutes(ctx sdk.Context) (types.RoutingGraph, error) {
	// Get all the pools
	pools, err := k.AllPools(ctx)
	if err != nil {
		return types.RoutingGraph{}, err
	}

	// Create a routingGraph to represent possible routes between tokens
	var routingGraph types.RoutingGraph

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
