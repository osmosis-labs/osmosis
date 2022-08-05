package ibc_rate_limit_test

import (
	"encoding/json"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/osmosis-labs/osmosis/v10/app"
	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	ibc_rate_limit "github.com/osmosis-labs/osmosis/v10/x/ibc-rate-limit"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MiddlewareTestSuite struct {
	apptesting.KeeperTestHelper

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	//path *ibctesting.Path

	RateLimitMiddlware ibc_rate_limit.RateLimitMiddleware
}

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return app.Setup(false), map[string]json.RawMessage{}
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.Setup()
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	//path := NewTransferPath(suite.chainA, suite.chainB)
	//suite.coordinator.SetupConnections(path)
	//
	ibctesting.DefaultTestingAppInit = SetupTestingApp

	//path.EndpointA.ChannelID = ibctesting.FirstChannelID
	//counterparty := channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
	//channel := &channeltypes.Channel{
	//	State:          channeltypes.INIT,
	//	Ordering:       channeltypes.UNORDERED,
	//	Counterparty:   counterparty,
	//	ConnectionHops: []string{path.EndpointA.ConnectionID},
	//	Version:        transfertypes.Version,
	//}
	//
	//msg := types.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0)
	//res, err := suite.chainA.SendMsgs(msg)
	//suite.Require().NoError(err) // message committed

}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version
	return path
}

func (suite *MiddlewareTestSuite) TestSendPacket() {
	path := NewTransferPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(path)                       // clientID, connectionID, channelID filled
	//suite.Require().Equal("07-tendermint-0", path.EndpointA.ClientID)
	//suite.Require().Equal("connection-0", path.EndpointA.ConnectionID)
	//suite.Require().Equal("channel-0", path.EndpointA.ChannelID)

	disabledTimeoutTimestamp := uint64(0)
	timeoutHeight := clienttypes.NewHeight(0, 100)
	packet := channeltypes.NewPacket(ibctesting.MockPacketData, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, disabledTimeoutTimestamp)
	channelCap := suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.SendPacket(suite.chainA.GetContext(), channelCap, packet)

	suite.Require().NoError(err)

	// receive on endpointB
	//path.EndpointB.RecvPacket(packet1)

	// acknowledge the receipt of the packet
	//path.EndpointA.AcknowledgePacket(packet1, ack)

}
