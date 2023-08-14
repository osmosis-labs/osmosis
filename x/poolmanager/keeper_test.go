package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v17/app/apptesting"
	"github.com/osmosis-labs/osmosis/v17/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

const testExpectedPoolId = 3

var (
	testPoolCreationFee    = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)}
	testDefaultTakerFee    = sdk.MustNewDecFromStr("0.0015")
	testStableswapTakerFee = sdk.MustNewDecFromStr("0.0002")
	testCustomPoolTakerFee = []types.CustomPoolTakerFee{
		{
			PoolId:   1,
			TakerFee: sdk.MustNewDecFromStr("0.0005"),
		},
		{
			PoolId:   2,
			TakerFee: sdk.MustNewDecFromStr("0.0001"),
		},
	}
	testOsmoTakerFeeDistribution = types.TakerFeeDistributionPercentage{
		StakingRewards: sdk.MustNewDecFromStr("0.3"),
		CommunityPool:  sdk.MustNewDecFromStr("0.7"),
	}
	testNonOsmoTakerFeeDistribution = types.TakerFeeDistributionPercentage{
		StakingRewards: sdk.MustNewDecFromStr("0.2"),
		CommunityPool:  sdk.MustNewDecFromStr("0.8"),
	}
	testAuthorizedQuoteDenoms                          = []string{"uosmo", "uion", "uatom"}
	testCommunityPoolDenomToSwapNonWhitelistedAssetsTo = "uusdc"

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
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

// createBalancerPoolsFromCoinsWithSpreadFactor creates balancer pools from given sets of coins and respective spread factors.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (s *KeeperTestSuite) createBalancerPoolsFromCoinsWithSpreadFactor(poolCoins []sdk.Coins, spreadFactor []sdk.Dec) {
	for i, curPoolCoins := range poolCoins {
		s.FundAcc(s.TestAccs[0], curPoolCoins)
		s.PrepareCustomBalancerPoolFromCoins(curPoolCoins, balancer.PoolParams{
			SwapFee: spreadFactor[i],
			ExitFee: sdk.ZeroDec(),
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
			SwapFee: sdk.ZeroDec(),
			ExitFee: sdk.ZeroDec(),
		})
	}
}

func (s *KeeperTestSuite) TestInitGenesis() {
	s.App.PoolManagerKeeper.InitGenesis(s.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee:                                testPoolCreationFee,
			DefaultTakerFee:                                testDefaultTakerFee,
			StableswapTakerFee:                             testStableswapTakerFee,
			CustomPoolTakerFee:                             testCustomPoolTakerFee,
			OsmoTakerFeeDistribution:                       testOsmoTakerFeeDistribution,
			NonOsmoTakerFeeDistribution:                    testNonOsmoTakerFeeDistribution,
			AuthorizedQuoteDenoms:                          testAuthorizedQuoteDenoms,
			CommunityPoolDenomToSwapNonWhitelistedAssetsTo: testCommunityPoolDenomToSwapNonWhitelistedAssetsTo,
		},
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
	})

	params := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	s.Require().Equal(uint64(testExpectedPoolId), s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx))
	s.Require().Equal(testPoolCreationFee, params.PoolCreationFee)
	s.Require().Equal(testDefaultTakerFee, params.DefaultTakerFee)
	s.Require().Equal(testStableswapTakerFee, params.StableswapTakerFee)
	s.Require().Equal(testCustomPoolTakerFee, params.CustomPoolTakerFee)
	s.Require().Equal(testOsmoTakerFeeDistribution, params.OsmoTakerFeeDistribution)
	s.Require().Equal(testNonOsmoTakerFeeDistribution, params.NonOsmoTakerFeeDistribution)
	s.Require().Equal(testAuthorizedQuoteDenoms, params.AuthorizedQuoteDenoms)
	s.Require().Equal(testPoolRoute, s.App.PoolManagerKeeper.GetAllPoolRoutes(s.Ctx))
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.App.PoolManagerKeeper.InitGenesis(s.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee:                                testPoolCreationFee,
			DefaultTakerFee:                                testDefaultTakerFee,
			StableswapTakerFee:                             testStableswapTakerFee,
			CustomPoolTakerFee:                             testCustomPoolTakerFee,
			OsmoTakerFeeDistribution:                       testOsmoTakerFeeDistribution,
			NonOsmoTakerFeeDistribution:                    testNonOsmoTakerFeeDistribution,
			AuthorizedQuoteDenoms:                          testAuthorizedQuoteDenoms,
			CommunityPoolDenomToSwapNonWhitelistedAssetsTo: testCommunityPoolDenomToSwapNonWhitelistedAssetsTo,
		},
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
	})

	genesis := s.App.PoolManagerKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(uint64(testExpectedPoolId), genesis.NextPoolId)
	s.Require().Equal(testPoolCreationFee, genesis.Params.PoolCreationFee)
	s.Require().Equal(testDefaultTakerFee, genesis.Params.DefaultTakerFee)
	s.Require().Equal(testStableswapTakerFee, genesis.Params.StableswapTakerFee)
	s.Require().Equal(testCustomPoolTakerFee, genesis.Params.CustomPoolTakerFee)
	s.Require().Equal(testOsmoTakerFeeDistribution, genesis.Params.OsmoTakerFeeDistribution)
	s.Require().Equal(testNonOsmoTakerFeeDistribution, genesis.Params.NonOsmoTakerFeeDistribution)
	s.Require().Equal(testAuthorizedQuoteDenoms, genesis.Params.AuthorizedQuoteDenoms)
	s.Require().Equal(testCommunityPoolDenomToSwapNonWhitelistedAssetsTo, genesis.Params.CommunityPoolDenomToSwapNonWhitelistedAssetsTo)
	s.Require().Equal(testPoolRoute, genesis.PoolRoutes)
}
