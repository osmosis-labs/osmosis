package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gogoproto/proto"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	poolmanager "github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

const testExpectedPoolId = 3

var (
	testPoolCreationFee          = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)}
	testDefaultTakerFee          = osmomath.MustNewDecFromStr("0.0015")
	testOsmoTakerFeeDistribution = types.TakerFeeDistributionPercentage{
		StakingRewards: osmomath.MustNewDecFromStr("0.3"),
		CommunityPool:  osmomath.MustNewDecFromStr("0.7"),
	}
	testNonOsmoTakerFeeDistribution = types.TakerFeeDistributionPercentage{
		StakingRewards: osmomath.MustNewDecFromStr("0.2"),
		CommunityPool:  osmomath.MustNewDecFromStr("0.8"),
	}
	testAdminAddresses                                 = []string{"osmo106x8q2nv7xsg7qrec2zgdf3vvq0t3gn49zvaha", "osmo105l5r3rjtynn7lg362r2m9hkpfvmgmjtkglsn9"}
	testCommunityPoolDenomToSwapNonWhitelistedAssetsTo = "uusdc"
	testAuthorizedQuoteDenoms                          = []string{appparams.BaseCoinUnit, "uion", "uatom"}

	testPoolRoute = []types.ModuleRoute{
		{
			PoolId:   1,
			PoolType: types.Balancer,
		},
		{
			PoolId:   2,
			PoolType: types.Stableswap,
		},
	}

	testTakerFeesTracker = types.TakerFeesTracker{
		TakerFeesToStakers:         sdk.Coins{sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000))},
		TakerFeesToCommunityPool:   sdk.Coins{sdk.NewCoin("uusdc", osmomath.NewInt(1000))},
		HeightAccountingStartsFrom: 100,
	}

	testPoolVolumes = []*types.PoolVolume{
		{
			PoolId:     1,
			PoolVolume: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000000))),
		},
		{
			PoolId:     2,
			PoolVolume: sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(20000000))),
		},
	}

	testDenomPairTakerFees = []types.DenomPairTakerFee{
		{
			TokenInDenom:  "uion",
			TokenOutDenom: appparams.BaseCoinUnit,
			TakerFee:      osmomath.MustNewDecFromStr("0.0016"),
		},
		{
			TokenInDenom:  "uatom",
			TokenOutDenom: appparams.BaseCoinUnit,
			TakerFee:      osmomath.MustNewDecFromStr("0.002"),
		},
	}
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Set the bond denom to be uosmo to make volume tracking tests more readable.
	skParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	skParams.BondDenom = appparams.BaseCoinUnit
	s.App.StakingKeeper.SetParams(s.Ctx, skParams)
	s.App.TxFeesKeeper.SetBaseDenom(s.Ctx, appparams.BaseCoinUnit)
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo = "baz"
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
}

// createBalancerPoolsFromCoinsWithSpreadFactor creates balancer pools from given sets of coins and respective spread factors.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (s *KeeperTestSuite) createBalancerPoolsFromCoinsWithSpreadFactor(poolCoins []sdk.Coins, spreadFactor []osmomath.Dec) {
	for i, curPoolCoins := range poolCoins {
		s.FundAcc(s.TestAccs[0], curPoolCoins)
		s.PrepareCustomBalancerPoolFromCoins(curPoolCoins, balancer.PoolParams{
			SwapFee: spreadFactor[i],
			ExitFee: osmomath.ZeroDec(),
		})
	}
}

// createBalancerPoolsFromCoins creates balancer pools from given sets of coins and zero swap fees.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (s *KeeperTestSuite) createBalancerPoolsFromCoins(poolCoins []sdk.Coins) {
	for _, curPoolCoins := range poolCoins {
		s.FundAcc(s.TestAccs[0], curPoolCoins)
		s.PrepareCustomBalancerPoolFromCoins(curPoolCoins, balancer.PoolParams{
			SwapFee: osmomath.ZeroDec(),
			ExitFee: osmomath.ZeroDec(),
		})
	}
}

func (s *KeeperTestSuite) TestInitGenesis() {
	s.App.PoolManagerKeeper.InitGenesis(s.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
			TakerFeeParams: types.TakerFeeParams{
				DefaultTakerFee:                                testDefaultTakerFee,
				OsmoTakerFeeDistribution:                       testOsmoTakerFeeDistribution,
				NonOsmoTakerFeeDistribution:                    testNonOsmoTakerFeeDistribution,
				AdminAddresses:                                 testAdminAddresses,
				CommunityPoolDenomToSwapNonWhitelistedAssetsTo: testCommunityPoolDenomToSwapNonWhitelistedAssetsTo,
			},
			AuthorizedQuoteDenoms: testAuthorizedQuoteDenoms,
		},
		NextPoolId:             testExpectedPoolId,
		PoolRoutes:             testPoolRoute,
		TakerFeesTracker:       &testTakerFeesTracker,
		PoolVolumes:            testPoolVolumes,
		DenomPairTakerFeeStore: testDenomPairTakerFees,
	})

	params := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	s.Require().Equal(uint64(testExpectedPoolId), s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx))
	s.Require().Equal(testPoolCreationFee, params.PoolCreationFee)
	s.Require().Equal(testDefaultTakerFee, params.TakerFeeParams.DefaultTakerFee)
	s.Require().Equal(testOsmoTakerFeeDistribution, params.TakerFeeParams.OsmoTakerFeeDistribution)
	s.Require().Equal(testNonOsmoTakerFeeDistribution, params.TakerFeeParams.NonOsmoTakerFeeDistribution)
	s.Require().Equal(testAdminAddresses, params.TakerFeeParams.AdminAddresses)
	s.Require().Equal(testCommunityPoolDenomToSwapNonWhitelistedAssetsTo, params.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo)
	s.Require().Equal(testAuthorizedQuoteDenoms, params.AuthorizedQuoteDenoms)
	s.Require().Equal(testPoolRoute, s.App.PoolManagerKeeper.GetAllPoolRoutes(s.Ctx))
	s.Require().Equal(testTakerFeesTracker.TakerFeesToStakers, s.App.PoolManagerKeeper.GetTakerFeeTrackerForStakers(s.Ctx))
	s.Require().Equal(testTakerFeesTracker.TakerFeesToCommunityPool, s.App.PoolManagerKeeper.GetTakerFeeTrackerForCommunityPool(s.Ctx))
	s.Require().Equal(testTakerFeesTracker.HeightAccountingStartsFrom, s.App.PoolManagerKeeper.GetTakerFeeTrackerStartHeight(s.Ctx))
	s.Require().Equal(testPoolVolumes[0].PoolVolume, s.App.PoolManagerKeeper.GetTotalVolumeForPool(s.Ctx, testPoolVolumes[0].PoolId))
	s.Require().Equal(testPoolVolumes[1].PoolVolume, s.App.PoolManagerKeeper.GetTotalVolumeForPool(s.Ctx, testPoolVolumes[1].PoolId))

	takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(s.Ctx, testDenomPairTakerFees[0].TokenInDenom, testDenomPairTakerFees[0].TokenOutDenom)
	s.Require().NoError(err)
	s.Require().Equal(testDenomPairTakerFees[0].TakerFee, takerFee)
	takerFee, err = s.App.PoolManagerKeeper.GetTradingPairTakerFee(s.Ctx, testDenomPairTakerFees[1].TokenInDenom, testDenomPairTakerFees[1].TokenOutDenom)
	s.Require().NoError(err)
	s.Require().Equal(testDenomPairTakerFees[1].TakerFee, takerFee)
}

func (s *KeeperTestSuite) TestExportGenesis() {
	// Need to create two pools to properly export pool volumes.
	s.PrepareBalancerPool()
	s.PrepareConcentratedPool()

	s.App.PoolManagerKeeper.InitGenesis(s.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
			TakerFeeParams: types.TakerFeeParams{
				DefaultTakerFee:                                testDefaultTakerFee,
				OsmoTakerFeeDistribution:                       testOsmoTakerFeeDistribution,
				NonOsmoTakerFeeDistribution:                    testNonOsmoTakerFeeDistribution,
				AdminAddresses:                                 testAdminAddresses,
				CommunityPoolDenomToSwapNonWhitelistedAssetsTo: testCommunityPoolDenomToSwapNonWhitelistedAssetsTo,
			},
			AuthorizedQuoteDenoms: testAuthorizedQuoteDenoms,
		},
		NextPoolId:             testExpectedPoolId,
		PoolRoutes:             testPoolRoute,
		TakerFeesTracker:       &testTakerFeesTracker,
		PoolVolumes:            testPoolVolumes,
		DenomPairTakerFeeStore: testDenomPairTakerFees,
	})

	genesis := s.App.PoolManagerKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(uint64(testExpectedPoolId), genesis.NextPoolId)
	s.Require().Equal(testPoolCreationFee, genesis.Params.PoolCreationFee)
	s.Require().Equal(testDefaultTakerFee, genesis.Params.TakerFeeParams.DefaultTakerFee)
	s.Require().Equal(testOsmoTakerFeeDistribution, genesis.Params.TakerFeeParams.OsmoTakerFeeDistribution)
	s.Require().Equal(testNonOsmoTakerFeeDistribution, genesis.Params.TakerFeeParams.NonOsmoTakerFeeDistribution)
	s.Require().Equal(testAdminAddresses, genesis.Params.TakerFeeParams.AdminAddresses)
	s.Require().Equal(testCommunityPoolDenomToSwapNonWhitelistedAssetsTo, genesis.Params.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo)
	s.Require().Equal(testAuthorizedQuoteDenoms, genesis.Params.AuthorizedQuoteDenoms)
	s.Require().Equal(testPoolRoute, genesis.PoolRoutes)
	s.Require().Equal(testTakerFeesTracker.TakerFeesToStakers, genesis.TakerFeesTracker.TakerFeesToStakers)
	s.Require().Equal(testTakerFeesTracker.TakerFeesToCommunityPool, genesis.TakerFeesTracker.TakerFeesToCommunityPool)
	s.Require().Equal(testTakerFeesTracker.HeightAccountingStartsFrom, genesis.TakerFeesTracker.HeightAccountingStartsFrom)
	s.Require().Equal(testPoolVolumes[0].PoolVolume, genesis.PoolVolumes[0].PoolVolume)
	s.Require().Equal(testPoolVolumes[1].PoolVolume, genesis.PoolVolumes[1].PoolVolume)
	s.Require().Equal(testDenomPairTakerFees, genesis.DenomPairTakerFeeStore)
}

// TestBeginBlock tests that, if any one of the cache trackers is empty, all cache trackers are updated.
// NOTE: We should only ever be in a state where all cache trackers are empty or all are non-empty, but we
// test various scenarios here to ensure that the cache trackers are updated correctly.
func (s *KeeperTestSuite) TestBeginBlock() {
	contractAddress := "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2"
	alloyedDenom := createAlloyedDenom(contractAddress)

	// Define the default values for all three cache trackers
	defaultCachedTakerFeeShareAgreementMap := map[string]types.TakerFeeShareAgreement{
		defaultTakerFeeShareAgreements[0].Denom: defaultTakerFeeShareAgreements[0],
	}
	defaultCachedRegisteredAlloyPoolByAlloyDenomMap := map[string]types.AlloyContractTakerFeeShareState{
		alloyedDenom: {
			ContractAddress:         contractAddress,
			TakerFeeShareAgreements: []types.TakerFeeShareAgreement{defaultTakerFeeShareAgreements[0]},
		},
	}

	tests := map[string]struct {
		storeSetup                                       func()
		expectedCachedTakerFeeShareAgreementMap          map[string]types.TakerFeeShareAgreement
		expectedCachedRegisteredAlloyPoolByAlloyDenomMap map[string]types.AlloyContractTakerFeeShareState
	}{
		"cachedTakerFeeShareAgreementMap is empty, cachedRegisteredAlloyPoolByAlloyDenomMap is empty, should update": {
			storeSetup: func() {
				s.App.PoolManagerKeeper.SetCacheTrackers(nil, nil)
			},
			expectedCachedTakerFeeShareAgreementMap:          defaultCachedTakerFeeShareAgreementMap,
			expectedCachedRegisteredAlloyPoolByAlloyDenomMap: defaultCachedRegisteredAlloyPoolByAlloyDenomMap,
		},
		"cachedTakerFeeShareAgreementMap is empty, cachedRegisteredAlloyPoolByAlloyDenomMap is not empty, should update": {
			storeSetup: func() {
				s.App.PoolManagerKeeper.SetCacheTrackers(nil, defaultCachedRegisteredAlloyPoolByAlloyDenomMap)
			},
			expectedCachedTakerFeeShareAgreementMap:          defaultCachedTakerFeeShareAgreementMap,
			expectedCachedRegisteredAlloyPoolByAlloyDenomMap: defaultCachedRegisteredAlloyPoolByAlloyDenomMap,
		},
		"cachedTakerFeeShareAgreementMap is not empty, cachedRegisteredAlloyPoolByAlloyDenomMap is empty, should update": {
			storeSetup: func() {
				s.App.PoolManagerKeeper.SetCacheTrackers(defaultCachedTakerFeeShareAgreementMap, nil)
			},
			expectedCachedTakerFeeShareAgreementMap:          defaultCachedTakerFeeShareAgreementMap,
			expectedCachedRegisteredAlloyPoolByAlloyDenomMap: defaultCachedRegisteredAlloyPoolByAlloyDenomMap,
		},
		"cachedTakerFeeShareAgreementMap is not empty, cachedRegisteredAlloyPoolByAlloyDenomMap is not empty, should update": {
			storeSetup: func() {
				s.App.PoolManagerKeeper.SetCacheTrackers(defaultCachedTakerFeeShareAgreementMap, defaultCachedRegisteredAlloyPoolByAlloyDenomMap)
			},
			expectedCachedTakerFeeShareAgreementMap:          defaultCachedTakerFeeShareAgreementMap,
			expectedCachedRegisteredAlloyPoolByAlloyDenomMap: defaultCachedRegisteredAlloyPoolByAlloyDenomMap,
		},
		"cachedTakerFeeShareAgreementMap is not empty, cachedRegisteredAlloyPoolByAlloyDenomMap is not empty, should not update": {
			storeSetup: func() {
				differentCachedTakerFeeShareAgreement := map[string]types.TakerFeeShareAgreement{
					defaultTakerFeeShareAgreements[0].Denom: {
						Denom:       defaultTakerFeeShareAgreements[0].Denom,
						SkimPercent: osmomath.MustNewDecFromStr("0.02"),
						SkimAddress: defaultTakerFeeShareAgreements[0].SkimAddress,
					},
				}
				differentCachedRegisteredAlloyPoolToState := map[string]types.AlloyContractTakerFeeShareState{
					createAlloyedDenom(contractAddress): {
						ContractAddress: contractAddress,
						TakerFeeShareAgreements: []types.TakerFeeShareAgreement{
							{
								Denom:       defaultTakerFeeShareAgreements[0].Denom,
								SkimPercent: osmomath.MustNewDecFromStr("0.02"),
								SkimAddress: defaultTakerFeeShareAgreements[0].SkimAddress,
							},
						},
					},
				}
				s.App.PoolManagerKeeper.SetCacheTrackers(differentCachedTakerFeeShareAgreement, differentCachedRegisteredAlloyPoolToState)
			},
			expectedCachedTakerFeeShareAgreementMap: map[string]types.TakerFeeShareAgreement{
				defaultTakerFeeShareAgreements[0].Denom: {
					Denom:       defaultTakerFeeShareAgreements[0].Denom,
					SkimPercent: osmomath.MustNewDecFromStr("0.02"),
					SkimAddress: defaultTakerFeeShareAgreements[0].SkimAddress,
				},
			},
			expectedCachedRegisteredAlloyPoolByAlloyDenomMap: map[string]types.AlloyContractTakerFeeShareState{
				createAlloyedDenom(contractAddress): {
					ContractAddress: contractAddress,
					TakerFeeShareAgreements: []types.TakerFeeShareAgreement{
						{
							Denom:       defaultTakerFeeShareAgreements[0].Denom,
							SkimPercent: osmomath.MustNewDecFromStr("0.02"),
							SkimAddress: defaultTakerFeeShareAgreements[0].SkimAddress,
						},
					},
				},
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			// Directly set the stores
			takerFeeShareAgreement := defaultTakerFeeShareAgreements[0]
			poolManagerKey := s.App.AppKeepers.GetKey(types.StoreKey)
			store := s.Ctx.KVStore(poolManagerKey)
			key := types.FormatTakerFeeShareAgreementKey(takerFeeShareAgreement.Denom)
			bz, err := proto.Marshal(&takerFeeShareAgreement)
			s.Require().NoError(err)
			store.Set(key, bz)

			alloyContractState := types.AlloyContractTakerFeeShareState{
				ContractAddress:         contractAddress,
				TakerFeeShareAgreements: []types.TakerFeeShareAgreement{takerFeeShareAgreement},
			}
			bz, err = proto.Marshal(&alloyContractState)
			s.Require().NoError(err)
			key = types.FormatRegisteredAlloyPoolKey(1, alloyedDenom)
			store.Set(key, bz)

			// Set up cachedStores
			tc.storeSetup()

			// Call BeginBlock
			s.App.PoolManagerKeeper.BeginBlock(s.Ctx)

			// Check expected values
			cachedTakerFeeShareAgreementMap, cachedRegisteredAlloyPoolByAlloyDenomMap := s.App.PoolManagerKeeper.GetCacheTrackers()
			s.Require().Equal(tc.expectedCachedTakerFeeShareAgreementMap, cachedTakerFeeShareAgreementMap)
			s.Require().Equal(tc.expectedCachedRegisteredAlloyPoolByAlloyDenomMap, cachedRegisteredAlloyPoolByAlloyDenomMap)
		})
	}
}

// TestEndBlock tests the behavior of the EndBlock method in the PoolManagerKeeper.
// Specifically, it verifies that the taker fee share alloy composition is updated correctly
// for registered alloyed pools at block heights that are multiples of the alloyedAssetCompositionUpdateRate.
func (s *KeeperTestSuite) TestEndBlock() {
	tests := map[string]struct {
		blockHeight                     int64
		swapFunc                        func()
		registerPool                    bool
		expectedTakerFeeShareAgreements []types.TakerFeeShareAgreement
		expectedError                   error
	}{
		"alloyed pool registered, alloyed pool changes, alloy composition changes at multiple of alloyedAssetCompositionUpdateRate": {
			blockHeight: poolmanager.AlloyedAssetCompositionUpdateRate,
			swapFunc: func() {
				joinCoins := sdk.NewCoins(sdk.NewInt64Coin(denomA, 1000000000))
				s.FundAcc(s.TestAccs[0], joinCoins)
				s.JoinTransmuterPool(s.TestAccs[0], 1, joinCoins)
			},
			registerPool: true,
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.66666666666666666"),
				osmomath.MustNewDecFromStr("0.33333333333333333"),
			}),
		},
		"alloyed pool registered, non alloyed pool changes, alloy composition does not change at multiple of alloyedAssetCompositionUpdateRate": {
			blockHeight: poolmanager.AlloyedAssetCompositionUpdateRate,
			swapFunc: func() {
				s.PrepareAllSupportedPools()
				s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, 1)
			},
			registerPool: true,
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.5"),
				osmomath.MustNewDecFromStr("0.5"),
			}),
		},
		"alloyed pool registered, alloy composition does not change at non-multiple of alloyedAssetCompositionUpdateRate": {
			blockHeight: poolmanager.AlloyedAssetCompositionUpdateRate + 1,
			swapFunc: func() {
				joinCoins := sdk.NewCoins(sdk.NewInt64Coin(denomA, 1000000000))
				s.FundAcc(s.TestAccs[0], joinCoins)
				s.JoinTransmuterPool(s.TestAccs[0], 1, joinCoins)
			},
			registerPool: true,
			expectedTakerFeeShareAgreements: modifySkimPercent(defaultTakerFeeShareAgreements[:2], []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.5"),
				osmomath.MustNewDecFromStr("0.5"),
			}),
		},
		"pool not registered, no changes to alloy composition": {
			blockHeight: poolmanager.AlloyedAssetCompositionUpdateRate,
			swapFunc: func() {
				joinCoins := sdk.NewCoins(sdk.NewInt64Coin(denomA, 1000000000))
				s.FundAcc(s.TestAccs[0], joinCoins)
				s.JoinTransmuterPool(s.TestAccs[0], 1, joinCoins)
			},
			registerPool:  false,
			expectedError: types.NoRegisteredAlloyedPoolError{PoolId: 1},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			cwPool := s.PrepareCustomTransmuterPoolV3(s.TestAccs[0], []string{denomA, denomB}, []uint16{1, 1})
			if tc.registerPool {
				s.App.PoolManagerKeeper.SetRegisteredAlloyedPool(s.Ctx, cwPool.GetId())
			}
			setTakerFeeShareAgreements(s.Ctx, s.App.PoolManagerKeeper, defaultTakerFeeShareAgreements[:2])

			// Set up stores
			tc.swapFunc()

			// Set block height
			s.Ctx = s.Ctx.WithBlockHeight(tc.blockHeight)

			// Call EndBlock
			s.App.PoolManagerKeeper.EndBlock(s.Ctx)

			// Check expected values
			takerFeeShareState, err := s.App.PoolManagerKeeper.GetRegisteredAlloyedPoolFromPoolId(s.Ctx, cwPool.GetId())
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedTakerFeeShareAgreements, takerFeeShareState.TakerFeeShareAgreements)
		})
	}
}

func (s *KeeperTestSuite) TestFundCommunityPoolIfNotWhitelisted() {
	tests := []struct {
		name           string
		whitelist      []string
		sender         sdk.AccAddress
		expectFundCall bool
	}{
		{
			name:           "sender is whitelisted",
			whitelist:      []string{"osmo1044qatzg4a0wm63jchrfdnn2u8nwdgxxt6e524"},
			sender:         sdk.MustAccAddressFromBech32("osmo1044qatzg4a0wm63jchrfdnn2u8nwdgxxt6e524"),
			expectFundCall: false,
		},
		{
			name:           "sender is not whitelisted",
			whitelist:      []string{"osmo1044qatzg4a0wm63jchrfdnn2u8nwdgxxt6e524"},
			sender:         sdk.MustAccAddressFromBech32("osmo1j537vtv60wz322n2sgfm4st7y3dm8e4e9js57h"),
			expectFundCall: true,
		},
		{
			name:           "whitelist is empty",
			whitelist:      []string{},
			sender:         sdk.MustAccAddressFromBech32("osmo1j537vtv60wz322n2sgfm4st7y3dm8e4e9js57h"),
			expectFundCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			oldParams := s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx)
			oldParams.UnrestrictedPoolCreatorWhitelist = tc.whitelist
			s.App.ConcentratedLiquidityKeeper.SetParams(s.Ctx, oldParams)

			// Fund the sender with pool creation fee
			poolCreationFee := s.App.PoolManagerKeeper.GetParams(s.Ctx).PoolCreationFee
			s.FundAcc(tc.sender, poolCreationFee)

			// Get the community pool balance
			preCommunityPoolBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName))

			err := s.App.PoolManagerKeeper.FundCommunityPoolIfNotWhitelisted(s.Ctx, tc.sender)
			s.Require().NoError(err)

			// Get the community pool balance after the function call
			postCommunityPoolBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName))

			if tc.expectFundCall {
				s.Require().Equal(postCommunityPoolBalance, preCommunityPoolBalance.Add(poolCreationFee...))
			} else {
				s.Require().Equal(preCommunityPoolBalance, postCommunityPoolBalance)
			}
		})
	}
}
