package poolmanager_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	routes := poolmanager.FindDirectRoute(graph, "token1", "token2")

	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	if routes[0].PoolId != 1 || routes[0].Token != "token2" {
		t.Errorf("Unexpected route: %+v", routes[0])
	}
}

func TestFindTwoHopRoute(t *testing.T) {
	routes := poolmanager.FindTwoHopRoute(graph, "token1", "token3")

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
	routes := poolmanager.FindThreeHopRoute(graph, "token1", "token4")

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

func (s *KeeperTestSuite) TestSetDenomPairRoutes() {
	routingGraph, err := s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Empty(routingGraph)

	s.PrepareAllSupportedPools()

	routingGraph, err = s.App.PoolManagerKeeper.SetDenomPairRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(routingGraph)

	// 4 pools, 2 routes per pool
	s.Require().Equal(8, len(routingGraph.Entries))
}

func (s *KeeperTestSuite) TestGetDenomPairRoute() {
	tests := map[string]struct {
		setup         func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension)
		tokenIn       sdk.Coin
		outDenom      string
		expectedRoute []uint64
		expectError   error
	}{
		"single hop is best": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("stake", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(20000000)), sdk.NewCoin("stake", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(21000000)), sdk.NewCoin("btc", sdk.NewInt(21000000))))
			},
			tokenIn:       sdk.NewCoin("btc", sdk.NewInt(10000000)),
			outDenom:      "stake",
			expectedRoute: []uint64{6},
		},
		"double hop is best, route via eth": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("stake", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(20000000)), sdk.NewCoin("stake", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(21000000)), sdk.NewCoin("btc", sdk.NewInt(21000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10000000)), sdk.NewCoin("btc", sdk.NewInt(10000000))))
			},
			tokenIn:       sdk.NewCoin("btc", sdk.NewInt(10000000)),
			outDenom:      "stake",
			expectedRoute: []uint64{5, 2},
		},
		"double hop is best, route via bar": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000)), sdk.NewCoin("stake", sdk.NewInt(10000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(21000000)), sdk.NewCoin("stake", sdk.NewInt(21000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(21000000)), sdk.NewCoin("btc", sdk.NewInt(21000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10000000)), sdk.NewCoin("btc", sdk.NewInt(10000000))))
			},
			tokenIn:       sdk.NewCoin("btc", sdk.NewInt(10000000)),
			outDenom:      "stake",
			expectedRoute: []uint64{4, 3},
		},
		"flip denoms should flip route": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(10000000)), sdk.NewCoin("stake", sdk.NewInt(10000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(21000000)), sdk.NewCoin("stake", sdk.NewInt(21000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(21000000)), sdk.NewCoin("btc", sdk.NewInt(21000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10000000)), sdk.NewCoin("btc", sdk.NewInt(10000000))))
			},
			tokenIn:       sdk.NewCoin("stake", sdk.NewInt(10000000)),
			outDenom:      "btc",
			expectedRoute: []uint64{3, 4},
		},
		"three route": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("stake", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(20000000)), sdk.NewCoin("stake", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(20000000)), sdk.NewCoin("btc", sdk.NewInt(20000000))))
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(10000000)),
			outDenom:      "stbtc",
			expectedRoute: []uint64{8, 9, 10},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.PrepareBalancerPool() // pool 1

			// Create cl pools for indirect routes
			ethStake := s.PrepareConcentratedPoolWithCoins("eth", "stake") // pool 2
			barStake := s.PrepareConcentratedPoolWithCoins("bar", "stake") // pool 3
			btcBar := s.PrepareConcentratedPoolWithCoins("btc", "bar")     // pool 4
			btcEth := s.PrepareConcentratedPoolWithCoins("btc", "eth")     // pool 5

			// Create cl pool for direct routes
			btcStake := s.PrepareConcentratedPoolWithCoins("btc", "stake") // pool 6

			// Create cw pools for direct route
			s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{"btc", "stake"}) // pool 7

			// Create three route (eth -> bar -> test -> stbtc)
			s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("eth", "bar")                 // pool 8
			s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("test", "bar")                // pool 9
			transPool := s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{"test", "stbtc"}) // pool 10
			fund := sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stbtc", sdk.NewInt(1000000000000000000)))
			s.FundAcc(s.TestAccs[0], fund)
			s.JoinTransmuterPool(s.TestAccs[0], transPool.GetId(), fund)

			tc.setup(ethStake, barStake, btcBar, btcEth, btcStake)

			// Create uosmo pairings to determine value in a base asset
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000)), sdk.NewCoin("btc", sdk.NewInt(1000)))...)   // pool 11
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000)), sdk.NewCoin("eth", sdk.NewInt(1000)))...)   // pool 12
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000)), sdk.NewCoin("stake", sdk.NewInt(1000)))...) // pool 13
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000)), sdk.NewCoin("bar", sdk.NewInt(1000)))...)   // pool 14
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000)), sdk.NewCoin("test", sdk.NewInt(1000)))...)  // pool 15
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000)), sdk.NewCoin("stbtc", sdk.NewInt(1000)))...) // pool 16

			// Run epoch, which sets the routes each day
			s.App.EpochsKeeper.AfterEpochEnd(s.Ctx, "day", 1)

			// Get the route
			route, err := poolmanagerKeeper.GetDenomPairRoute(s.Ctx, tc.tokenIn, tc.outDenom)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRoute, route)
		})
	}
}

func TestParseRouteKey(t *testing.T) {
	routeKey := `[pool_id:1 token:"uion" pool_id:2 token:"ibc/123ABC" pool_id:3 token:"uosmo"]`
	expected := []types.Route{
		{PoolId: 1, Token: "uion"},
		{PoolId: 2, Token: "ibc/123ABC"},
		{PoolId: 3, Token: "uosmo"},
	}

	routes, err := poolmanager.ParseRouteKey(routeKey)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(routes, expected) {
		t.Errorf("Expected %+v, got %+v", expected, routes)
	}
}
