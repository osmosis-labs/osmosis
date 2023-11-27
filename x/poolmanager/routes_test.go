package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

// Manually define a graph so we can test the getters
var graph = types.RoutingGraphMap{
	Graph: map[string]*types.InnerMap{
		"token1": {
			InnerMap: map[string]*types.Routes{
				"token2": {
					Routes: []*types.Route{
						{PoolId: 1, Token: "token2"},
					},
				},
				"token3": {
					Routes: []*types.Route{
						{PoolId: 2, Token: "token3"},
					},
				},
			},
		},
		"token2": {
			InnerMap: map[string]*types.Routes{
				"token3": {
					Routes: []*types.Route{
						{PoolId: 3, Token: "token3"},
					},
				},
				"token4": {
					Routes: []*types.Route{
						{PoolId: 4, Token: "token4"},
					},
				},
			},
		},
		"token3": {
			InnerMap: map[string]*types.Routes{
				"token4": {
					Routes: []*types.Route{
						{PoolId: 5, Token: "token4"},
					},
				},
			},
		},
	},
}

func TestFindDirectRoute(t *testing.T) {
	routes := poolmanager.FindRoutes(graph, "token1", "token2", 1)[0]

	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	if routes[0].PoolId != 1 || routes[0].Token != "token2" {
		t.Errorf("Unexpected route: %+v", routes[0])
	}
}

func TestFindTwoHopRoute(t *testing.T) {
	routes := poolmanager.FindRoutes(graph, "token1", "token3", 2)

	totalRoutes := 0
	for _, subRoutes := range routes {
		totalRoutes += len(subRoutes)
	}

	if totalRoutes != 2 {
		t.Errorf("Expected 2 routes, got %d", totalRoutes)
	}

	if routes[0][0].PoolId != 1 || routes[0][0].Token != "token2" {
		t.Errorf("Unexpected route: %+v", routes[0][0])
	}

	if routes[0][1].PoolId != 3 || routes[0][1].Token != "token3" {
		t.Errorf("Unexpected route: %+v", routes[0][1])
	}
}

func TestFindThreeHopRoute(t *testing.T) {
	routes := poolmanager.FindRoutes(graph, "token1", "token4", 3)

	totalRoutes := 0
	for _, subRoutes := range routes {
		totalRoutes += len(subRoutes)
	}

	if totalRoutes != 3 {
		t.Errorf("Expected 3 routes, got %d", totalRoutes)
	}

	if routes[0][0].PoolId != 1 || routes[0][0].Token != "token2" {
		t.Errorf("Unexpected route: %+v", routes[0][0])
	}

	if routes[0][1].PoolId != 3 || routes[0][1].Token != "token3" {
		t.Errorf("Unexpected route: %+v", routes[0][1])
	}

	if routes[0][2].PoolId != 5 || routes[0][2].Token != "token4" {
		t.Errorf("Unexpected route: %+v", routes[0][2])
	}
}

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

func (s *KeeperTestSuite) TestGetDirectOSMORouteWithMostLiquidity() {
	// Create two identical pools
	pool1 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("uosmo", "bar")
	pool2 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("uosmo", "bar")

	// Pool 1 now has more liquidity
	s.CreateFullRangePosition(pool1, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("bar", sdk.NewInt(10000000))))

	// Set routes and get it from state
	_, err := s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err := s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// Pool 1 should be the route with most liquidity
	route, err := s.App.PoolManagerKeeper.GetDirectOSMORouteWithMostLiquidity(s.Ctx, "bar", routeMap)
	s.Require().NoError(err)
	s.Require().Equal(pool1.GetId(), route)

	// Pool 2 now has more liquidity
	s.CreateFullRangePosition(pool2, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(20000000)), sdk.NewCoin("bar", sdk.NewInt(20000000))))

	// Set routes and get it from state
	_, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err = s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// Pool 2 should be the route with most liquidity
	route, err = s.App.PoolManagerKeeper.GetDirectOSMORouteWithMostLiquidity(s.Ctx, "bar", routeMap)
	s.Require().NoError(err)
	s.Require().Equal(pool2.GetId(), route)
}

func (s *KeeperTestSuite) TestInputAmountToOSMO() {
	// Set up a pool paired with uosmo at 1:1 ratio
	pool1 := s.PrepareConcentratedPoolWithCoins("uosmo", "bar")
	s.CreateFullRangePosition(pool1, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("bar", sdk.NewInt(10000000))))

	// Routes not set, should return 0 with no error
	osmoAmt, err := s.App.PoolManagerKeeper.InputAmountToOSMO(s.Ctx, "bar", sdk.NewInt(10000000), types.RoutingGraphMap{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.ZeroInt(), osmoAmt)

	// Set routes and get it from state
	_, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err := s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// With 1:1 ratio, input amount should be equal to output amount
	osmoAmt, err = s.App.PoolManagerKeeper.InputAmountToOSMO(s.Ctx, "bar", sdk.NewInt(10000000), routeMap)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), osmoAmt)

	// Set up a pool paired with uosmo at 2:1 ratio
	pool2 := s.PrepareConcentratedPoolWithCoins("uosmo", "foo")
	s.CreateFullRangePosition(pool2, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(20000000)), sdk.NewCoin("foo", sdk.NewInt(10000000))))

	// Set routes and get it from state
	_, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err = s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// With 2:1 ratio, input amount should be half of the output amount
	osmoAmt, err = s.App.PoolManagerKeeper.InputAmountToOSMO(s.Ctx, "foo", sdk.NewInt(10000000), routeMap)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(20000000), osmoAmt)
}

func (s *KeeperTestSuite) TestGetPoolLiquidityOfDenom() {
	poolInfo := s.PrepareAllSupportedPools()

	// Balancer
	poolLiq, err := s.App.PoolManagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, poolInfo.BalancerPoolID, "bar")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(5000000), poolLiq)

	// StableSwap
	poolLiq, err = s.App.PoolManagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, poolInfo.StableSwapPoolID, "bar")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), poolLiq)

	// Cosmwasm
	token := sdk.NewCoins(sdk.NewCoin("axlusdc", sdk.NewInt(10000000)))
	s.FundAcc(s.TestAccs[0], token)
	s.JoinTransmuterPool(s.TestAccs[0], poolInfo.CosmWasmPoolID, token)
	poolLiq, err = s.App.PoolManagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, poolInfo.CosmWasmPoolID, "axlusdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), poolLiq)

	// Concentrated
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolInfo.ConcentratedPoolID)
	s.Require().NoError(err)
	clPoolExtension, ok := clPool.(cltypes.ConcentratedPoolExtension)
	s.Require().True(ok)
	s.CreateFullRangePosition(clPoolExtension, sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000)), sdk.NewCoin("eth", sdk.NewInt(10000000))))
	poolLiq, err = s.App.PoolManagerKeeper.GetPoolLiquidityOfDenom(s.Ctx, poolInfo.ConcentratedPoolID, "eth")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), poolLiq)
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
