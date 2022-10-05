package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer    types.MsgServer
	defaultDenom types.MsgCreateDenomResponse
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) CreateDefaultDenom() {
	res, _ := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.MsgCreateDenom())
	fmt.Println(res)
}
