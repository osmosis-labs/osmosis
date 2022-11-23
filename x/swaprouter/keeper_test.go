package swaprouter_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

const testExpectedPoolId = 3

var testPoolCreationFee = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	suite.Setup()

	suite.App.SwapRouterKeeper.InitGenesis(suite.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
	})

	suite.Require().Equal(uint64(testExpectedPoolId), suite.App.SwapRouterKeeper.GetNextPoolIdAndIncrement(suite.Ctx))
	suite.Require().Equal(testPoolCreationFee, suite.App.SwapRouterKeeper.GetParams(suite.Ctx).PoolCreationFee)
}

func (suite *KeeperTestSuite) TestExportGenesis() {
	suite.Setup()

	suite.App.SwapRouterKeeper.InitGenesis(suite.Ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
	})

	genesis := suite.App.SwapRouterKeeper.ExportGenesis(suite.Ctx)
	suite.Require().Equal(uint64(testExpectedPoolId), genesis.NextPoolId)
	suite.Require().Equal(testPoolCreationFee, genesis.Params.PoolCreationFee)
}
