package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v19/x/protorev/types"
)

// TestGetNumberOfTrades tests GetNumberOfTrades and IncrementNumberOfTrades
func (s *KeeperTestSuite) TestGetNumberOfTrades() {
	// Should be zero by default
	numberOfTrades, err := s.App.ProtoRevKeeper.GetNumberOfTrades(s.Ctx)
	s.Require().Error(err)
	s.Require().Equal(sdk.NewInt(0), numberOfTrades)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.IncrementNumberOfTrades(s.Ctx)
	s.Require().NoError(err)

	// Check the updated result
	numberOfTrades, err = s.App.ProtoRevKeeper.GetNumberOfTrades(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(1), numberOfTrades)
}

// TestGetProfitsByDenom tests GetProfitsByDenom, UpdateProfitsByDenom, and GetAllProfits
func (s *KeeperTestSuite) TestGetProfitsByDenom() {
	// Should be zero by default
	profits, err := s.App.ProtoRevKeeper.GetProfitsByDenom(s.Ctx, types.OsmosisDenomination)
	s.Require().Error(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.ZeroInt()), profits)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByDenom(s.Ctx, types.OsmosisDenomination, sdk.NewInt(9000))
	s.Require().NoError(err)

	// Check the updated result
	profits, err = s.App.ProtoRevKeeper.GetProfitsByDenom(s.Ctx, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(9000)), profits)

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByDenom(s.Ctx, types.OsmosisDenomination, sdk.NewInt(5000))
	s.Require().NoError(err)

	// Check the updated result after the second trade
	profits, err = s.App.ProtoRevKeeper.GetProfitsByDenom(s.Ctx, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(14000)), profits)

	// Check the result of GetAllProfits
	allProfits := s.App.ProtoRevKeeper.GetAllProfits(s.Ctx)
	s.Require().Equal([]sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(14000)}}, allProfits)

	// Pseudo execute a third trade in a different denom
	err = s.App.ProtoRevKeeper.UpdateProfitsByDenom(s.Ctx, "Atom", sdk.NewInt(1000))
	s.Require().NoError(err)

	// Check the result of GetAllProfits
	allProfits = s.App.ProtoRevKeeper.GetAllProfits(s.Ctx)
	s.Require().Equal([]sdk.Coin{{Denom: "Atom", Amount: sdk.NewInt(1000)}, {Denom: types.OsmosisDenomination, Amount: sdk.NewInt(14000)}}, allProfits)
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
	s.Require().Equal(sdk.NewInt(0), trades)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.IncrementTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().NoError(err)

	// Check the updated result
	trades, err = s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(1), trades)

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
	s.Require().Equal(sdk.NewInt(1), trades)

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
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.ZeroInt()), profit)

	// Pseudo execute a trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination, sdk.NewInt(1000))
	s.Require().NoError(err)

	// Check the updated result
	profit, err = s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), profit)

	// Check the result of GetAllProfitsByRoute
	profits = s.App.ProtoRevKeeper.GetAllProfitsByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().Equal([]sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(1000)}}, profits)

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.UpdateProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, "Atom", sdk.NewInt(2000))
	s.Require().NoError(err)

	// Check the updated result after the second trade
	profit, err = s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, "Atom")
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin("Atom", sdk.NewInt(2000)), profit)

	// Check the result of GetAllProfitsByRoute
	profits = s.App.ProtoRevKeeper.GetAllProfitsByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().Contains(profits, sdk.Coin{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(1000)})
	s.Require().Contains(profits, sdk.Coin{Denom: "Atom", Amount: sdk.NewInt(2000)})
}

// TestUpdateStatistics tests UpdateStatistics which is a wrapper for much of the statistics keeper
// functionality.
func (s *KeeperTestSuite) TestUpdateStatistics() {
	// Pseudo execute a trade
	err := s.App.ProtoRevKeeper.UpdateStatistics(s.Ctx,
		poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}},
		types.OsmosisDenomination, sdk.NewInt(1000),
	)
	s.Require().NoError(err)

	// Check the result of GetTradesByRoute
	trades, err := s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{1, 2, 3})
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(1), trades)

	// Check the result of GetProfitsByRoute
	profit, err := s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), profit)

	// Check the result of GetAllRoutes
	routes, err := s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(1, len(routes))

	// Pseudo execute a second trade
	err = s.App.ProtoRevKeeper.UpdateStatistics(s.Ctx,
		poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}, {TokenOutDenom: "", PoolId: 4}},
		types.OsmosisDenomination, sdk.NewInt(1100),
	)
	s.Require().NoError(err)

	// Check the result of GetTradesByRoute
	trades, err = s.App.ProtoRevKeeper.GetTradesByRoute(s.Ctx, []uint64{2, 3, 4})
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(1), trades)

	// Check the result of GetProfitsByRoute
	profit, err = s.App.ProtoRevKeeper.GetProfitsByRoute(s.Ctx, []uint64{2, 3, 4}, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1100)), profit)

	// Check the result of GetAllRoutes
	routes, err = s.App.ProtoRevKeeper.GetAllRoutes(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(2, len(routes))
}
