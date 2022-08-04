package ibc_rate_limit_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	//ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	ibc_rate_limit "github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MiddlewareTestSuite struct {
	apptesting.KeeperTestHelper

	// Uncommenting this line (and the import) makes everything fail
	//coordinator *ibctesting.Coordinator

	App                *app.OsmosisApp
	Ctx                sdk.Context
	RateLimitMiddlware ibc_rate_limit.RateLimitMiddleware
}

func (suite *MiddlewareTestSuite) SetupCustomApp() {
	suite.App = app.Setup(false)
	//suite.RateLimitMiddlware = suite.App.Router().Route()
}

func (suite *MiddlewareTestSuite) SetupTest() {
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

// Uncommenting this line (and the import) makes everything fail
//func NewTransferPath(chainA, chainB *ibctesting.TestChain) {}

func (suite *MiddlewareTestSuite) CreateMockPacket() channeltypes.Packet {
	return channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "sourcePort",
		SourceChannel:      "sourceChannel",
		DestinationPort:    "destPort",
		DestinationChannel: "destChannel",
		Data:               []byte("mock packet data"),
		TimeoutHeight:      clienttypes.NewHeight(0, 100),
	}
}

func (suite *MiddlewareTestSuite) TestSendPacket() {
	suite.T().Log("Say bye")
}
