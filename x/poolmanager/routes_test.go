package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v23/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
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

	// Prepare pools
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("bar", sdk.NewInt(10000000000)))
	s.PrepareConcentratedPoolWithCoins("uosmo", "foo")
	s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{"uosmo", "uion"})

	// Create a pool that determines what the value of uosmo is in terms of USDC
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.MinValueForRoute.Denom = "usdc"
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
	osmoUsdPool := s.PrepareConcentratedPoolWithCoins("uosmo", s.App.PoolManagerKeeper.GetParams(s.Ctx).MinValueForRoute.Denom)
	s.CreateFullRangePosition(osmoUsdPool, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("usdc", sdk.NewInt(10000000000))))

	// Set routes in state
	routingGraph, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(routingGraph)

	// Get routes from state and compare to expected
	routingMap, err = s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)
	expectedRoutingMap = poolmanager.ConvertToMap(&routingGraph)
	s.Require().Equal(expectedRoutingMap, routingMap)

	// 3 entries are expected, since foo and uion pools do not reach the min liquidity threshold
	s.Require().Equal(3, len(routingGraph.Entries))

	// Change min liquidity threshold to 0
	poolManagerParams = s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.MinValueForRoute.Amount = sdk.NewInt(0)
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	// Set routes in state
	routingGraph, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(routingGraph)

	// Get routes from state and compare to expected
	routingMap, err = s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)
	expectedRoutingMap = poolmanager.ConvertToMap(&routingGraph)
	s.Require().Equal(expectedRoutingMap, routingMap)

	// 5 entries for all denoms (uosmo, uion, bar, foo, usdc)
	s.Require().Equal(5, len(routingGraph.Entries))
}

func (s *KeeperTestSuite) TestGetDirectRouteWithMostLiquidity() {
	// Create two identical pools
	pool1 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("uosmo", "bar")
	pool2 := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("uosmo", "bar")

	// Create a pool to denominate uosmo in terms of usdc
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.MinValueForRoute.Denom = "usdc"
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
	osmoUsdPool := s.PrepareConcentratedPoolWithCoins("uosmo", s.App.PoolManagerKeeper.GetParams(s.Ctx).MinValueForRoute.Denom)
	s.CreateFullRangePosition(osmoUsdPool, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("usdc", sdk.NewInt(10000000000))))

	// Pool 1 now has more liquidity
	s.CreateFullRangePosition(pool1, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("bar", sdk.NewInt(10000000))))

	// Set routes and get it from state
	_, err := s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err := s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// Pool 1 should be the route with most liquidity
	route, err := s.App.PoolManagerKeeper.GetDirectRouteWithMostLiquidity(s.Ctx, "bar", "uosmo", routeMap)
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
	route, err = s.App.PoolManagerKeeper.GetDirectRouteWithMostLiquidity(s.Ctx, "bar", "uosmo", routeMap)
	s.Require().NoError(err)
	s.Require().Equal(pool2.GetId(), route)
}

func (s *KeeperTestSuite) TestInputAmountToTargetDenom() {
	// Set up a pool paired with uosmo at 1:1 ratio
	pool1 := s.PrepareConcentratedPoolWithCoins("uosmo", "bar")
	s.CreateFullRangePosition(pool1, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("bar", sdk.NewInt(10000000000))))

	// Set a usdc pool to denominate uosmo in terms of usdc
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.MinValueForRoute.Denom = "usdc"
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
	osmoUsdPool := s.PrepareConcentratedPoolWithCoins("uosmo", s.App.PoolManagerKeeper.GetParams(s.Ctx).MinValueForRoute.Denom)
	s.CreateFullRangePosition(osmoUsdPool, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("usdc", sdk.NewInt(10000000000))))

	// Routes not set, should return 0 with no error
	osmoAmt, err := s.App.PoolManagerKeeper.InputAmountToTargetDenom(s.Ctx, "bar", "uosmo", sdk.NewInt(10000000), types.RoutingGraphMap{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.ZeroInt(), osmoAmt)

	// Set routes and get it from state
	_, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err := s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// With 1:1 ratio, input amount should be equal to output amount
	osmoAmt, err = s.App.PoolManagerKeeper.InputAmountToTargetDenom(s.Ctx, "bar", "uosmo", sdk.NewInt(10000000), routeMap)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), osmoAmt)

	// Set up a pool paired with uosmo at 2:1 ratio
	pool2 := s.PrepareConcentratedPoolWithCoins("uosmo", "foo")
	s.CreateFullRangePosition(pool2, sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(20000000000)), sdk.NewCoin("foo", sdk.NewInt(10000000000))))

	// Set routes and get it from state
	_, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	routeMap, err = s.App.PoolManagerKeeper.GetRouteMap(s.Ctx)
	s.Require().NoError(err)

	// With 2:1 ratio, input amount should be half of the output amount
	osmoAmt, err = s.App.PoolManagerKeeper.InputAmountToTargetDenom(s.Ctx, "foo", "uosmo", sdk.NewInt(10000000), routeMap)
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

func (s *KeeperTestSuite) TestPoolLiquidityToTargetDenom() {
	poolInfo := s.PrepareAllSupportedPools()
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("bar", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("baz", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("foo", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("uosmo", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("axlusdc", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("gravusdc", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000000)), sdk.NewCoin("eth", sdk.NewInt(10000000000)))

	// Set routes
	_, routeMap, err := s.App.PoolManagerKeeper.GenerateAllDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)

	// Balancer
	// 5000000 bar, 5000000 baz, 5000000 foo, 5000000 uosmo = 20000000 usdc
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfo.BalancerPoolID)
	s.Require().NoError(err)
	poolLiq, err := s.App.PoolManagerKeeper.PoolLiquidityToTargetDenom(s.Ctx, balancerPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(20000000), poolLiq)

	// StableSwap
	// 10000000 bar, 10000000 baz, 10000000 foo = 30000000 usdc
	stableSwapPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfo.StableSwapPoolID)
	s.Require().NoError(err)
	poolLiq, err = s.App.PoolManagerKeeper.PoolLiquidityToTargetDenom(s.Ctx, stableSwapPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(30000000), poolLiq)

	// Cosmwasm
	// 10000000 axlusdc, 10000000 gravusdc = 20000000 usdc
	token := sdk.NewCoins(sdk.NewCoin("axlusdc", sdk.NewInt(10000000)), sdk.NewCoin("gravusdc", sdk.NewInt(10000000)))
	s.FundAcc(s.TestAccs[0], token)
	s.JoinTransmuterPool(s.TestAccs[0], poolInfo.CosmWasmPoolID, token)
	cosmWasmPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfo.CosmWasmPoolID)
	s.Require().NoError(err)
	poolLiq, err = s.App.PoolManagerKeeper.PoolLiquidityToTargetDenom(s.Ctx, cosmWasmPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(20000000), poolLiq)

	// Concentrated
	// 10000000 eth, 9999991 usdc = 19999991 usdc
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolInfo.ConcentratedPoolID)
	s.Require().NoError(err)
	clPoolExtension, ok := clPool.(cltypes.ConcentratedPoolExtension)
	s.Require().True(ok)
	s.CreateFullRangePosition(clPoolExtension, sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000)), sdk.NewCoin("eth", sdk.NewInt(10000000))))
	poolLiq, err = s.App.PoolManagerKeeper.PoolLiquidityToTargetDenom(s.Ctx, clPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(19999991), poolLiq)
}

func (s *KeeperTestSuite) TestPoolLiquidityFromOSMOToTargetDenom() {
	poolInfo := s.PrepareAllSupportedPools()

	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("bar", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("baz", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("foo", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("axlusdc", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("gravusdc", sdk.NewInt(10000000000)))
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000000)), sdk.NewCoin("eth", sdk.NewInt(10000000000)))

	// 0.5 spot price
	s.PrepareBalancerPoolWithCoins(sdk.NewCoin("usdc", sdk.NewInt(20000000000)), sdk.NewCoin("uosmo", sdk.NewInt(10000000000)))

	// Generate routes
	_, routeMap, err := s.App.PoolManagerKeeper.GenerateAllDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)

	// Balancer
	// 5000000 bar, 5000000 baz, 5000000 foo, 5000000 uosmo = 20000000 usdc * 0.5 spot price = 10000000 uosmo
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfo.BalancerPoolID)
	s.Require().NoError(err)
	poolLiq, err := s.App.PoolManagerKeeper.PoolLiquidityFromOSMOToTargetDenom(s.Ctx, balancerPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), poolLiq)

	// StableSwap
	// 10000000 bar, 10000000 baz, 10000000 foo = 30000000 usdc * 0.5 spot price = 15000000 uosmo
	stableSwapPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfo.StableSwapPoolID)
	s.Require().NoError(err)
	poolLiq, err = s.App.PoolManagerKeeper.PoolLiquidityFromOSMOToTargetDenom(s.Ctx, stableSwapPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(15000000), poolLiq)

	// Cosmwasm
	// 10000000 axlusdc, 10000000 gravusdc = 20000000 usdc * 0.5 spot price = 10000000 uosmo
	token := sdk.NewCoins(sdk.NewCoin("axlusdc", sdk.NewInt(10000000)), sdk.NewCoin("gravusdc", sdk.NewInt(10000000)))
	s.FundAcc(s.TestAccs[0], token)
	s.JoinTransmuterPool(s.TestAccs[0], poolInfo.CosmWasmPoolID, token)
	cosmWasmPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, poolInfo.CosmWasmPoolID)
	s.Require().NoError(err)
	poolLiq, err = s.App.PoolManagerKeeper.PoolLiquidityFromOSMOToTargetDenom(s.Ctx, cosmWasmPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000000), poolLiq)

	// Concentrated
	// 10000000 eth, 4999996 usdc = 14999996 usdc * 0.5 spot price = 7499998 uosmo
	clPool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, poolInfo.ConcentratedPoolID)
	s.Require().NoError(err)
	clPoolExtension, ok := clPool.(cltypes.ConcentratedPoolExtension)
	s.Require().True(ok)
	s.CreateFullRangePosition(clPoolExtension, sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(10000000)), sdk.NewCoin("eth", sdk.NewInt(10000000))))
	poolLiq, err = s.App.PoolManagerKeeper.PoolLiquidityFromOSMOToTargetDenom(s.Ctx, clPool, routeMap, "usdc")
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(7499998), poolLiq)
}
