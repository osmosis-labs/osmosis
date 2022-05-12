package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (suite *KeeperTestSuite) TestApplyUpdateParam() {
	suite.SetupTest()

	// Mint some assets to the accounts.
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, defaultAcctFunds)
	}

	// Create the pool at first
	msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}, defaultPoolAssets, defaultFutureGovernor)
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.Require().NoError(err)

	suite.App.GAMMKeeper.ApplyUpdateParam(suite.Ctx, types.UpdatePoolParam{
		PoolId:    poolId,
		RiskLevel: types.RiskLevel_UNPOOL_ALLOWED,
	})

	pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
	suite.Require().NoError(err)

	suite.Require().Equal(pool.GetRiskLevel(suite.Ctx), types.RiskLevel_UNPOOL_ALLOWED)
}
