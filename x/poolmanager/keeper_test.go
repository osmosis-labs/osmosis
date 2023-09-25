package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
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
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
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
}

func (s *KeeperTestSuite) TestExportGenesis() {
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
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
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
}
