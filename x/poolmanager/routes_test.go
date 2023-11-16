package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestGetPoolModule tests that the correct pool module is returned for a given pool id.
// Additionally, validates that the expected errors are produced when expected.
func (s *KeeperTestSuite) TestDenomPairRoute() {
	tests := map[string]struct {
		expectedRoute []uint64
		expectError   error
	}{
		"direct route is best": {
			expectedRoute: []uint64{5},
		},
		"indirect route is best, route via eth": {
			expectedRoute: []uint64{4, 1},
		},
		"indirect route is best, route via bar": {
			expectedRoute: []uint64{3, 2},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolmanagerKeeper := s.App.PoolManagerKeeper

			s.PrepareBalancerPool()

			// Create pools for indirect route
			ethStake := s.PrepareConcentratedPoolWithCoins("eth", "stake") // pool 1
			barStake := s.PrepareConcentratedPoolWithCoins("bar", "stake") // pool 2
			btcBar := s.PrepareConcentratedPoolWithCoins("btc", "bar")     // pool 3
			btcEth := s.PrepareConcentratedPoolWithCoins("btc", "eth")     // pool 4

			// Create pool for direct route
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2000000000000000000)), sdk.NewCoin("adam", sdk.NewInt(2000000000000000000)))...) // pool 5

			// Create uosmo pairings to determine value in a base asset
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(100000000)), sdk.NewCoin("adam", sdk.NewInt(100000000)))...)
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(100000000000)), sdk.NewCoin("eth", sdk.NewInt(100000000)))...)
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(100000000)), sdk.NewCoin("stake", sdk.NewInt(100000000)))...)
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(100000000)), sdk.NewCoin("bar", sdk.NewInt(100000000)))...)

			// Create positions as per the test case, which determines what the best route is
			s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
			s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
			s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
			s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))

			// Set the routes, this normally happens at the end of an epoch or at time of upgrade to v21
			err := poolmanagerKeeper.SetDenomPairRoutes(s.Ctx)
			s.Require().NoError(err)

			// Get the route
			route, err := poolmanagerKeeper.GetDenomPairRoute(s.Ctx, "btc", "stake")
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRoute, route)
		})
	}
}
