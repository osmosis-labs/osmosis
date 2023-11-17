package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/types"
)

// TestGetPoolModule tests that the correct pool module is returned for a given pool id.
// Additionally, validates that the expected errors are produced when expected.
func (s *KeeperTestSuite) TestDenomPairRoute() {
	tests := map[string]struct {
		setup         func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension)
		expectedRoute []uint64
		expectError   error
	}{
		"direct route is best": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1100000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1100000000000000000))))
			},
			expectedRoute: []uint64{6},
		},
		"indirect route is best, route via eth": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1100000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1100000000000000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
			},
			expectedRoute: []uint64{5, 2},
		},
		"indirect route is best, route via bar": {
			setup: func(ethStake, barStake, btcBar, btcEth, btcStake cltypes.ConcentratedPoolExtension) {
				// Create positions as per the test case, which determines what the best route is
				s.CreateFullRangePosition(ethStake, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(barStake, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1000000000000000000)), sdk.NewCoin("stake", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcBar, sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(1100000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1100000000000000000))))
				s.CreateFullRangePosition(btcEth, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
				s.CreateFullRangePosition(btcStake, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000000000000000000)), sdk.NewCoin("btc", sdk.NewInt(1000000000000000000))))
			},
			expectedRoute: []uint64{4, 3},
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

			tc.setup(ethStake, barStake, btcBar, btcEth, btcStake)

			// Create uosmo pairings to determine value in a base asset
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("btc", sdk.NewInt(10000000)))...)
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("eth", sdk.NewInt(10000000)))...)
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("stake", sdk.NewInt(10000000)))...)
			s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000)), sdk.NewCoin("bar", sdk.NewInt(10000000)))...)

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
