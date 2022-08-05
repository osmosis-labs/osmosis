package ibc_rate_limit_test

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (suite *MiddlewareTestSuite) DontRunTestSendPacketNoIBC() {
	// This does the same as ibctesting but manually (to not depend on the extra methods on OsmosisApp)

	channel := channeltypes.NewChannel(
		channeltypes.OPEN, channeltypes.ORDERED,
		channeltypes.NewCounterparty("sourcePort", "sourceChannel"),
		[]string{"sourceConnection"}, ibctesting.DefaultChannelVersion)

	suite.App.IBCKeeper.ChannelKeeper.SetChannel(suite.Ctx, "sourcePort", "sourceChannel", channel)

	suite.App.IBCKeeper.ChannelKeeper.SetNextChannelSequence(suite.Ctx, 2)
	suite.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.Ctx, "sourcePort", "sourceChannel", 2)
	suite.App.IBCKeeper.ChannelKeeper.SetNextSequenceRecv(suite.Ctx, "sourcePort", "sourceChannel", 2)
	suite.App.IBCKeeper.ChannelKeeper.SetNextSequenceAck(suite.Ctx, "sourcePort", "sourceChannel", 2)

	suite.T().Log("scopped", suite.App.TransferKeeper.IsBound(suite.Ctx, "sourcePort"))

	//_, ok := suite.App.ScopedTransferKeeper.GetCapability(suite.Ctx, host.PortPath("sourcePort2"))
	//if !ok {
	//	// create capability using the IBC capability keeper
	//	suite.T().Log("creating")
	//	cap, err := suite.App.ScopedTransferKeeper.NewCapability(suite.Ctx, host.PortPath("sourcePort"))
	//	suite.T().Log("cap", cap)
	//	require.NoError(suite.T(), err)
	//
	//	// claim capability using the scopedKeeper
	//	err = suite.App.ScopedTransferKeeper.ClaimCapability(suite.Ctx, cap, host.PortPath("sourcePort"))
	//	require.NoError(suite.T(), err)
	//	suite.T().Log("created")
	//
	//}

	err := suite.App.TransferKeeper.SendTransfer(
		suite.Ctx,
		"sourcePort",
		"sourceChannel",
		sdk.NewCoin("nosmo", sdk.NewInt(1)),
		suite.TestAccs[0],
		"receiver",
		clienttypes.NewHeight(100, 100),
		100,
	)
	suite.T().Log("Transfer sent")

	suite.Require().NoError(err)
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

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
