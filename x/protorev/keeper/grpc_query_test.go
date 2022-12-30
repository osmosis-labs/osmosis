package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

// TestParams tests the query for params
func (suite *KeeperTestSuite) TestParams() {
	ctx := sdk.WrapSDKContext(suite.Ctx)
	expectedParams := types.DefaultParams()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expectedParams, res.Params)
}

// TestGetProtoRevNumberOfTrades tests the query for number of trades
func (suite *KeeperTestSuite) TestGetProtoRevNumberOfTrades() {
	// Initially should throw an error
	_, err := suite.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	suite.Require().Error(err)

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	// Check the updated result
	res, err := suite.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), res.NumberOfTrades)

	// Pseudo execute 3 more trades
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	res, err = suite.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(4).Uint64(), res.NumberOfTrades.Uint64())
}

// TestGetProtoRevProfitsByDenom tests the query for profits by denom
func (suite *KeeperTestSuite) TestGetProtoRevProfitsByDenom() {
	req := &types.QueryGetProtoRevProfitsByDenomRequest{
		Denom: types.OsmosisDenomination,
	}
	_, err := suite.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().Error(err)

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))

	suite.Require().NoError(err)
	suite.Commit()

	res, err := suite.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(10000), res.Profit.Amount)
}

// TestGetProtoRevAllProfits tests the query for all profits
func (suite *KeeperTestSuite) TestGetProtoRevAllProfits() {
	req := &types.QueryGetProtoRevAllProfitsRequest{}
	res, err := suite.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(res.Profits))

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(9000))
	suite.Require().NoError(err)
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{}, types.AtomDenomination, sdk.NewInt(3000))
	suite.Require().NoError(err)

	res, err = suite.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	atom := sdk.NewCoin(types.AtomDenomination, sdk.NewInt(3000))
	osmo := sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(9000))
	suite.Require().Equal([]*sdk.Coin{&atom, &osmo}, res.Profits)
}

// TestGetProtoRevStatisticsByRoute tests the query for statistics by route
func (suite *KeeperTestSuite) TestGetProtoRevStatisticsByRoute() {
	// Request with no trades should return an error
	req := &types.QueryGetProtoRevStatisticsByRouteRequest{
		Route: []uint64{1, 2, 3},
	}

	res, err := suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.AtomDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics.NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.AtomDenomination, Amount: sdk.NewInt(10000)}}, res.Statistics.Profits)

	// Pseudo execute another trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.AtomDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics.NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.AtomDenomination, Amount: sdk.NewInt(90000)}}, res.Statistics.Profits)

	// Pseudo execute another trade in a different denom
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	suite.Require().Equal(sdk.NewInt(3), res.Statistics.NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.AtomDenomination, Amount: sdk.NewInt(90000)}, {Denom: types.OsmosisDenomination, Amount: sdk.NewInt(80000)}}, res.Statistics.Profits)
}

// TestGetProtoRevAllRouteStatistics tests the query for all route statistics
func (suite *KeeperTestSuite) TestGetProtoRevAllRouteStatistics() {
	req := &types.QueryGetProtoRevAllRouteStatisticsRequest{}

	res, err := suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[0].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(10000)}}, res.Statistics[0].Profits)

	// Pseudo execute another trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics[0].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(90000)}}, res.Statistics[0].Profits)

	// Pseudo execute another trade on a different route
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 4}}, types.OsmosisDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics[0].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(90000)}}, res.Statistics[0].Profits)
	suite.Require().Equal([]uint64{1, 2, 4}, res.Statistics[1].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[1].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(80000)}}, res.Statistics[1].Profits)

	// Pseudo execute another trade on a different route and denom
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, swaproutertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 5}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 4}}, types.AtomDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics[0].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(90000)}}, res.Statistics[0].Profits)
	suite.Require().Equal([]uint64{1, 2, 4}, res.Statistics[1].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[1].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.OsmosisDenomination, Amount: sdk.NewInt(80000)}}, res.Statistics[1].Profits)
	suite.Require().Equal([]uint64{5, 2, 4}, res.Statistics[2].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[2].NumberOfTrades)
	suite.Require().Equal([]*sdk.Coin{{Denom: types.AtomDenomination, Amount: sdk.NewInt(80000)}}, res.Statistics[2].Profits)
}

// TestGetProtoRevTokenPairArbRoutes tests the query to retrieve all token pair arb routes
func (suite *KeeperTestSuite) TestGetProtoRevTokenPairArbRoutes() {
	req := &types.QueryGetProtoRevTokenPairArbRoutesRequest{}
	res, err := suite.queryClient.GetProtoRevTokenPairArbRoutes(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(len(suite.tokenPairArbRoutes), len(res.Routes))

	for _, route := range res.Routes {
		suite.Require().Contains(suite.tokenPairArbRoutes, route)
	}
}
