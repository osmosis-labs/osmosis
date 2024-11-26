package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// TestGetNumberOfTrades tests GetNumberOfTrades and IncrementNumberOfTrades
func (s *KeeperTestSuite) TestGetNumberOfTrades() {
	// Should be zero by default
	numberOfTrades, err := s.App.ProtoRevKeeper.GetNumberOfTrades(s.Ctx)
	s.Require().Error(err)
	s.Require().Equal(osmomath.NewInt(0), numberOfTrades)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.IncrementNumberOfTrades(s.Ctx)
	s.Require().NoError(err)

	// Check the updated result
	numberOfTrades, err = s.App.ProtoRevKeeper.GetNumberOfTrades(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), numberOfTrades)
}

// TestGetProfitsByDenom tests GetProfitsByDenom, UpdateProfitsByDenom, and GetAllProfits
func (s *KeeperTestSuite) TestGetProfitsByDenom() {
	// Should be zero by default
	profits, err := s.App.ProtoRevKeeper.GetProfitsByDenom(s.Ctx, types.OsmosisDenomination)
	s.Require().Error(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.ZeroInt()), profits)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByDenom(s.Ctx, types.OsmosisDenomination, osmomath.NewInt(9000))
	s.Require().NoError(err)

	// Check the updated result
	profits, err = s.App.ProtoRevKeeper.GetProfitsByDenom(s.Ctx, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(9000)), profits)

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByDenom(s.Ctx, types.OsmosisDenomination, osmomath.NewInt(5000))
	s.Require().NoError(err)

	// Check the updated result after the second trade
	profits, err = s.App.ProtoRevKeeper.GetProfitsByDenom(s.Ctx, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(14000)), profits)

	// Check the result of GetAllProfits
	allProfits := s.App.ProtoRevKeeper.GetAllProfits(s.Ctx)
	s.Require().Equal([]sdk.Coin{{Denom: types.OsmosisDenomination, Amount: osmomath.NewInt(14000)}}, allProfits)

	// Pseudo execute a third trade in a different denom
	err = s.App.ProtoRevKeeper.UpdateProfitsByDenom(s.Ctx, "Atom", osmomath.NewInt(1000))
	s.Require().NoError(err)

	// Check the result of GetAllProfits
	allProfits = s.App.ProtoRevKeeper.GetAllProfits(s.Ctx)
	s.Require().Equal([]sdk.Coin{{Denom: "Atom", Amount: osmomath.NewInt(1000)}, {Denom: types.OsmosisDenomination, Amount: osmomath.NewInt(14000)}}, allProfits)
}

// TestGetTradesByRoute tests GetTradesByRoute, IncrementTradesByRoute, and GetAllRoutes
func (s *KeeperTestSuite) TestGetTradesByRoute() {
	// There should be no routes that have been executed by default
	routes, err := s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(0, len(routes))

	// Check the number of trades for a route that has not been executed
	trades, err := s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().Error(err)
	s.Require().Equal(osmomath.NewInt(0), trades)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.IncrementTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().NoError(err)

	// Check the updated result
	trades, err = s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), trades)

	// Check the result of GetAllRoutes
	routes, err = s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(1, len(routes))
	s.Require().Equal([]uint64{1, 2, 3}, routes[0])

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.IncrementTradesByRoute(s.Ctx, []uint64{2, 3, 4})
	s.Require().NoError(err)

	// Check the updated result after the second trade
	trades, err = s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{2, 3, 4})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), trades)

	// Check the result of GetAllRoutes
	routes, err = s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(2, len(routes))
	s.Require().Equal([]uint64{1, 2, 3}, routes[0])
	s.Require().Equal([]uint64{2, 3, 4}, routes[1])
}

// TestGetProfitsByRoute tests GetProfitsByRoute, UpdateProfitsByRoute, and GetAllProfitsByRoute
func (s *KeeperTestSuite) TestGetProfitsByRoute() {
	// There should be no profits that have been executed by default
	profits := s.App.ProtoRevKeeper.GetAllProfitsByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().Equal([]sdk.Coin{}, profits)

	// Check the profits for a route that has not been executed
	profit, err := s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	s.Require().Error(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.ZeroInt()), profit)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination, osmomath.NewInt(1000))
	s.Require().NoError(err)

	// Check the updated result
	profit, err = s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)), profit)

	// Check the result of GetAllProfitsByRoute
	profits = s.App.ProtoRevKeeper.GetAllProfitsByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().Equal([]sdk.Coin{{Denom: types.OsmosisDenomination, Amount: osmomath.NewInt(1000)}}, profits)

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, "Atom", osmomath.NewInt(2000))
	s.Require().NoError(err)

	// Check the updated result after the second trade
	profit, err = s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, "Atom")
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin("Atom", osmomath.NewInt(2000)), profit)

	// Check the result of GetAllProfitsByRoute
	profits = s.App.ProtoRevKeeper.GetAllProfitsByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().Contains(profits, sdk.Coin{Denom: types.OsmosisDenomination, Amount: osmomath.NewInt(1000)})
	s.Require().Contains(profits, sdk.Coin{Denom: "Atom", Amount: osmomath.NewInt(2000)})
}

// TestUpdateStatistics tests UpdateStatistics which is a wrapper for much of the statistics keeper
// functionality.
func (s *KeeperTestSuite) TestUpdateStatistics() {
	// Pseudo execute a trade
	err := s.App.ProtoRevKeeper.UpdateStatistics(s.Ctx,
		poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}},
		types.OsmosisDenomination, osmomath.NewInt(1000),
	)
	s.Require().NoError(err)

	// Check the result of GetTradesByRoute
	trades, err := s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), trades)

	// Check the result of GetProfitsByRoute
	profit, err := s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)), profit)

	// Check the result of GetAllRoutes
	routes, err := s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(1, len(routes))

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.UpdateStatistics(s.Ctx,
		poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}, {TokenOutDenom: "", PoolId: 4}},
		types.OsmosisDenomination, osmomath.NewInt(1100),
	)
	s.Require().NoError(err)

	// Check the result of GetTradesByRoute
	trades, err = s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{2, 3, 4})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), trades)

	// Check the result of GetProfitsByRoute
	profit, err = s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{2, 3, 4}, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1100)), profit)

	// Check the result of GetAllRoutes
	routes, err = s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(2, len(routes))
}

func (s *KeeperTestSuite) TestGetSetCyclicArbProfitTrackerValue() {
	tests := map[string]struct {
		firstCyclicArbValue  sdk.Coins
		secondCyclicArbValue sdk.Coins
	}{
		"happy path: replace single coin with increased single coin": {
			firstCyclicArbValue:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			secondCyclicArbValue: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(200))),
		},
		"replace single coin with decreased single coin": {
			firstCyclicArbValue:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			secondCyclicArbValue: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(50))),
		},
		"replace single coin with different denom": {
			firstCyclicArbValue:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			secondCyclicArbValue: sdk.NewCoins(sdk.NewCoin("usdc", osmomath.NewInt(100))),
		},
		"replace single coin with multiple coins": {
			firstCyclicArbValue:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100))),
			secondCyclicArbValue: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100)), sdk.NewCoin("usdc", osmomath.NewInt(200))),
		},
		"replace multiple coins with single coin": {
			firstCyclicArbValue:  sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(100)), sdk.NewCoin("usdc", osmomath.NewInt(200))),
			secondCyclicArbValue: sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(200))),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerValue(s.Ctx))

			s.App.ProtoRevKeeper.SetCyclicArbProfitTrackerValue(s.Ctx, tc.firstCyclicArbValue)
			actualFirstCyclicArbValue := s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerValue(s.Ctx)
			s.Require().Equal(tc.firstCyclicArbValue, actualFirstCyclicArbValue)

			s.App.ProtoRevKeeper.SetCyclicArbProfitTrackerValue(s.Ctx, tc.secondCyclicArbValue)
			actualSecondCyclicArbValue := s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerValue(s.Ctx)
			s.Require().Equal(tc.secondCyclicArbValue, actualSecondCyclicArbValue)
		})
	}
}

func (s *KeeperTestSuite) TestGetSetCyclicArbProfitTrackerStartHeight() {
	tests := map[string]struct {
		firstCyclicArbStartHeight  int64
		secondCyclicArbStartHeight int64
	}{
		"replace tracker height with a higher height": {
			firstCyclicArbStartHeight:  100,
			secondCyclicArbStartHeight: 5000,
		},
		"replace tracker height with a lower height": {
			firstCyclicArbStartHeight:  100,
			secondCyclicArbStartHeight: 50,
		},
		"replace tracker height back to zero": {
			firstCyclicArbStartHeight:  100,
			secondCyclicArbStartHeight: 0,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			s.Require().Empty(s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerStartHeight(s.Ctx))

			s.App.ProtoRevKeeper.SetCyclicArbProfitTrackerStartHeight(s.Ctx, tc.firstCyclicArbStartHeight)
			actualFirstCyclicArbStartHeight := s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerStartHeight(s.Ctx)
			s.Require().Equal(tc.firstCyclicArbStartHeight, actualFirstCyclicArbStartHeight)

			s.App.ProtoRevKeeper.SetCyclicArbProfitTrackerStartHeight(s.Ctx, tc.secondCyclicArbStartHeight)
			actualSecondCyclicArbStartHeight := s.App.ProtoRevKeeper.GetCyclicArbProfitTrackerStartHeight(s.Ctx)
			s.Require().Equal(tc.secondCyclicArbStartHeight, actualSecondCyclicArbStartHeight)
		})
	}
}
