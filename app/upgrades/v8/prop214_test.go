package v8_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	v8 "github.com/osmosis-labs/osmosis/v7/app/upgrades/v8"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestProp214() {
	poolId := suite.PrepareBalancerPool()
	v8.Prop214(suite.Ctx, suite.App.GAMMKeeper)

	pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
	suite.Require().NoError(err)

	swapFee := pool.GetSwapFee(suite.Ctx)
	expectedSwapFee := sdk.MustNewDecFromStr("0.002")

	suite.Require().Equal(expectedSwapFee, swapFee)
}
