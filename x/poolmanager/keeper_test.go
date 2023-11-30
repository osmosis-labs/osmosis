package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/app/apptesting"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
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
	testAuthorizedQuoteDenoms                          = []string{"uosmo", "uion", "uatom"}

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
		TakerFeesToStakers:         sdk.Coins{sdk.NewCoin("uosmo", sdk.NewInt(1000))},
		TakerFeesToCommunityPool:   sdk.Coins{sdk.NewCoin("uusdc", sdk.NewInt(1000))},
		HeightAccountingStartsFrom: 100,
	}

	testPoolVolumes = []*types.PoolVolume{
		{
			PoolId:     1,
			PoolVolume: sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000000))),
		},
		{
			PoolId:     2,
			PoolVolume: sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(20000000))),
		},
	}

	testDenomPairTakerFees = []types.DenomPairTakerFee{
		{
			Denom0:   "uion",
			Denom1:   "uosmo",
			TakerFee: osmomath.MustNewDecFromStr("0.0016"),
		},
		{
			Denom0:   "uatom",
			Denom1:   "uosmo",
			TakerFee: osmomath.MustNewDecFromStr("0.002"),
		},
	}
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Set the bond denom to be uosmo to make volume tracking tests more readable.
	skParams := s.App.StakingKeeper.GetParams(s.Ctx)
	skParams.BondDenom = "uosmo"
	s.App.StakingKeeper.SetParams(s.Ctx, skParams)
	s.App.TxFeesKeeper.SetBaseDenom(s.Ctx, "uosmo")
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

	takerFee, err := s.App.PoolManagerKeeper.GetTradingPairTakerFee(s.Ctx, testDenomPairTakerFees[0].Denom0, testDenomPairTakerFees[0].Denom1)
	s.Require().NoError(err)
	s.Require().Equal(testDenomPairTakerFees[0].TakerFee, takerFee)
	takerFee, err = s.App.PoolManagerKeeper.GetTradingPairTakerFee(s.Ctx, testDenomPairTakerFees[1].Denom0, testDenomPairTakerFees[1].Denom1)
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
