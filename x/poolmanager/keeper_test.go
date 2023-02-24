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

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

// createPoolFromType creates a basic pool of the given type for testing.
func (suite *KeeperTestSuite) createPoolFromType(poolType types.PoolType) {
	switch poolType {
	case types.Balancer:
		suite.PrepareBalancerPool()
		return
	case types.Stableswap:
		suite.PrepareBasicStableswapPool()
		return
	case types.Concentrated:
		// TODO
		return
	}
}

// createBalancerPoolsFromCoinsWithSwapFee creates balancer pools from given sets of coins and respective swap fees.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (suite *KeeperTestSuite) createBalancerPoolsFromCoinsWithSwapFee(poolCoins []sdk.Coins, swapFee []sdk.Dec) {
	for i, curPoolCoins := range poolCoins {
		suite.FundAcc(suite.TestAccs[0], curPoolCoins)
		suite.PrepareCustomBalancerPoolFromCoins(curPoolCoins, balancer.PoolParams{
			SwapFee: swapFee[i],
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

func (suite *KeeperTestSuite) TestExportGenesis() {
	suite.Setup()

	suite.App.PoolManagerKeeper.InitGenesis(suite.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
		PoolRoutes: testPoolRoute,
	})

	genesis := suite.App.PoolManagerKeeper.ExportGenesis(suite.Ctx)
	suite.Require().Equal(uint64(testExpectedPoolId), genesis.NextPoolId)
	suite.Require().Equal(testPoolCreationFee, genesis.Params.PoolCreationFee)
	suite.Require().Equal(testPoolRoute, genesis.PoolRoutes)
}
