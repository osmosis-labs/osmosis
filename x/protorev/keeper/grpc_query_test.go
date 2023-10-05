package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/v19/x/protorev/types"
)

// TestParams tests the query for params
func (s *KeeperTestSuite) TestParams() {
	ctx := sdk.WrapSDKContext(s.Ctx)
	expectedParams := s.App.ProtoRevKeeper.GetParams(s.Ctx)

	res, err := s.queryClient.Params(ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(expectedParams, res.Params)
}

// TestGetProtoRevNumberOfTrades tests the query for number of trades
func (s *KeeperTestSuite) TestGetProtoRevNumberOfTrades() {
	// Initially should throw an error
	_, err := s.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(s.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	s.Require().Error(err)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	// Check the updated result
	res, err := s.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(s.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), res.NumberOfTrades)

	// Pseudo execute 3 more trades
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	res, err = s.queryClient.GetProtoRevNumberOfTrades(sdk.WrapSDKContext(s.Ctx), &types.QueryGetProtoRevNumberOfTradesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(4).Uint64(), res.NumberOfTrades.Uint64())
}

// TestGetProtoRevProfitsByDenom tests the query for profits by denom
func (s *KeeperTestSuite) TestGetProtoRevProfitsByDenom() {
	req := &types.QueryGetProtoRevProfitsByDenomRequest{
		Denom: types.OsmosisDenomination,
	}
	_, err := s.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))

	s.Require().NoError(err)
	s.Commit()

	res, err := s.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000), res.Profit.Amount)

	// Pseudo execute a trade in a different denom
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", osmomath.NewInt(10000))

	s.Require().NoError(err)
	s.Commit()

	_, err = s.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	req = &types.QueryGetProtoRevProfitsByDenomRequest{
		Denom: "Atom",
	}
	res, err = s.queryClient.GetProtoRevProfitsByDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000), res.Profit.Amount)
	s.Require().Equal("Atom", res.Profit.Denom)
}

// TestGetProtoRevAllProfits tests the query for all profits
func (s *KeeperTestSuite) TestGetProtoRevAllProfits() {
	req := &types.QueryGetProtoRevAllProfitsRequest{}
	res, err := s.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(0, len(res.Profits))

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(9000))
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", osmomath.NewInt(3000))
	s.Require().NoError(err)

	res, err = s.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	atom := sdk.NewCoin("Atom", osmomath.NewInt(3000))
	osmo := sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(9000))
	s.Require().Contains(res.Profits, atom)
	s.Require().Contains(res.Profits, osmo)

	// Pseudo execute more trades
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", osmomath.NewInt(10000))
	s.Require().NoError(err)

	res, err = s.queryClient.GetProtoRevAllProfits(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	atom = sdk.NewCoin("Atom", osmomath.NewInt(13000))
	osmo = sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(19000))
	s.Require().Contains(res.Profits, atom)
	s.Require().Contains(res.Profits, osmo)
}

// TestGetProtoRevStatisticsByRoute tests the query for statistics by route
func (s *KeeperTestSuite) TestGetProtoRevStatisticsByRoute() {
	// Request with no trades should return an error
	req := &types.QueryGetProtoRevStatisticsByRouteRequest{
		Route: []uint64{1, 2, 3},
	}

	res, err := s.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, "Atom", osmomath.NewInt(10000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	s.Require().Equal(osmomath.OneInt(), res.Statistics.NumberOfTrades)
	coin := sdk.NewCoin("Atom", osmomath.NewInt(10000))
	s.Require().Contains(res.Statistics.Profits, coin)

	// Pseudo execute another trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, "Atom", osmomath.NewInt(80000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	s.Require().Equal(osmomath.NewInt(2), res.Statistics.NumberOfTrades)
	coin = sdk.NewCoin("Atom", osmomath.NewInt(90000))
	s.Require().Contains(res.Statistics.Profits, coin)

	// Pseudo execute another trade in a different denom (might happen in multidenom pools > 2 denoms)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, osmomath.NewInt(80000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevStatisticsByRoute(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	s.Require().Equal(osmomath.NewInt(3), res.Statistics.NumberOfTrades)
	atomCoin := sdk.NewCoin("Atom", osmomath.NewInt(90000))
	osmoCoin := sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(80000))
	s.Require().Contains(res.Statistics.Profits, atomCoin)
	s.Require().Contains(res.Statistics.Profits, osmoCoin)
}

// TestGetProtoRevAllRouteStatistics tests the query for all route statistics
func (s *KeeperTestSuite) TestGetProtoRevAllRouteStatistics() {
	req := &types.QueryGetProtoRevAllRouteStatisticsRequest{}

	res, err := s.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(1, len(res.Statistics))
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	s.Require().Equal(osmomath.OneInt(), res.Statistics[0].NumberOfTrades)
	osmoCoin := sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	// Pseudo execute another trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, osmomath.NewInt(80000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(1, len(res.Statistics))
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	s.Require().Equal(osmomath.NewInt(2), res.Statistics[0].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(90000))
	s.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	// Pseudo execute another trade on a different route
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 4}}, types.OsmosisDenomination, osmomath.NewInt(70000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(2, len(res.Statistics))
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	s.Require().Equal(osmomath.NewInt(2), res.Statistics[0].NumberOfTrades)
	s.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	s.Require().Equal([]uint64{1, 2, 4}, res.Statistics[1].Route)
	s.Require().Equal(osmomath.OneInt(), res.Statistics[1].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(70000))
	s.Require().Contains(res.Statistics[1].Profits, osmoCoin)

	// Pseudo execute another trade on a different route and denom
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 5}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 4}}, "Atom", osmomath.NewInt(80000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(3, len(res.Statistics))
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics[0].Route)
	s.Require().Equal(osmomath.NewInt(2), res.Statistics[0].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(90000))
	s.Require().Contains(res.Statistics[0].Profits, osmoCoin)

	s.Require().Equal([]uint64{1, 2, 4}, res.Statistics[1].Route)
	s.Require().Equal(osmomath.OneInt(), res.Statistics[1].NumberOfTrades)
	osmoCoin = sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(70000))
	s.Require().Contains(res.Statistics[1].Profits, osmoCoin)

	s.Require().Equal([]uint64{5, 2, 4}, res.Statistics[2].Route)
	s.Require().Equal(osmomath.OneInt(), res.Statistics[2].NumberOfTrades)
	atomCoin := sdk.NewCoin("Atom", osmomath.NewInt(80000))
	s.Require().Contains(res.Statistics[2].Profits, atomCoin)
}

// TestGetProtoRevTokenPairArbRoutes tests the query to retrieve all token pair arb routes
func (s *KeeperTestSuite) TestGetProtoRevTokenPairArbRoutes() {
	req := &types.QueryGetProtoRevTokenPairArbRoutesRequest{}
	res, err := s.queryClient.GetProtoRevTokenPairArbRoutes(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(len(s.tokenPairArbRoutes), len(res.Routes))

	for _, route := range res.Routes {
		s.Require().Contains(s.tokenPairArbRoutes, route)
	}
}

// TestGetProtoRevAdminAccount tests the query to retrieve the admin account
func (s *KeeperTestSuite) TestGetProtoRevAdminAccount() {
	req := &types.QueryGetProtoRevAdminAccountRequest{}
	res, err := s.queryClient.GetProtoRevAdminAccount(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(s.adminAccount.String(), res.AdminAccount)
}

// TestGetProtoRevDeveloperAccount tests the query to retrieve the developer account
func (s *KeeperTestSuite) TestGetProtoRevDeveloperAccount() {
	// By default it should be empty
	req := &types.QueryGetProtoRevDeveloperAccountRequest{}
	res, err := s.queryClient.GetProtoRevDeveloperAccount(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Set the developer account
	developerAccount := apptesting.CreateRandomAccounts(1)[0]
	s.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, developerAccount)

	// Verify the developer account
	res, err = s.queryClient.GetProtoRevDeveloperAccount(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(developerAccount.String(), res.DeveloperAccount)
}

// TestGetProtoRevInfoByPoolType tests the query to retrieve the pool info
func (s *KeeperTestSuite) TestGetProtoRevInfoByPoolType() {
	// Set the pool weights
	poolInfo := types.InfoByPoolType{
		Stable:       types.StablePoolInfo{Weight: 1},
		Balancer:     types.BalancerPoolInfo{Weight: 1},
		Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
		Cosmwasm: types.CosmwasmPoolInfo{WeightMaps: []types.WeightMap{
			{ContractAddress: "test", Weight: 1},
		}},
	}
	s.App.AppKeepers.ProtoRevKeeper.SetInfoByPoolType(s.Ctx, poolInfo)

	req := &types.QueryGetProtoRevInfoByPoolTypeRequest{}
	res, err := s.queryClient.GetProtoRevInfoByPoolType(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(poolInfo, res.InfoByPoolType)
}

// TestGetProtoRevMaxPoolPointsPerTx tests the query to retrieve the max pool points per tx
func (s *KeeperTestSuite) TestGetProtoRevMaxPoolPointsPerTx() {
	// Set the max pool points per tx
	maxPoolPointsPerTx := types.MaxPoolPointsPerTx - 1
	err := s.App.AppKeepers.ProtoRevKeeper.SetMaxPointsPerTx(s.Ctx, maxPoolPointsPerTx)
	s.Require().NoError(err)

	req := &types.QueryGetProtoRevMaxPoolPointsPerTxRequest{}
	res, err := s.queryClient.GetProtoRevMaxPoolPointsPerTx(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(maxPoolPointsPerTx, res.MaxPoolPointsPerTx)
}

// TestGetProtoRevMaxPoolPointsPerBlock tests the query to retrieve the max pool points per block
func (s *KeeperTestSuite) TestGetProtoRevMaxPoolPointsPerBlock() {
	// Set the max pool points per block
	maxPoolPointsPerBlock := types.MaxPoolPointsPerBlock - 1
	err := s.App.AppKeepers.ProtoRevKeeper.SetMaxPointsPerBlock(s.Ctx, maxPoolPointsPerBlock)
	s.Require().NoError(err)

	req := &types.QueryGetProtoRevMaxPoolPointsPerBlockRequest{}
	res, err := s.queryClient.GetProtoRevMaxPoolPointsPerBlock(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(maxPoolPointsPerBlock, res.MaxPoolPointsPerBlock)
}

// TestGetProtoRevBaseDenoms tests the query to retrieve the base denoms
func (s *KeeperTestSuite) TestGetProtoRevBaseDenoms() {
	// base denoms already set in setup
	baseDenoms, err := s.App.AppKeepers.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)

	req := &types.QueryGetProtoRevBaseDenomsRequest{}
	res, err := s.queryClient.GetProtoRevBaseDenoms(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(baseDenoms, res.BaseDenoms)
}

// TestGetProtoRevEnabled tests the query to retrieve the enabled status of protorev
func (s *KeeperTestSuite) TestGetProtoRevEnabledQuery() {
	// Set the enabled status
	enabled := false
	s.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(s.Ctx, enabled)

	req := &types.QueryGetProtoRevEnabledRequest{}
	res, err := s.queryClient.GetProtoRevEnabled(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(enabled, res.Enabled)

	// Set the enabled status
	enabled = true
	s.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(s.Ctx, enabled)

	res, err = s.queryClient.GetProtoRevEnabled(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(enabled, res.Enabled)
}

// TestGetProtoRevPool tests the query for getting the highest liquidity pool stored
func (s *KeeperTestSuite) TestGetProtoRevPool() {
	// Request without setting pool for the base denom and other denom should return an error
	req := &types.QueryGetProtoRevPoolRequest{
		BaseDenom:  "uosmo",
		OtherDenom: "atom",
	}
	res, err := s.queryClient.GetProtoRevPool(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Request for a pool that is stored should return the pool id
	// The pool is set at startup for the test suite
	req = &types.QueryGetProtoRevPoolRequest{
		BaseDenom:  "Atom",
		OtherDenom: "akash",
	}
	res, err = s.queryClient.GetProtoRevPool(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err)
	s.Require().Equal(res.PoolId, uint64(1))
}
