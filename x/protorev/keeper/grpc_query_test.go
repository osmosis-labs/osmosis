package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// TestParams tests the query for params
func (s *KeeperTestSuite) TestParams() {
	ctx := s.Ctx
	expectedParams := s.App.ProtoRevKeeper.GetParams(s.Ctx)

	res, err := s.queryClient.Params(ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(expectedParams, res.Params)
}

// TestGetProtoRevNumberOfTrades tests the query for number of trades
func (s *KeeperTestSuite) TestGetProtoRevNumberOfTrades() {
	s.SetupPoolsTest()
	// Initially should throw an error
	_, err := s.queryClient.GetProtoRevNumberOfTrades(s.Ctx, &types.QueryGetProtoRevNumberOfTradesRequest{})
	s.Require().Error(err)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	// Check the updated result
	res, err := s.queryClient.GetProtoRevNumberOfTrades(s.Ctx, &types.QueryGetProtoRevNumberOfTradesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(1), res.NumberOfTrades)

	// Pseudo execute 3 more trades
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	res, err = s.queryClient.GetProtoRevNumberOfTrades(s.Ctx, &types.QueryGetProtoRevNumberOfTradesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(4).Uint64(), res.NumberOfTrades.Uint64())
}

// TestGetProtoRevProfitsByDenom tests the query for profits by denom
func (s *KeeperTestSuite) TestGetProtoRevProfitsByDenom() {
	s.SetupPoolsTest()
	req := &types.QueryGetProtoRevProfitsByDenomRequest{
		Denom: types.OsmosisDenomination,
	}
	_, err := s.queryClient.GetProtoRevProfitsByDenom(s.Ctx, req)
	s.Require().Error(err)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(10000))

	s.Require().NoError(err)
	s.Commit()

	res, err := s.queryClient.GetProtoRevProfitsByDenom(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000), res.Profit.Amount)

	// Pseudo execute a trade in a different denom
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", osmomath.NewInt(10000))

	s.Require().NoError(err)
	s.Commit()

	_, err = s.queryClient.GetProtoRevProfitsByDenom(s.Ctx, req)
	s.Require().NoError(err)
	req = &types.QueryGetProtoRevProfitsByDenomRequest{
		Denom: "Atom",
	}
	res, err = s.queryClient.GetProtoRevProfitsByDenom(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewInt(10000), res.Profit.Amount)
	s.Require().Equal("Atom", res.Profit.Denom)
}

// TestGetProtoRevAllProfits tests the query for all profits
func (s *KeeperTestSuite) TestGetProtoRevAllProfits() {
	s.SetupPoolsTest()
	req := &types.QueryGetProtoRevAllProfitsRequest{}
	res, err := s.queryClient.GetProtoRevAllProfits(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(0, len(res.Profits))

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, types.OsmosisDenomination, osmomath.NewInt(9000))
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, "Atom", osmomath.NewInt(3000))
	s.Require().NoError(err)

	res, err = s.queryClient.GetProtoRevAllProfits(s.Ctx, req)
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

	res, err = s.queryClient.GetProtoRevAllProfits(s.Ctx, req)
	s.Require().NoError(err)
	atom = sdk.NewCoin("Atom", osmomath.NewInt(13000))
	osmo = sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(19000))
	s.Require().Contains(res.Profits, atom)
	s.Require().Contains(res.Profits, osmo)
}

// TestGetProtoRevStatisticsByRoute tests the query for statistics by route
func (s *KeeperTestSuite) TestGetProtoRevStatisticsByRoute() {
	s.SetupPoolsTest()
	// Request with no trades should return an error
	req := &types.QueryGetProtoRevStatisticsByRouteRequest{
		Route: []uint64{1, 2, 3},
	}

	res, err := s.queryClient.GetProtoRevStatisticsByRoute(s.Ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, "Atom", osmomath.NewInt(10000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevStatisticsByRoute(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	s.Require().Equal(osmomath.OneInt(), res.Statistics.NumberOfTrades)
	coin := sdk.NewCoin("Atom", osmomath.NewInt(10000))
	s.Require().Contains(res.Statistics.Profits, coin)

	// Pseudo execute another trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, "Atom", osmomath.NewInt(80000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevStatisticsByRoute(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal([]uint64{1, 2, 3}, res.Statistics.Route)
	s.Require().Equal(osmomath.NewInt(2), res.Statistics.NumberOfTrades)
	coin = sdk.NewCoin("Atom", osmomath.NewInt(90000))
	s.Require().Contains(res.Statistics.Profits, coin)

	// Pseudo execute another trade in a different denom (might happen in multidenom pools > 2 denoms)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, osmomath.NewInt(80000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevStatisticsByRoute(s.Ctx, req)
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
	s.SetupPoolsTest()
	req := &types.QueryGetProtoRevAllRouteStatisticsRequest{}

	res, err := s.queryClient.GetProtoRevAllRouteStatistics(s.Ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Pseudo execute a trade
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{{TokenOutDenom: "", PoolId: 1}, {TokenOutDenom: "", PoolId: 2}, {TokenOutDenom: "", PoolId: 3}}, types.OsmosisDenomination, osmomath.NewInt(10000))
	s.Require().NoError(err)

	// Verify statistics
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(s.Ctx, req)
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
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(s.Ctx, req)
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
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(s.Ctx, req)
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
	res, err = s.queryClient.GetProtoRevAllRouteStatistics(s.Ctx, req)
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
	s.SetupPoolsTest()
	req := &types.QueryGetProtoRevTokenPairArbRoutesRequest{}
	res, err := s.queryClient.GetProtoRevTokenPairArbRoutes(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(len(s.tokenPairArbRoutes), len(res.Routes))

	for _, route := range res.Routes {
		s.Require().Contains(s.tokenPairArbRoutes, route)
	}
}

// TestGetProtoRevAdminAccount tests the query to retrieve the admin account
func (s *KeeperTestSuite) TestGetProtoRevAdminAccount() {
	req := &types.QueryGetProtoRevAdminAccountRequest{}
	res, err := s.queryClient.GetProtoRevAdminAccount(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(s.adminAccount.String(), res.AdminAccount)
}

// TestGetProtoRevDeveloperAccount tests the query to retrieve the developer account
func (s *KeeperTestSuite) TestGetProtoRevDeveloperAccount() {
	// By default it should be empty
	req := &types.QueryGetProtoRevDeveloperAccountRequest{}
	res, err := s.queryClient.GetProtoRevDeveloperAccount(s.Ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Set the developer account
	developerAccount := apptesting.CreateRandomAccounts(1)[0]
	s.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, developerAccount)

	// Verify the developer account
	res, err = s.queryClient.GetProtoRevDeveloperAccount(s.Ctx, req)
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
	res, err := s.queryClient.GetProtoRevInfoByPoolType(s.Ctx, req)
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
	res, err := s.queryClient.GetProtoRevMaxPoolPointsPerTx(s.Ctx, req)
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
	res, err := s.queryClient.GetProtoRevMaxPoolPointsPerBlock(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(maxPoolPointsPerBlock, res.MaxPoolPointsPerBlock)
}

// TestGetProtoRevBaseDenoms tests the query to retrieve the base denoms
func (s *KeeperTestSuite) TestGetProtoRevBaseDenoms() {
	// base denoms already set in setup
	baseDenoms, err := s.App.AppKeepers.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)

	req := &types.QueryGetProtoRevBaseDenomsRequest{}
	res, err := s.queryClient.GetProtoRevBaseDenoms(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(baseDenoms, res.BaseDenoms)
}

// TestGetProtoRevEnabled tests the query to retrieve the enabled status of protorev
func (s *KeeperTestSuite) TestGetProtoRevEnabledQuery() {
	// Set the enabled status
	enabled := false
	s.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(s.Ctx, enabled)

	req := &types.QueryGetProtoRevEnabledRequest{}
	res, err := s.queryClient.GetProtoRevEnabled(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(enabled, res.Enabled)

	// Set the enabled status
	enabled = true
	s.App.AppKeepers.ProtoRevKeeper.SetProtoRevEnabled(s.Ctx, enabled)

	res, err = s.queryClient.GetProtoRevEnabled(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(enabled, res.Enabled)
}

// TestGetProtoRevPool tests the query for getting the highest liquidity pool stored
func (s *KeeperTestSuite) TestGetProtoRevPool() {
	s.SetupPoolsTest()
	// Request without setting pool for the base denom and other denom should return an error
	req := &types.QueryGetProtoRevPoolRequest{
		BaseDenom:  appparams.BaseCoinUnit,
		OtherDenom: "atom",
	}
	res, err := s.queryClient.GetProtoRevPool(s.Ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	// Request for a pool that is stored should return the pool id
	// The pool is set at startup for the test suite
	req = &types.QueryGetProtoRevPoolRequest{
		BaseDenom:  "Atom",
		OtherDenom: "akash",
	}
	res, err = s.queryClient.GetProtoRevPool(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(res.PoolId, uint64(1))
}

// TestGetAllProtocolRevenue tests the query for all protocol revenue profits
func (s *KeeperTestSuite) TestGetAllProtocolRevenueGRPCQuery() {
	baseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)
	communityPoolDenom := "Akash"
	atom := "Atom"
	ion := "uion"

	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.MustNewDecFromStr("0.02")
	poolManagerParams.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo = communityPoolDenom
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	req := &types.QueryGetAllProtocolRevenueRequest{}
	res, err := s.queryClient.GetAllProtocolRevenue(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Empty(res.AllProtocolRevenue)

	// Swap on a pool to charge taker fee
	swapInCoin := sdk.NewCoin(atom, osmomath.NewInt(1000))
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(atom, osmomath.NewInt(10000))))
	atomCommPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoins(sdk.NewCoin(atom, osmomath.NewInt(10000)), sdk.NewCoin(communityPoolDenom, osmomath.NewInt(10000)))...)
	atomBaseDenomPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(atom, baseDenom)

	// Store cache context prior to swap so we can use it to calculate how much outToken we should expect after the epoch hook is called and taker fees are swapped.
	cacheCtx, _ := s.Ctx.CacheContext()

	_, _, err = s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], atomCommPoolID, swapInCoin, communityPoolDenom, osmomath.ZeroInt())
	s.Require().NoError(err)
	expectedTakerFeeFromInput := swapInCoin.Amount.ToLegacyDec().Mul(poolManagerParams.TakerFeeParams.DefaultTakerFee)
	expectedTakerFeeToCommunityPoolAmt := expectedTakerFeeFromInput.Mul(poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool).TruncateInt()
	expectedTakerFeeToStakersAmt := expectedTakerFeeFromInput.Sub(expectedTakerFeeToCommunityPoolAmt.ToLegacyDec()).TruncateInt()
	expectedTakerFeeToStakers := sdk.NewCoins(sdk.NewCoin(atom, expectedTakerFeeToStakersAmt))
	expectedTakerFeeToCommunityPool := sdk.NewCoins(sdk.NewCoin(atom, expectedTakerFeeToCommunityPoolAmt))

	// We swap taker fees to stakers to the base denom
	baseDenomAmt, err := s.App.PoolManagerKeeper.SwapExactAmountInNoTakerFee(cacheCtx, s.TestAccs[0], atomBaseDenomPool.GetId(), expectedTakerFeeToStakers[0], baseDenom, osmomath.ZeroInt())
	s.Require().NoError(err)
	expectedTakerFeeToStakers = sdk.NewCoins(sdk.NewCoin(baseDenom, baseDenomAmt))

	// We swap taker fees to community pool that are not whitelisted to the CommunityPoolDenomToSwapNonWhitelistedAssetsTo
	communityPoolDenomAmt, err := s.App.PoolManagerKeeper.SwapExactAmountInNoTakerFee(cacheCtx, s.TestAccs[0], atomCommPoolID, expectedTakerFeeToCommunityPool[0], communityPoolDenom, osmomath.ZeroInt())
	s.Require().NoError(err)
	expectedTakerFeeToCommunityPool = sdk.NewCoins(sdk.NewCoin(communityPoolDenom, communityPoolDenomAmt))

	// Charge txfee of 1000 uion
	txFeeCharged := sdk.NewCoins(sdk.NewCoin(ion, osmomath.NewInt(1000)))
	s.SetupTxFeeAnteHandlerAndChargeFee(s.clientCtx, sdk.NewDecCoins(sdk.NewInt64DecCoin(ion, 1000000)), 0, true, false, txFeeCharged)

	// Pseudo collect cyclic arb profits
	cyclicArbProfits := sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(9000)), sdk.NewCoin(atom, osmomath.NewInt(3000)))
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[0].Denom, cyclicArbProfits[0].Amount)
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[1].Denom, cyclicArbProfits[1].Amount)
	s.Require().NoError(err)

	// Tracker logic is executed has been moved to once per epoch, so we trigger it here.
	distrEpochIdentifier := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier
	err = s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 1)
	s.Require().NoError(err)

	// Check protocol revenue
	res, err = s.queryClient.GetAllProtocolRevenue(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.Coin(cyclicArbProfits), res.AllProtocolRevenue.CyclicArbTracker.CyclicArb)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToStakers), res.AllProtocolRevenue.TakerFeesTracker.TakerFeesToStakers)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToCommunityPool), res.AllProtocolRevenue.TakerFeesTracker.TakerFeesToCommunityPool)

	// A second round of the same thing
	// Swap on a pool to charge taker fee
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(atom, osmomath.NewInt(10000))))
	_, _, err = s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], atomCommPoolID, swapInCoin, communityPoolDenom, osmomath.ZeroInt())
	s.Require().NoError(err)

	// Charge txfee of 1000 uion
	s.SetupTxFeeAnteHandlerAndChargeFee(s.clientCtx, sdk.NewDecCoins(sdk.NewInt64DecCoin(ion, 1000000)), 0, true, false, txFeeCharged)

	// Pseudo collect cyclic arb profits
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[0].Denom, cyclicArbProfits[0].Amount)
	s.Require().NoError(err)
	err = s.App.AppKeepers.ProtoRevKeeper.UpdateStatistics(s.Ctx, poolmanagertypes.SwapAmountInRoutes{}, cyclicArbProfits[1].Denom, cyclicArbProfits[1].Amount)
	s.Require().NoError(err)

	// Tracker logic is executed has been moved to once per epoch, so we trigger it here.
	err = s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, distrEpochIdentifier, 2)
	s.Require().NoError(err)

	// Check protocol revenue
	res, err = s.queryClient.GetAllProtocolRevenue(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().Equal([]sdk.Coin(cyclicArbProfits.Add(cyclicArbProfits...)), res.AllProtocolRevenue.CyclicArbTracker.CyclicArb)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToStakers.Add(expectedTakerFeeToStakers...)), res.AllProtocolRevenue.TakerFeesTracker.TakerFeesToStakers)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToCommunityPool.Add(expectedTakerFeeToCommunityPool...)), res.AllProtocolRevenue.TakerFeesTracker.TakerFeesToCommunityPool)
}
