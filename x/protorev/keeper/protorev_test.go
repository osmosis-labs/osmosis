package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// TestGetTokenPairArbRoutes tests the GetTokenPairArbRoutes function.
func (s *KeeperTestSuite) TestGetTokenPairArbRoutes() {
	s.SetupPoolsTest()
	// Tests that we can properly retrieve all of the routes that were set up
	for _, tokenPair := range s.tokenPairArbRoutes {
		tokenPairArbRoutes, err := s.App.ProtoRevKeeper.GetTokenPairArbRoutes(s.Ctx, tokenPair.TokenIn, tokenPair.TokenOut)

		s.Require().NoError(err)
		s.Require().Equal(tokenPair, tokenPairArbRoutes)
	}

	// Testing to see if we will not find a route that does not exist
	_, err := s.App.ProtoRevKeeper.GetTokenPairArbRoutes(s.Ctx, "osmo", "abc")
	s.Require().Error(err)
}

// TestGetAllTokenPairArbRoutes tests the GetAllTokenPairArbRoutes function.
func (s *KeeperTestSuite) TestGetAllTokenPairArbRoutes() {
	s.SetupPoolsTest()
	// Tests that we can properly retrieve all of the routes that were set up
	tokenPairArbRoutes, err := s.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(s.Ctx)

	s.Require().NoError(err)

	s.Require().Equal(len(s.tokenPairArbRoutes), len(tokenPairArbRoutes))
	for _, tokenPair := range s.tokenPairArbRoutes {
		s.Require().Contains(tokenPairArbRoutes, tokenPair)
	}
}

// TestDeleteAllTokenPairArbRoutes tests the DeleteAllTokenPairArbRoutes function.
func (s *KeeperTestSuite) TestDeleteAllTokenPairArbRoutes() {
	s.SetupPoolsTest()
	// Tests that we can properly retrieve all of the routes that were set up
	tokenPairArbRoutes, err := s.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(s.Ctx)

	s.Require().NoError(err)
	s.Require().Equal(len(s.tokenPairArbRoutes), len(tokenPairArbRoutes))
	for _, tokenPair := range s.tokenPairArbRoutes {
		s.Require().Contains(tokenPairArbRoutes, tokenPair)
	}

	// Delete all routes
	s.App.ProtoRevKeeper.DeleteAllTokenPairArbRoutes(s.Ctx)

	// Test after deletion
	tokenPairArbRoutes, err = s.App.ProtoRevKeeper.GetAllTokenPairArbRoutes(s.Ctx)

	s.Require().NoError(err)
	s.Require().Equal(0, len(tokenPairArbRoutes))
}

// TestGetAllBaseDenoms tests the GetAllBaseDenoms, SetBaseDenoms, and DeleteBaseDenoms functions.
func (s *KeeperTestSuite) TestGetAllBaseDenoms() {
	// Should be initialized on genesis
	baseDenoms, err := s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(3, len(baseDenoms))
	s.Require().Equal(baseDenoms[0].Denom, types.OsmosisDenomination)
	s.Require().Equal(baseDenoms[1].Denom, "Atom")
	s.Require().Equal(baseDenoms[2].Denom, "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7")

	// Should be able to set the base denoms
	err = s.App.ProtoRevKeeper.SetBaseDenoms(s.Ctx, []types.BaseDenom{{Denom: "osmo"}, {Denom: "atom"}, {Denom: "weth"}})
	s.Require().NoError(err)
	baseDenoms, err = s.App.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(3, len(baseDenoms))
	s.Require().Equal(baseDenoms[0].Denom, "osmo")
	s.Require().Equal(baseDenoms[1].Denom, "atom")
	s.Require().Equal(baseDenoms[2].Denom, "weth")
}

// TestGetPoolForDenomPair tests the GetPoolForDenomPair, SetPoolForDenomPair, and DeleteAllPoolsForBaseDenom functions.
func (s *KeeperTestSuite) TestGetPoolForDenomPair() {
	// Should be able to set a pool for a denom pair
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "Atom", types.OsmosisDenomination, 1000)
	pool, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, "Atom", types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1000), pool)

	// Should be able to add another pool for a denom pair
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, "Atom", "weth", 2000)
	pool, err = s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, "Atom", "weth")
	s.Require().NoError(err)
	s.Require().Equal(uint64(2000), pool)

	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, types.OsmosisDenomination, "Atom", 3000)
	pool, err = s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, types.OsmosisDenomination, "Atom")
	s.Require().NoError(err)
	s.Require().Equal(uint64(3000), pool)

	// Should be able to delete all pools for a base denom
	s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, "Atom")
	_, err = s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, "Atom", types.OsmosisDenomination)
	s.Require().Error(err)
	_, err = s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, "Atom", "weth")
	s.Require().Error(err)

	// Other denoms should still exist
	pool, err = s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, types.OsmosisDenomination, "Atom")
	s.Require().NoError(err)
	s.Require().Equal(uint64(3000), pool)
}

// TestGetDaysSinceModuleGenesis tests the GetDaysSinceModuleGenesis and SetDaysSinceModuleGenesis functions.
func (s *KeeperTestSuite) TestGetDaysSinceModuleGenesis() {
	// Should be initialized to 0 on genesis
	daysSinceGenesis, err := s.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), daysSinceGenesis)

	// Should be able to set the days since genesis
	s.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(s.Ctx, 1)
	daysSinceGenesis, err = s.App.ProtoRevKeeper.GetDaysSinceModuleGenesis(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), daysSinceGenesis)
}

// TestGetDeveloperFees tests the GetDeveloperFees, SetDeveloperFees, and GetAllDeveloperFees functions.
func (s *KeeperTestSuite) TestGetDeveloperFees() {
	// Should be initialized to [] on genesis
	fees, err := s.App.ProtoRevKeeper.GetAllDeveloperFees(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(0, len(fees))

	// Should be no osmo fees on genesis
	osmoFees, err := s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, types.OsmosisDenomination)
	s.Require().Error(err)
	s.Require().Equal(sdk.Coin{}, osmoFees)

	// Should be no atom fees on genesis
	atomFees, err := s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, "Atom")
	s.Require().Error(err)
	s.Require().Equal(sdk.Coin{}, atomFees)

	// Should be able to set the fees
	err = s.App.ProtoRevKeeper.SetDeveloperFees(s.Ctx, sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(100)))
	s.Require().NoError(err)
	err = s.App.ProtoRevKeeper.SetDeveloperFees(s.Ctx, sdk.NewCoin("Atom", osmomath.NewInt(100)))
	s.Require().NoError(err)
	err = s.App.ProtoRevKeeper.SetDeveloperFees(s.Ctx, sdk.NewCoin("weth", osmomath.NewInt(100)))
	s.Require().NoError(err)

	// Should be able to get the fees
	osmoFees, err = s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(100)), osmoFees)
	atomFees, err = s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, "Atom")
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin("Atom", osmomath.NewInt(100)), atomFees)
	wethFees, err := s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, "weth")
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin("weth", osmomath.NewInt(100)), wethFees)

	fees, err = s.App.ProtoRevKeeper.GetAllDeveloperFees(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(3, len(fees))
	s.Require().Contains(fees, osmoFees)
	s.Require().Contains(fees, atomFees)
}

// TestDeleteDeveloperFees tests the DeleteDeveloperFees function.
func (s *KeeperTestSuite) TestDeleteDeveloperFees() {
	err := s.App.ProtoRevKeeper.SetDeveloperFees(s.Ctx, sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(100)))
	s.Require().NoError(err)

	// Should be able to get the fees
	osmoFees, err := s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, types.OsmosisDenomination)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(100)), osmoFees)

	// Should be able to delete the fees
	s.App.ProtoRevKeeper.DeleteDeveloperFees(s.Ctx, types.OsmosisDenomination)

	// Should be no osmo fees after deletion
	osmoFees, err = s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, types.OsmosisDenomination)
	s.Require().Error(err)
	s.Require().Equal(sdk.Coin{}, osmoFees)
}

// TestGetProtoRevEnabled tests the GetProtoRevEnabled and SetProtoRevEnabled functions.
func (s *KeeperTestSuite) TestGetProtoRevEnabled() {
	// Should be initialized to true on genesis
	protoRevEnabled := s.App.ProtoRevKeeper.GetProtoRevEnabled(s.Ctx)
	s.Require().Equal(true, protoRevEnabled)

	// Should be able to set the protoRevEnabled
	s.App.ProtoRevKeeper.SetProtoRevEnabled(s.Ctx, false)
	protoRevEnabled = s.App.ProtoRevKeeper.GetProtoRevEnabled(s.Ctx)
	s.Require().Equal(false, protoRevEnabled)
}

// TestGetAdminAccount tests the GetAdminAccount and SetAdminAccount functions.
func (s *KeeperTestSuite) TestGetAdminAccount() {
	// Should be initialized (look at keeper_test.go)
	adminAccount := s.App.ProtoRevKeeper.GetAdminAccount(s.Ctx)
	s.Require().Equal(s.adminAccount, adminAccount)

	// Should be able to set the admin account
	s.App.ProtoRevKeeper.SetAdminAccount(s.Ctx, s.TestAccs[0])
	adminAccount = s.App.ProtoRevKeeper.GetAdminAccount(s.Ctx)
	s.Require().Equal(s.TestAccs[0], adminAccount)
}

// TestGetDeveloperAccount tests the GetDeveloperAccount and SetDeveloperAccount functions.
func (s *KeeperTestSuite) TestGetDeveloperAccount() {
	// Should be null on genesis
	developerAccount, err := s.App.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
	s.Require().Error(err)
	s.Require().Nil(developerAccount)

	// Should be able to set the developer account
	s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, s.TestAccs[0])
	developerAccount, err = s.App.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(s.TestAccs[0], developerAccount)
}

// TestGetMaxPointsPerTx tests the GetMaxPointsPerTx and SetMaxPointsPerTx functions.
func (s *KeeperTestSuite) TestGetMaxPointsPerTx() {
	// Should be initialized on genesis
	maxPoints, err := s.App.ProtoRevKeeper.GetMaxPointsPerTx(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(18), maxPoints)

	// Should be able to set the max points per tx
	err = s.App.ProtoRevKeeper.SetMaxPointsPerTx(s.Ctx, 4)
	s.Require().NoError(err)
	maxPoints, err = s.App.ProtoRevKeeper.GetMaxPointsPerTx(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), maxPoints)

	// Can only be set between 1 and types.MaxPoolPointsPerTx
	err = s.App.ProtoRevKeeper.SetMaxPointsPerTx(s.Ctx, 0)
	s.Require().Error(err)
	err = s.App.ProtoRevKeeper.SetMaxPointsPerTx(s.Ctx, types.MaxPoolPointsPerTx+1)
	s.Require().Error(err)
}

// TestGetPointCountForBlock tests the GetPointCountForBlock, IncrementPointCountForBlock and SetPointCountForBlock functions.
func (s *KeeperTestSuite) TestGetPointCountForBlock() {
	// Should be initialized to 0 on genesis
	pointCount, err := s.App.ProtoRevKeeper.GetPointCountForBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), pointCount)

	// Should be able to set the point count
	s.App.ProtoRevKeeper.SetPointCountForBlock(s.Ctx, 4)
	pointCount, err = s.App.ProtoRevKeeper.GetPointCountForBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), pointCount)

	// Should be able to increment the point count
	err = s.App.ProtoRevKeeper.IncrementPointCountForBlock(s.Ctx, 10)
	s.Require().NoError(err)
	pointCount, err = s.App.ProtoRevKeeper.GetPointCountForBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(14), pointCount)
}

// TestGetLatestBlockHeight tests the GetLatestBlockHeight and SetLatestBlockHeight functions.
func (s *KeeperTestSuite) TestGetLatestBlockHeight() {
	// Should be initialized on genesis
	blockHeight, err := s.App.ProtoRevKeeper.GetLatestBlockHeight(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), blockHeight)

	// Should be able to set the blockHeight
	s.App.ProtoRevKeeper.SetLatestBlockHeight(s.Ctx, 4)
	blockHeight, err = s.App.ProtoRevKeeper.GetLatestBlockHeight(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), blockHeight)
}

// TestGetMaxPointsPerBlock tests the GetMaxPointsPerBlock and SetMaxPointsPerBlock functions.
func (s *KeeperTestSuite) TestGetMaxPointsPerBlock() {
	// Should be initialized on genesis
	maxPoints, err := s.App.ProtoRevKeeper.GetMaxPointsPerBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(100), maxPoints)

	// Should be able to set the max points per block
	err = s.App.ProtoRevKeeper.SetMaxPointsPerBlock(s.Ctx, 4)
	s.Require().NoError(err)
	maxPoints, err = s.App.ProtoRevKeeper.GetMaxPointsPerBlock(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), maxPoints)

	// Can only initialize between 1 and types.MaxPoolPointsPerBlock
	err = s.App.ProtoRevKeeper.SetMaxPointsPerBlock(s.Ctx, 0)
	s.Require().Error(err)
	err = s.App.ProtoRevKeeper.SetMaxPointsPerBlock(s.Ctx, types.MaxPoolPointsPerBlock+1)
	s.Require().Error(err)
}

// TestGetInfoByPoolType tests the GetInfoByPoolType and SetInfoByPoolType functions.
func (s *KeeperTestSuite) TestGetInfoByPoolType() {
	// Should be able to set the InfoByPoolType
	newRouteWeights := types.DefaultPoolTypeInfo
	newRouteWeights.Balancer.Weight = 100
	newRouteWeights.Cosmwasm.WeightMaps = []types.WeightMap{
		{
			ContractAddress: "contractAddress",
			Weight:          1,
		},
	}

	s.App.ProtoRevKeeper.SetInfoByPoolType(s.Ctx, newRouteWeights)

	poolWeights := s.App.ProtoRevKeeper.GetInfoByPoolType(s.Ctx)
	s.Require().Equal(newRouteWeights, poolWeights)
}

func (s *KeeperTestSuite) TestGetAllProtocolRevenue() {
	s.SetupPoolsTest()
	baseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)
	communityPoolDenom := "Akash"
	atom := "Atom"
	ion := "uion"

	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.MustNewDecFromStr("0.02")
	poolManagerParams.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo = communityPoolDenom
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	allProtoRev := s.App.ProtoRevKeeper.GetAllProtocolRevenue(s.Ctx)
	s.Require().Equal(types.AllProtocolRevenue{
		TakerFeesTracker: poolmanagertypes.TakerFeesTracker{
			TakerFeesToStakers:         []sdk.Coin{},
			TakerFeesToCommunityPool:   []sdk.Coin{},
			HeightAccountingStartsFrom: 0,
		},
		CyclicArbTracker: types.CyclicArbTracker{
			CyclicArb:                  []sdk.Coin{},
			HeightAccountingStartsFrom: 0,
		},
	}, allProtoRev)

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
	allProtoRev = s.App.ProtoRevKeeper.GetAllProtocolRevenue(s.Ctx)
	s.Require().Equal([]sdk.Coin(cyclicArbProfits), allProtoRev.CyclicArbTracker.CyclicArb)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToStakers), allProtoRev.TakerFeesTracker.TakerFeesToStakers)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToCommunityPool), allProtoRev.TakerFeesTracker.TakerFeesToCommunityPool)

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
	allProtoRev = s.App.ProtoRevKeeper.GetAllProtocolRevenue(s.Ctx)
	s.Require().Equal([]sdk.Coin(cyclicArbProfits.Add(cyclicArbProfits...)), allProtoRev.CyclicArbTracker.CyclicArb)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToStakers.Add(expectedTakerFeeToStakers...)), allProtoRev.TakerFeesTracker.TakerFeesToStakers)
	s.Require().Equal([]sdk.Coin(expectedTakerFeeToCommunityPool.Add(expectedTakerFeeToCommunityPool...)), allProtoRev.TakerFeesTracker.TakerFeesToCommunityPool)
}
