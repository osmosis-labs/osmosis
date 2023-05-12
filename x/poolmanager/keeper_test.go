package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

const testExpectedPoolId = 3

var (
	testPoolCreationFee = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)}
	testPoolRoute       = []types.ModuleRoute{
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

// createBalancerPoolsFromCoinsWithSwapFee creates balancer pools from given sets of coins and respective swap fees.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (s *KeeperTestSuite) createBalancerPoolsFromCoinsWithSwapFee(poolCoins []sdk.Coins, swapFee []sdk.Dec) {
	for i, curPoolCoins := range poolCoins {
		s.FundAcc(s.TestAccs[0], curPoolCoins)
		s.PrepareCustomBalancerPoolFromCoins(curPoolCoins, balancer.PoolParams{
			SwapFee: swapFee[i],
			ExitFee: sdk.ZeroDec(),
		})
	}
}

// createBalancerPoolsFromCoins creates balancer pools from given sets of coins and zero swap fees.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (suite *KeeperTestSuite) createBalancerPoolsFromCoins(poolCoins []sdk.Coins) {
	for _, curPoolCoins := range poolCoins {
		suite.FundAcc(suite.TestAccs[0], curPoolCoins)
		suite.PrepareCustomBalancerPoolFromCoins(curPoolCoins, balancer.PoolParams{
			SwapFee: sdk.ZeroDec(),
			ExitFee: sdk.ZeroDec(),
		})
	}
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	suite.Setup()

	suite.App.PoolManagerKeeper.InitGenesis(suite.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
	})

	suite.Require().Equal(uint64(testExpectedPoolId), suite.App.PoolManagerKeeper.GetNextPoolId(suite.Ctx))
	suite.Require().Equal(testPoolCreationFee, suite.App.PoolManagerKeeper.GetParams(suite.Ctx).PoolCreationFee)
	suite.Require().Equal(testPoolRoute, suite.App.PoolManagerKeeper.GetAllPoolRoutes(suite.Ctx))
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.Setup()

	s.App.PoolManagerKeeper.InitGenesis(s.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
	})

	genesis := s.App.PoolManagerKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(uint64(testExpectedPoolId), genesis.NextPoolId)
	s.Require().Equal(testPoolCreationFee, genesis.Params.PoolCreationFee)
	s.Require().Equal(testPoolRoute, genesis.PoolRoutes)
}
