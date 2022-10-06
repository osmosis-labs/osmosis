package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	cleanup func()
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func (suite *KeeperTestSuite) Cleanup() {
	suite.cleanup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// func (suite *KeeperTestSuite) CreateDefaultDenom() {
// 	res, _ := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.MsgCreateDenom())
// 	fmt.Println(res)
// }
