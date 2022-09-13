package ibc_metadata_test

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/osmosis-labs/osmosis/v12/app"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmosisibctesting"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-metadata/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MiddlewareTestSuite struct {
	apptesting.KeeperTestHelper

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *osmosisibctesting.TestChain
	chainB *osmosisibctesting.TestChain
	path   *ibctesting.Path
}

// Setup
func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	osmosisApp := app.Setup(false)
	return osmosisApp, app.NewDefaultGenesisState()
}

func (suite *MiddlewareTestSuite) InstantiateSwapRouterContract(chain *osmosisibctesting.TestChain) sdk.AccAddress {
	initMsgBz := []byte(fmt.Sprintf(`{"owner": "%v"}`, suite.TestAccs[0].String()))
	addr, err := chain.InstantiateContract(1, initMsgBz, "swap router contract")
	suite.Require().NoError(err)
	return addr
}

func (suite *MiddlewareTestSuite) NewMessage(fromEndpoint *ibctesting.Endpoint, fromAccount, toAccount authtypes.AccountI, amount sdk.Int) sdk.Msg {
	return transfertypes.NewMsgTransfer(
		fromEndpoint.ChannelConfig.PortID,
		fromEndpoint.ChannelID,
		sdk.NewCoin(sdk.DefaultBondDenom, amount),
		fromAccount.GetAddress().String(),
		toAccount.GetAddress().String(),
		clienttypes.NewHeight(0, 100),
		0,
	)
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.Setup()
	ibctesting.DefaultTestingAppInit = SetupTestingApp
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(1)),
	}
	// Remove epochs to prevent  minting
	suite.chainA.MoveEpochsToTheFuture()
	suite.chainB = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(2)),
	}
	suite.path = osmosisibctesting.NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(suite.path)
}

func (suite *MiddlewareTestSuite) TestSendTransferWithoutMetadata() {
	msg := suite.NewMessage(suite.path.EndpointA, suite.chainA.SenderAccount, suite.chainB.SenderAccount, sdk.NewInt(1))
	_, err := suite.chainA.SendMsgsNoCheck(msg)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)
}

func (suite *MiddlewareTestSuite) TestSendTransferWithMetadata() {
	channelCap := suite.chainA.GetChannelCapability(
		suite.path.EndpointA.ChannelConfig.PortID,
		suite.path.EndpointA.ChannelID)
	packetData := types.FungibleTokenPacketData{
		Denom:    sdk.DefaultBondDenom,
		Amount:   "1",
		Sender:   suite.chainA.SenderAccount.GetAddress().String(),
		Receiver: suite.chainB.SenderAccount.GetAddress().String(),
		Metadata: []byte(`{"callback": "something"}`),
	}

	var packet = channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		suite.path.EndpointA.ChannelConfig.PortID,
		suite.path.EndpointA.ChannelID,
		suite.path.EndpointB.ChannelConfig.PortID,
		suite.path.EndpointB.ChannelID,
		clienttypes.NewHeight(0, 100),
		0,
	)
	err := suite.chainA.GetOsmosisApp().MetadataICS4Wrapper.SendPacket(
		suite.chainA.GetContext(), channelCap, packet)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)
}

func (suite *MiddlewareTestSuite) receivePacket(metadata []byte) []byte {
	channelCap := suite.chainB.GetChannelCapability(
		suite.path.EndpointB.ChannelConfig.PortID,
		suite.path.EndpointB.ChannelID)
	packetData := types.FungibleTokenPacketData{
		Denom:    sdk.DefaultBondDenom,
		Amount:   "1",
		Sender:   suite.chainB.SenderAccount.GetAddress().String(),
		Receiver: suite.chainA.SenderAccount.GetAddress().String(),
		Metadata: metadata,
	}

	var packet = channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		suite.path.EndpointB.ChannelConfig.PortID,
		suite.path.EndpointB.ChannelID,
		suite.path.EndpointA.ChannelConfig.PortID,
		suite.path.EndpointA.ChannelID,
		clienttypes.NewHeight(0, 100),
		0,
	)
	err := suite.chainB.GetOsmosisApp().MetadataICS4Wrapper.SendPacket(
		suite.chainB.GetContext(), channelCap, packet)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)

	// Update both clients
	err = suite.path.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// recv in chain a
	res, err := suite.path.EndpointA.RecvPacketWithResult(packet)

	// get the ack from the chain a's response
	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// manually send the acknowledgement to chain b
	err = suite.path.EndpointA.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)
	return ack
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithoutMetadata() {
	suite.receivePacket(nil)
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithBadMetadata() {
	ack := suite.receivePacket([]byte(`{"callback": 1234}`))
	suite.Require().Contains(string(ack), "error")
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithMetadata() {
	ackBytes := suite.receivePacket([]byte(`{"callback": "test2"}`))
	fmt.Println(string(ackBytes))
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	fmt.Println(ack)
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithSwap() {
	err := suite.chainA.StoreContractCode("./testdata/swaprouter.wasm")
	suite.Require().NoError(err)
	addr := suite.InstantiateSwapRouterContract(suite.chainA)
	contractAddr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	suite.Require().NoError(err)
	ackBytes := suite.receivePacket([]byte(fmt.Sprintf(`{"callback": "%v"}`, contractAddr)))
	fmt.Println(string(ackBytes))
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	fmt.Println(ack)
}
