package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// TestParams tests the query for params
func (suite *KeeperTestSuite) TestParams() {
	ctx := sdk.WrapSDKContext(suite.Ctx)
	expectedParams := suite.App.ProtoRevKeeper.GetParams(suite.Ctx)

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
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	// Check the updated result
	res, err := suite.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(suite.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(1), res.NumberOfTrades)

	// Pseudo execute 3 more trades
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
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
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))

	suite.Require().NoError(err)
	suite.Commit()

	res, err := suite.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(10000), res.Profit.Amount)

	// Pseudo execute a trade in a different denom
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", sdk.NewInt(10000))

	suite.Require().NoError(err)
	suite.Commit()

	_, err = suite.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	req = &types.QueryGetProtoRevProfitsByDenomRequest{
		Denom: "Atom",
	}
	res, err = suite.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt(10000), res.Profit.Amount)
	suite.Require().Equal("Atom", res.Profit.Denom)
}

// TestGetProtoRevAllProfits tests the query for all profits
func (suite *KeeperTestSuite) TestGetProtoRevAllProfits() {
	req := &types.QueryGetProtoRevAllProfitsRequest{}
	res, err := suite.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(res.Profits))

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(9000))
	suite.Require().NoError(err)
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", sdk.NewInt(3000))
	suite.Require().NoError(err)

	res, err = suite.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	atom := sdk.NewCoin("Atom", sdk.NewInt(3000))
	osmo := sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(9000))
	suite.Require().Contains(res.Profits, atom)
	suite.Require().Contains(res.Profits, osmo)

	// Pseudo execute more trades
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", sdk.NewInt(10000))
	suite.Require().NoError(err)

	res, err = suite.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	atom = sdk.NewCoin("Atom", sdk.NewInt(13000))
	osmo = sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(19000))
	suite.Require().Contains(res.Profits, atom)
	suite.Require().Contains(res.Profits, osmo)
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
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, "Atom", sdk.NewInt(10000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics.NumberOfTrades)
	coin := sdk.NewCoin("Atom", sdk.NewInt(10000))
	suite.Require().Contains(res.Statistics.Profits, coin)

	// Pseudo execute another trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, "Atom", sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics.NumberOfTrades)
	coin = sdk.NewCoin("Atom", sdk.NewInt(90000))
	suite.Require().Contains(res.Statistics.Profits, coin)

	// Pseudo execute another trade in a different denom (might happen in multidenom pools > 2 denoms)
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	suite.Require().Equal(sdk.NewInt(3), res.Statistics.NumberOfTrades)
	atomCoin := sdk.NewCoin("Atom", sdk.NewInt(90000))
	osmoCoin := sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(80000))
	suite.Require().Contains(res.Statistics.Profits, atomCoin)
	suite.Require().Contains(res.Statistics.Profits, osmoCoin)
}

// TestGetProtoRevAllRouteStatistics tests the query for all route statistics
func (suite *KeeperTestSuite) TestGetProtoRevAllRouteStatistics() {
	req := &types.QueryGetProtoRevAllRouteStatisticsRequest{}

	res, err := suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	// Pseudo execute a trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[0].NumberOfTrades)
	osmoCoin := sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))
	suite.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	// Pseudo execute another trade
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics[0].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(90000))
	suite.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	// Pseudo execute another trade on a different route
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 4}}, types.OsmosisDenomination, sdk.NewInt(70000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics[0].NumberOfTrades)
	suite.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	suite.Require().Equal([]uint64{1, 2, 4}, res.Statistics[1].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[1].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(70000))
	suite.Require().Contains(res.Statistics[1].Profits, osmoCoin)

	// Pseudo execute another trade on a different route and denom
	err = suite.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(suite.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 5}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 4}}, "Atom", sdk.NewInt(80000))
	suite.Require().NoError(err)

	// Verify statistics
	res, err = suite.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(3, len(res.Statistics))
	suite.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	suite.Require().Equal(sdk.NewInt(2), res.Statistics[0].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(90000))
	suite.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	suite.Require().Equal([]uint64{1, 2, 4}, res.Statistics[1].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[1].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(70000))
	suite.Require().Contains(res.Statistics[1].Profits, osmoCoin)

	suite.Require().Equal([]uint64{5, 2, 4}, res.Statistics[2].Route)
	suite.Require().Equal(sdk.OneInt(), res.Statistics[2].NumberOfTrades)
	atomCoin := sdk.NewCoin("Atom", sdk.NewInt(80000))
	suite.Require().Contains(res.Statistics[2].Profits, atomCoin)
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

// TestGetProtoRevAdminAccount tests the query to retrieve the admin account
func (suite *KeeperTestSuite) TestGetProtoRevAdminAccount() {
	req := &types.QueryGetProtoRevAdminAccountRequest{}
	res, err := suite.queryClient.GetProtoRevAdminAccount(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.adminAccount.String(), res.AdminAccount)
}

// TestGetProtoRevDeveloperAccount tests the query to retrieve the developer account
func (suite *KeeperTestSuite) TestGetProtoRevDeveloperAccount() {
	// By default it should be empty
	req := &types.QueryGetProtoRevDeveloperAccountRequest{}
	res, err := suite.queryClient.GetProtoRevDeveloperAccount(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	// Set the developer account
	developerAccount := apptesting.CreateRandomAccounts(1)[0]
	suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, developerAccount)

	// Verify the developer account
	res, err = suite.queryClient.GetProtoRevDeveloperAccount(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(developerAccount.String(), res.DeveloperAccount)
}

// TestGetProtoRevPoolWeights tests the query to retrieve the pool weights
func (suite *KeeperTestSuite) TestGetProtoRevPoolWeights() {
	// Set the pool weights
	poolWeights := types.PoolWeights{
		StableWeight:       5,
		BalancerWeight:     1,
		ConcentratedWeight: 3,
	}
	suite.App.AppKeepers.ProtoRevKeeper.SetPoolWeights(suite.Ctx, poolWeights)

	req := &types.QueryGetProtoRevPoolWeightsRequest{}
	res, err := suite.queryClient.GetProtoRevPoolWeights(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(poolWeights, res.PoolWeights)
}

// TestGetProtoRevMaxPoolPointsPerTx tests the query to retrieve the max pool points per tx
func (suite *KeeperTestSuite) TestGetProtoRevMaxPoolPointsPerTx() {
	// Set the max pool points per tx
	maxPoolPointsPerTx := types.MaxPoolPointsPerTx - 1
	err := suite.App.AppKeepers.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, maxPoolPointsPerTx)
	suite.Require().NoError(err)

	req := &types.QueryGetProtoRevMaxPoolPointsPerTxRequest{}
	res, err := suite.queryClient.GetProtoRevMaxPoolPointsPerTx(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(maxPoolPointsPerTx, res.MaxPoolPointsPerTx)
}

// TestGetProtoRevMaxPoolPointsPerBlock tests the query to retrieve the max pool points per block
func (suite *KeeperTestSuite) TestGetProtoRevMaxPoolPointsPerBlock() {
	// Set the max pool points per block
	maxPoolPointsPerBlock := types.MaxPoolPointsPerBlock - 1
	err := suite.App.AppKeepers.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, maxPoolPointsPerBlock)
	suite.Require().NoError(err)

	req := &types.QueryGetProtoRevMaxPoolPointsPerBlockRequest{}
	res, err := suite.queryClient.GetProtoRevMaxPoolPointsPerBlock(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(maxPoolPointsPerBlock, res.MaxPoolPointsPerBlock)
}

// TestGetProtoRevBaseDenoms tests the query to retrieve the base denoms
func (suite *KeeperTestSuite) TestGetProtoRevBaseDenoms() {
	// base denoms already set in setup
	baseDenoms, err := suite.App.AppKeepers.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
	suite.Require().NoError(err)

	req := &types.QueryGetProtoRevBaseDenomsRequest{}
	res, err := suite.queryClient.GetProtoRevBaseDenoms(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(baseDenoms, res.BaseDenoms)
}

// TestGetProtoRevEnabled tests the query to retrieve the enabled status of protorev
func (suite *KeeperTestSuite) TestGetProtoRevEnabledQuery() {
	// Set the enabled status
	enabled := false
	suite.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(suite.Ctx, enabled)

	req := &types.QueryGetProtoRevEnabledRequest{}
	res, err := suite.queryClient.GetProtoRevEnabled(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(enabled, res.Enabled)

	// Set the enabled status
	enabled = true
	suite.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(suite.Ctx, enabled)

	res, err = suite.queryClient.GetProtoRevEnabled(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(enabled, res.Enabled)
}

// TestGetProtoRevPool tests the query for getting the highest liquidity pool stored
func (suite *KeeperTestSuite) TestGetProtoRevPool() {
	// Request without setting pool for the base denom and other denom should return an error
	req := &types.QueryGetProtoRevPoolRequest{
		BaseDenom:  "uosmo",
		OtherDenom: "atom",
	}
	res, err := suite.queryClient.GetProtoRevPool(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	// Request for a pool that is stored should return the pool id
	// The pool is set at startup for the test suite
	req = &types.QueryGetProtoRevPoolRequest{
		BaseDenom:  "Atom",
		OtherDenom: "akash",
	}
	res, err = suite.queryClient.GetProtoRevPool(sdk.WrapSDKContext(suite.Ctx), req)
	suite.Require().NoError(err)
	suite.Require().Equal(res.PoolId, uint64(1))
}
