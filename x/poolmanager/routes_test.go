package poolmanager_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v20/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

func (s *KeeperTestSuite) TestGetSetDenomPairRoutes() {
	// Set routes in state
	routingGraph, err := s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Empty(routingGraph)

	// Get routes from state and compare to expected
	routingMap, err := s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)
	expectedRoutingMap := poolmanager.ConvertToMap(&routingGraph)
	s.Require().Equal(expectedRoutingMap, routingMap)

	s.PrepareAllSupportedPools()

	// Set routes in state
	routingGraph, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(routingGraph)

	// Get routes from state and compare to expected
	routingMap, err = s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)
	expectedRoutingMap = poolmanager.ConvertToMap(&routingGraph)
	s.Require().Equal(expectedRoutingMap, routingMap)

	// 4 pools, 2 routes per pool
	s.Require().Equal(8, len(routingGraph.Entries))
}

func TestConvertToMap(t *testing.T) {
	// Define a RoutingGraph
	routingGraph := &types.RoutingGraph{
		Entries: []*types.RoutingGraphEntry{
			{
				Key: "token1",
				Value: &types.Inner{
					Entries: []*types.InnerMapEntry{
						{
							Key: "token2",
							Value: &types.Routes{
								Routes: []*types.Route{
									{PoolId: 1, Token: "token2"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Call the function
	result := poolmanager.ConvertToMap(routingGraph)

	// Check the result
	if len(result.Graph) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(result.Graph))
	}

	innerMap, ok := result.Graph["token1"]
	if !ok {
		t.Errorf("Expected to find 'token1' key")
	}

	routes, ok := innerMap.InnerMap["token2"]
	if !ok {
		t.Errorf("Expected to find 'token2' key")
	}

	if len(routes.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes.Routes))
	}

	if routes.Routes[0].PoolId != 1 || routes.Routes[0].Token != "token2" {
		t.Errorf("Unexpected route: %+v", routes.Routes[0])
	}
}
