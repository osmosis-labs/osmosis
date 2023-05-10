package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// TestGetNumberOfTrades tests GetNumberOfTrades and IncrementNumberOfTrades
func (suite *KeeperTestSuite) TestGetNumberOfTrades() {
	// Should be zero by default
	numberOfTrades, err := suite.App.ProtoRevKeeper.GetNumberOfTrades(suite.Ctx)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.NewInt(0), numberOfTrades)

	// Pseudo execute a trade
	err = suite.App.ProtoRevKeeper.IncrementNumberOfTrades(suite.Ctx)
	suite.Require().NoError(err)

	// Check the updated result
	numberOfTrades, err = suite.App.ProtoRevKeeper.GetNumberOfTrades(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), numberOfTrades)
}

// TestGetProfitsByDenom tests GetProfitsByDenom, UpdateProfitsByDenom, and GetAllProfits
func (suite *KeeperTestSuite) TestGetProfitsByDenom() {
	// Should be zero by default
	profits, err := suite.App.ProtoRevKeeper.GetProfitsByDenom(suite.Ctx, types.OsmosisDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.ZeroInt()), profits)

	// Pseudo execute a trade
	err = suite.App.ProtoRevKeeper.UpdateProfitsByDenom(suite.Ctx, types.OsmosisDenomination, sdk.NewInt(9000))
	suite.Require().NoError(err)

	// Check the updated result
	profits, err = suite.App.ProtoRevKeeper.GetProfitsByDenom(suite.Ctx, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(9000)), profits)

	// Pseudo execute a second trade
	err = suite.App.ProtoRevKeeper.UpdateProfitsByDenom(suite.Ctx, types.OsmosisDenomination, sdk.NewInt(5000))
	suite.Require().NoError(err)

	// Check the updated result after the second trade
	profits, err = suite.App.ProtoRevKeeper.GetProfitsByDenom(suite.Ctx, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(14000)), profits)

	// Check the result of GetAllProfits
	allProfits := suite.App.ProtoRevKeeper.GetAllProfits(suite.Ctx)
	suite.Require().Equal([]sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(14000)}}, allProfits)

	// Pseudo execute a third trade in a different denom
	err = suite.App.ProtoRevKeeper.UpdateProfitsByDenom(suite.Ctx, "Atom", sdk.NewInt(1000))
	suite.Require().NoError(err)

	// Check the result of GetAllProfits
	allProfits = suite.App.ProtoRevKeeper.GetAllProfits(suite.Ctx)
	suite.Require().Equal([]sdk.Coin{{Denom: "Atom", Amount: sdk.NewInt(1000)}, {Denom: types.OsmosisDenomination, Amount: sdk.NewInt(14000)}}, allProfits)
}

// TestGetTradesByRoute tests GetTradesByRoute, IncrementTradesByRoute, and GetAllRoutes
func (suite *KeeperTestSuite) TestGetTradesByRoute() {
	// There should be no routes that have been executed by default
	routes, err := suite.App.ProtoRevKeeper.GetAllRoutes(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(routes))

	// Check the number of trades for a route that has not been executed
	trades, err := suite.App.ProtoRevKeeper.GetTradesByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().Error(err)
	suite.Require().Equal(sdk.NewInt(0), trades)

	// Pseudo execute a trade
	err = suite.App.ProtoRevKeeper.IncrementTradesByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().NoError(err)

	// Check the updated result
	trades, err = suite.App.ProtoRevKeeper.GetTradesByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), trades)

	// Check the result of GetAllRoutes
	routes, err = suite.App.ProtoRevKeeper.GetAllRoutes(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(routes))
	suite.Require().Equal([]uint64{1, 2, 3}, routes[0])

	// Pseudo execute a second trade
	err = suite.App.ProtoRevKeeper.IncrementTradesByRoute(suite.Ctx, []uint64{2, 3, 4})
	suite.Require().NoError(err)

	// Check the updated result after the second trade
	trades, err = suite.App.ProtoRevKeeper.GetTradesByRoute(suite.Ctx, []uint64{2, 3, 4})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), trades)

	// Check the result of GetAllRoutes
	routes, err = suite.App.ProtoRevKeeper.GetAllRoutes(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(routes))
	suite.Require().Equal([]uint64{1, 2, 3}, routes[0])
	suite.Require().Equal([]uint64{2, 3, 4}, routes[1])
}

// TestGetProfitsByRoute tests GetProfitsByRoute, UpdateProfitsByRoute, and GetAllProfitsByRoute
func (suite *KeeperTestSuite) TestGetProfitsByRoute() {
	// There should be no profits that have been executed by default
	profits := suite.App.ProtoRevKeeper.GetAllProfitsByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().Equal([]sdk.Coin{}, profits)

	// Check the profits for a route that has not been executed
	profit, err := suite.App.ProtoRevKeeper.GetProfitsByRoute(suite.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	suite.Require().Error(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.ZeroInt()), profit)

	// Pseudo execute a trade
	err = suite.App.ProtoRevKeeper.UpdateProfitsByRoute(suite.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination, sdk.NewInt(1000))
	suite.Require().NoError(err)

	// Check the updated result
	profit, err = suite.App.ProtoRevKeeper.GetProfitsByRoute(suite.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), profit)

	// Check the result of GetAllProfitsByRoute
	profits = suite.App.ProtoRevKeeper.GetAllProfitsByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().Equal([]sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(1000)}}, profits)

	// Pseudo execute a second trade
	err = suite.App.ProtoRevKeeper.UpdateProfitsByRoute(suite.Ctx, []uint64{1, 2, 3}, "Atom", sdk.NewInt(2000))
	suite.Require().NoError(err)

	// Check the updated result after the second trade
	profit, err = suite.App.ProtoRevKeeper.GetProfitsByRoute(suite.Ctx, []uint64{1, 2, 3}, "Atom")
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin("Atom", sdk.NewInt(2000)), profit)

	// Check the result of GetAllProfitsByRoute
	profits = suite.App.ProtoRevKeeper.GetAllProfitsByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().Contains(profits, sdk.Coin{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(1000)})
	suite.Require().Contains(profits, sdk.Coin{Denom: "Atom", Amount: sdk.NewInt(2000)})
}

// TestUpdateStatistics tests UpdateStatistics which is a wrapper for much of the statistics keeper
// functionality.
func (suite *KeeperTestSuite) TestUpdateStatistics() {
	// Pseudo execute a trade
	err := suite.App.ProtoRevKeeper.UpdateStatistics(suite.Ctx,
		poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}},
		types.OsmosisDenomination, sdk.NewInt(1000),
	)
	suite.Require().NoError(err)

	// Check the result of GetTradesByRoute
	trades, err := suite.App.ProtoRevKeeper.GetTradesByRoute(suite.Ctx, []uint64{1, 2, 3})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), trades)

	// Check the result of GetProfitsByRoute
	profit, err := suite.App.ProtoRevKeeper.GetProfitsByRoute(suite.Ctx, []uint64{1, 2, 3}, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), profit)

	// Check the result of GetAllRoutes
	routes, err := suite.App.ProtoRevKeeper.GetAllRoutes(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(routes))

	// Pseudo execute a second trade
	err = suite.App.ProtoRevKeeper.UpdateStatistics(suite.Ctx,
		poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}, {TokenOutDenom: "", PoolId: 4}},
		types.OsmosisDenomination, sdk.NewInt(1100),
	)
	suite.Require().NoError(err)

	// Check the result of GetTradesByRoute
	trades, err = suite.App.ProtoRevKeeper.GetTradesByRoute(suite.Ctx, []uint64{2, 3, 4})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), trades)

	// Check the result of GetProfitsByRoute
	profit, err = suite.App.ProtoRevKeeper.GetProfitsByRoute(suite.Ctx, []uint64{2, 3, 4}, types.OsmosisDenomination)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1100)), profit)

	// Check the result of GetAllRoutes
	routes, err = suite.App.ProtoRevKeeper.GetAllRoutes(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(routes))
}
