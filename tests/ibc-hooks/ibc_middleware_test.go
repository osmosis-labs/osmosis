package ibc_hooks_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/tidwall/gjson"

	"github.com/CosmWasm/wasmd/x/wasm/types"

	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
	txfeetypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/osmosis-labs/osmosis/v27/tests/osmosisibctesting"

	"github.com/osmosis-labs/osmosis/v27/tests/ibc-hooks/testutils"
)

type HooksTestSuite struct {
	apptesting.KeeperTestHelper

	coordinator *ibctesting.Coordinator

	chainA *osmosisibctesting.TestChain
	chainB *osmosisibctesting.TestChain
	chainC *osmosisibctesting.TestChain

	pathAB *ibctesting.Path
	pathAC *ibctesting.Path
	pathBC *ibctesting.Path
	// This is used to test cw20s. It will only get assigned in the cw20 test
	pathCW20 *ibctesting.Path
}

var oldConsensusMinFee = txfeetypes.ConsensusMinFee

const defaultPoolAmount int64 = 100000

// TODO: This needs to get removed. Waiting on https://github.com/cosmos/ibc-go/issues/3123
func (suite *HooksTestSuite) TearDownSuite() {
	txfeetypes.ConsensusMinFee = oldConsensusMinFee
}

func TestIBCHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func (suite *HooksTestSuite) SetupTest() {
	suite.SkipIfWSL()
	// TODO: This needs to get removed. Waiting on https://github.com/cosmos/ibc-go/issues/3123
	txfeetypes.ConsensusMinFee = osmomath.ZeroDec()

	suite.Setup()
	ibctesting.DefaultTestingAppInit = osmosisibctesting.SetupTestingApp
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(1)),
	}
	suite.chainB = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(2)),
	}
	suite.chainC = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(3)),
	}
	err := suite.chainA.MoveEpochsToTheFuture()
	suite.Require().NoError(err)
	err = suite.chainB.MoveEpochsToTheFuture()
	suite.Require().NoError(err)
	err = suite.chainC.MoveEpochsToTheFuture()
	suite.Require().NoError(err)
	suite.pathAB = NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(suite.pathAB)
	suite.pathBC = NewTransferPath(suite.chainB, suite.chainC)
	suite.coordinator.Setup(suite.pathBC)
	suite.pathAC = NewTransferPath(suite.chainA, suite.chainC)
	suite.coordinator.Setup(suite.pathAC)
}

func NewTransferPath(chainA, chainB *osmosisibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

type Chain int64

const (
	ChainA Chain = iota
	ChainB
	ChainC
)

func (suite *HooksTestSuite) GetChain(name Chain) *osmosisibctesting.TestChain {
	switch name {
	case ChainA:
		return suite.chainA
	case ChainB:
		return suite.chainB
	case ChainC:
		return suite.chainC
	}
	return nil
}

type Direction int64

const (
	AtoB Direction = iota
	BtoA
	AtoC
	CtoA
	BtoC
	CtoB
	CW20toA
	AtoCW20
)

func (suite *HooksTestSuite) GetEndpoints(direction Direction) (sender *ibctesting.Endpoint, receiver *ibctesting.Endpoint) {
	switch direction {
	case AtoB:
		sender = suite.pathAB.EndpointA
		receiver = suite.pathAB.EndpointB
	case BtoA:
		sender = suite.pathAB.EndpointB
		receiver = suite.pathAB.EndpointA
	case AtoC:
		sender = suite.pathAC.EndpointA
		receiver = suite.pathAC.EndpointB
	case CtoA:
		sender = suite.pathAC.EndpointB
		receiver = suite.pathAC.EndpointA
	case BtoC:
		sender = suite.pathBC.EndpointA
		receiver = suite.pathBC.EndpointB
	case CtoB:
		sender = suite.pathBC.EndpointB
		receiver = suite.pathBC.EndpointA
	case CW20toA:
		sender = suite.pathCW20.EndpointB
		receiver = suite.pathCW20.EndpointA
	case AtoCW20:
		sender = suite.pathCW20.EndpointA
		receiver = suite.pathCW20.EndpointB
	default:
		panic("invalid direction")
	}
	return sender, receiver
}

// Get direction from chain pair
func (suite *HooksTestSuite) GetDirection(chainA, chainB Chain) Direction {
	switch {
	case chainA == ChainA && chainB == ChainB:
		return AtoB
	case chainA == ChainB && chainB == ChainA:
		return BtoA
	case chainA == ChainA && chainB == ChainC:
		return AtoC
	case chainA == ChainC && chainB == ChainA:
		return CtoA
	case chainA == ChainB && chainB == ChainC:
		return BtoC
	case chainA == ChainC && chainB == ChainB:
		return CtoB
	default:
		fmt.Println(chainA, chainB)
		panic("invalid chain pair")
	}
}

func (suite *HooksTestSuite) GetSenderChannel(chainA, chainB Chain) string {
	sender, _ := suite.GetEndpoints(suite.GetDirection(chainA, chainB))
	return sender.ChannelID
}

func (suite *HooksTestSuite) GetReceiverChannel(chainA, chainB Chain) string {
	_, receiver := suite.GetEndpoints(suite.GetDirection(chainA, chainB))
	return receiver.ChannelID
}

func (suite *HooksTestSuite) TestDeriveIntermediateSender() {

	testCases := []struct {
		channel         string
		originalSender  string
		bech32Prefix    string
		expectedAddress string
	}{
		{
			channel:         "channel-0",
			originalSender:  "cosmos1tfejvgp5yzd8ypvn9t0e2uv2kcjf2laa8upya8",
			bech32Prefix:    "osmo",
			expectedAddress: "osmo1sguz3gtyl2tjsdulwxmtprd68xtd43yyep6g5c554utz642sr8rqcgw0q6",
		},
		{
			channel:         "channel-1",
			originalSender:  "cosmos1tfejvgp5yzd8ypvn9t0e2uv2kcjf2laa8upya8",
			bech32Prefix:    "osmo",
			expectedAddress: "osmo1svnare87kluww5hnltv24m4dg72hst0qqwm5xslsvnwd22gftcussaz5l7",
		},
		{
			channel:         "channel-0",
			originalSender:  "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj",
			bech32Prefix:    "osmo",
			expectedAddress: "osmo1vz8evs4ek3vnz4f8wy86nw9ayzn67y28vtxzjgxv6achc4pa8gesqldfz0",
		},
		{
			channel:         "channel-0",
			originalSender:  "osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj",
			bech32Prefix:    "cosmos",
			expectedAddress: "cosmos1vz8evs4ek3vnz4f8wy86nw9ayzn67y28vtxzjgxv6achc4pa8ges4z434f",
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Test failed for case (channel=%s, originalSender=%s, bech32Prefix=%s).",
			tc.channel, tc.originalSender, tc.bech32Prefix), func() {
			actualAddress, err := ibchookskeeper.DeriveIntermediateSender(tc.channel, tc.originalSender, tc.bech32Prefix)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedAddress, actualAddress)
		})
	}
}

func (suite *HooksTestSuite) TestOnRecvPacketHooks() {
	var (
		trace    transfertypes.DenomTrace
		amount   osmomath.Int
		receiver string
		status   testutils.Status
	)

	testCases := []struct {
		msg      string
		malleate func(*testutils.Status)
		expPass  bool
	}{
		{"override", func(status *testutils.Status) {
			suite.chainB.GetOsmosisApp().TransferStack.
				ICS4Middleware.Hooks = testutils.TestRecvOverrideHooks{Status: status}
		}, true},
		{"before and after", func(status *testutils.Status) {
			suite.chainB.GetOsmosisApp().TransferStack.
				ICS4Middleware.Hooks = testutils.TestRecvBeforeAfterHooks{Status: status}
		}, true},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.msg, func() {
			suite.SetupTest() // reset

			path := NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(path)
			receiver = suite.chainB.SenderAccount.GetAddress().String() // must be explicitly changed in malleate
			status = testutils.Status{}

			amount = osmomath.NewInt(100) // must be explicitly changed in malleate
			seq := uint64(1)

			trace = transfertypes.ParseDenomTrace(sdk.DefaultBondDenom)

			// send coin from chainA to chainB
			transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.NewCoin(trace.IBCDenom(), amount), suite.chainA.SenderAccount.GetAddress().String(), receiver, clienttypes.NewHeight(1, 110), 0, "")
			_, err := suite.chainA.SendMsgs(transferMsg)
			suite.Require().NoError(err) // message committed

			tc.malleate(&status)

			data := transfertypes.NewFungibleTokenPacketData(trace.GetFullDenomPath(), amount.String(), suite.chainA.SenderAccount.GetAddress().String(), receiver, "")
			packet := channeltypes.NewPacket(data.GetBytes(), seq, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, clienttypes.NewHeight(1, 100), 0)

			ack := suite.chainB.GetOsmosisApp().TransferStack.
				OnRecvPacket(suite.chainB.GetContext(), packet, suite.chainA.SenderAccount.GetAddress())

			if tc.expPass {
				suite.Require().True(ack.Success())
			} else {
				suite.Require().False(ack.Success())
			}

			if _, ok := suite.chainB.GetOsmosisApp().TransferStack.
				ICS4Middleware.Hooks.(testutils.TestRecvOverrideHooks); ok {
				suite.Require().True(status.OverrideRan)
				suite.Require().False(status.BeforeRan)
				suite.Require().False(status.AfterRan)
			}

			if _, ok := suite.chainB.GetOsmosisApp().TransferStack.
				ICS4Middleware.Hooks.(testutils.TestRecvBeforeAfterHooks); ok {
				suite.Require().False(status.OverrideRan)
				suite.Require().True(status.BeforeRan)
				suite.Require().True(status.AfterRan)
			}
		})
	}
}

func (suite *HooksTestSuite) makeMockPacket(receiver, memo string, prevSequence uint64) channeltypes.Packet {
	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    sdk.DefaultBondDenom,
		Amount:   "1",
		Sender:   suite.chainB.SenderAccount.GetAddress().String(),
		Receiver: receiver,
		Memo:     memo,
	}

	return channeltypes.NewPacket(
		packetData.GetBytes(),
		prevSequence+1,
		suite.pathAB.EndpointB.ChannelConfig.PortID,
		suite.pathAB.EndpointB.ChannelID,
		suite.pathAB.EndpointA.ChannelConfig.PortID,
		suite.pathAB.EndpointA.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)
}

func (suite *HooksTestSuite) receivePacket(receiver, memo string) []byte {
	return suite.receivePacketWithSequence(receiver, memo, 0)
}

func (suite *HooksTestSuite) receivePacketWithSequence(receiver, memo string, prevSequence uint64) []byte {
	channelCap := suite.chainB.GetChannelCapability(
		suite.pathAB.EndpointB.ChannelConfig.PortID,
		suite.pathAB.EndpointB.ChannelID)

	packet := suite.makeMockPacket(receiver, memo, prevSequence)

	_, err := suite.chainB.GetOsmosisApp().HooksICS4Wrapper.SendPacket(
		suite.chainB.GetContext(), channelCap, packet.SourcePort, packet.SourceChannel, packet.TimeoutHeight, packet.TimeoutTimestamp, packet.Data)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)

	// Update both clients
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// recv in chain a
	res, err := suite.pathAB.EndpointA.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	// get the ack from the chain a's response
	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// manually send the acknowledgement to chain b
	err = suite.pathAB.EndpointA.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)
	return ack
}

func (suite *HooksTestSuite) TestRecvTransferWithMetadata() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/echo.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, "{}", 1)

	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")
}

// After successfully executing a wasm call, the contract should have the funds sent via IBC
func (suite *HooksTestSuite) TestFundsAreTransferredToTheContract() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/echo.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, "{}", 1)

	// Check that the contract has no funds
	localDenom := osmoutils.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
	balance := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	suite.Require().Equal(osmomath.NewInt(0), balance.Amount)

	// Execute the contract via IBC
	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")
	suite.Require().Equal(ack["result"], "eyJjb250cmFjdF9yZXN1bHQiOiJkR2hwY3lCemFHOTFiR1FnWldOb2J3PT0iLCJpYmNfYWNrIjoiZXlKeVpYTjFiSFFpT2lKQlVUMDlJbjA9In0=")

	// Check that the token has now been transferred to the contract
	balance = suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	suite.Require().Equal(osmomath.NewInt(1), balance.Amount)
}

// If the wasm call wails, the contract acknowledgement should be an error and the funds returned
func (suite *HooksTestSuite) TestFundsAreReturnedOnFailedContractExec() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/echo.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, "{}", 1)

	// Check that the contract has no funds
	localDenom := osmoutils.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
	balance := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	suite.Require().Equal(osmomath.NewInt(0), balance.Amount)

	// Execute the contract via IBC with a message that the contract will reject
	ackBytes := suite.receivePacket(addr.String(), fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"not_echo": {"msg": "test"} } } }`, addr))
	ackStr := string(ackBytes)
	fmt.Println(ackStr)
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().Contains(ack, "error")

	// Check that the token has now been transferred to the contract
	balance = suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	fmt.Println(balance)
	suite.Require().Equal(osmomath.NewInt(0), balance.Amount)
}

func (suite *HooksTestSuite) TestPacketsThatShouldBeSkipped() {
	var sequence uint64
	receiver := suite.chainB.SenderAccount.GetAddress().String()

	testCases := []struct {
		memo           string
		expPassthrough bool
	}{
		{"", true},
		{"{01]", true}, // bad json
		{"{}", true},
		{`{"something": ""}`, true},
		{`{"wasm": "test"}`, false},
		{`{"wasm": []`, true}, // invalid top level JSON
		{`{"wasm": {}`, true}, // invalid top level JSON
		{`{"wasm": []}`, false},
		{`{"wasm": {}}`, false},
		{`{"wasm": {"contract": "something"}}`, false},
		{`{"wasm": {"contract": "osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj"}}`, false},
		{`{"wasm": {"msg": "something"}}`, false},
		// invalid receiver
		{`{"wasm": {"contract": "osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4epasmvnj", "msg": {}}}`, false},
		// msg not an object
		{fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": 1}}`, receiver), false},
	}

	for _, tc := range testCases {
		ackBytes := suite.receivePacketWithSequence(receiver, tc.memo, sequence)
		ackStr := string(ackBytes)
		fmt.Println(ackStr)
		var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
		err := json.Unmarshal(ackBytes, &ack)
		suite.Require().NoError(err)
		if tc.expPassthrough {
			suite.Require().Equal("AQ==", ack["result"], tc.memo)
		} else {
			suite.Require().Contains(ackStr, "error", tc.memo)
		}
		sequence += 1
	}
}

// After successfully executing a wasm call, the contract should have the funds sent via IBC
func (suite *HooksTestSuite) TestFundTracking() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/counter.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)

	// Check that the contract has no funds
	localDenom := osmoutils.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
	balance := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	suite.Require().Equal(osmomath.NewInt(0), balance.Amount)

	// Execute the contract via IBC
	suite.receivePacket(
		addr.String(),
		fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"increment": {} } } }`, addr))

	senderLocalAcc, err := ibchookskeeper.DeriveIntermediateSender("channel-0", suite.chainB.SenderAccount.GetAddress().String(), "osmo")
	suite.Require().NoError(err)

	state := suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderLocalAcc)))
	suite.Require().Equal(`{"count":0}`, state)

	state = suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderLocalAcc)))
	suite.Require().Equal(`{"total_funds":[{"denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","amount":"1"}]}`, state)

	suite.receivePacketWithSequence(
		addr.String(),
		fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"increment": {} } } }`, addr), 1)

	state = suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderLocalAcc)))
	suite.Require().Equal(`{"count":1}`, state)

	state = suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderLocalAcc)))
	suite.Require().Equal(`{"total_funds":[{"denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","amount":"2"}]}`, state)

	// Check that the token has now been transferred to the contract
	balance = suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	suite.Require().Equal(osmomath.NewInt(2), balance.Amount)
}

// custom MsgTransfer constructor that supports Memo
func NewMsgTransfer(token sdk.Coin, sender, receiver, channel, memo string) *transfertypes.MsgTransfer {
	return &transfertypes.MsgTransfer{
		SourcePort:       "transfer",
		SourceChannel:    channel,
		Token:            token,
		Sender:           sender,
		Receiver:         receiver,
		TimeoutHeight:    clienttypes.NewHeight(1, 500),
		TimeoutTimestamp: 0,
		Memo:             memo,
	}
}

func (suite *HooksTestSuite) RelayPacket(packet channeltypes.Packet, direction Direction) (*abci.ExecTxResult, []byte) {
	sender, receiver := suite.GetEndpoints(direction)

	err := receiver.UpdateClient()
	suite.Require().NoError(err)

	// receiver Receives
	receiveResult, err := receiver.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)

	if strings.Contains(string(ack), "error") {
		fmt.Println(receiveResult.Events)
		errorCtx := gjson.Get(receiveResult.Log, "0.events.#(type==ibccallbackerror-ibc-acknowledgement-error)#.attributes.#(key==ibccallbackerror-error-context)#.value")
		fmt.Println("ibc-ack-error:", errorCtx)
	}

	// sender Acknowledges
	err = sender.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)

	err = sender.UpdateClient()
	suite.Require().NoError(err)
	err = receiver.UpdateClient()
	suite.Require().NoError(err)

	return receiveResult, ack
}

func (suite *HooksTestSuite) RelayPacketNoAck(packet channeltypes.Packet, direction Direction) *abci.ExecTxResult {
	sender, receiver := suite.GetEndpoints(direction)

	err := receiver.UpdateClient()
	suite.Require().NoError(err)

	// receiver Receives
	receiveResult, err := receiver.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	err = sender.UpdateClient()
	suite.Require().NoError(err)
	err = receiver.UpdateClient()
	suite.Require().NoError(err)

	return receiveResult
}

func (suite *HooksTestSuite) FullSend(msg sdk.Msg, direction Direction) (*abci.ExecTxResult, *abci.ExecTxResult, string, error) {
	var sender *osmosisibctesting.TestChain
	switch direction {
	case AtoB:
		sender = suite.chainA
	case BtoA:
		sender = suite.chainB
	case BtoC:
		sender = suite.chainB
	case CtoB:
		sender = suite.chainC
	case AtoC:
		sender = suite.chainA
	case CtoA:
		sender = suite.chainC
	case CW20toA:
		sender = suite.chainB
	case AtoCW20:
		sender = suite.chainA
	default:
		panic("invalid direction")
	}
	sendResult, err := sender.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	suite.Require().NoError(err)

	receiveResult, ack := suite.RelayPacket(packet, direction)
	if strings.Contains(string(ack), "error") {
		errorCtx := gjson.Get(receiveResult.Log, "0.events.#(type==ibc-acknowledgement-error)#.attributes.#(key==error-context)#.value")
		fmt.Println("ibc-ack-error:", errorCtx)
	}
	return sendResult, receiveResult, string(ack), err
}

func (suite *HooksTestSuite) TestAcks() {
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/counter.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)

	// Generate swap instructions for the contract
	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), addr.String(), "channel-0", callbackMemo)
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)

	// The test contract will increment the counter for itself every time it receives an ack
	state := suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":1}`, state)

	_, _, _, err = suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)
	state = suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":2}`, state)
}

func (suite *HooksTestSuite) TestTimeouts() {
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/counter.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)

	// Generate swap instructions for the contract
	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), addr.String(), "channel-0", callbackMemo)
	transferMsg.TimeoutTimestamp = uint64(suite.coordinator.CurrentTime.Add(time.Minute).UnixNano())
	sendResult, err := suite.chainA.SendMsgsNoCheck(transferMsg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	suite.Require().NoError(err)

	// Move chainB forward one block
	suite.chainB.NextBlock()
	// One month later
	suite.coordinator.IncrementTimeBy(time.Hour)
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	err = suite.pathAB.EndpointA.TimeoutPacket(packet)
	suite.Require().NoError(err)

	// The test contract will increment the counter for itself by 10 when a packet times out
	state := suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":10}`, state)
}

func (suite *HooksTestSuite) TestSendWithoutMemo() {
	// Sending a packet without memo to ensure that the ibc_callback middleware doesn't interfere with a regular send
	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), "channel-0", "")
	_, _, ack, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)
	suite.Require().Contains(ack, "result")
}

// This is a copy of the SetupGammPoolsWithBondDenomMultiplier from the  test helpers, but using chainA instead of the default
func (suite *HooksTestSuite) SetupPools(chainName Chain, multipliers []osmomath.Dec) []gammtypes.CFMMPoolI {
	chain := suite.GetChain(chainName)
	acc1 := chain.SenderAccount.GetAddress()
	bondDenom, err := chain.GetOsmosisApp().StakingKeeper.BondDenom(chain.GetContext())
	suite.Require().NoError(err)

	pools := []gammtypes.CFMMPoolI{}
	for index, multiplier := range multipliers {
		token := fmt.Sprintf("token%d", index)
		uosmoAmount := gammtypes.InitPoolSharesSupply.ToLegacyDec().Mul(multiplier).RoundInt()

		var (
			defaultFutureGovernor = ""

			// pool assets
			defaultFooAsset = balancer.PoolAsset{
				Weight: osmomath.NewInt(100),
				Token:  sdk.NewCoin(bondDenom, uosmoAmount),
			}
			defaultBarAsset = balancer.PoolAsset{
				Weight: osmomath.NewInt(100),
				Token:  sdk.NewCoin(token, osmomath.NewInt(10000)),
			}

			poolAssets = []balancer.PoolAsset{defaultFooAsset, defaultBarAsset}
		)

		poolParams := balancer.PoolParams{
			SwapFee: osmomath.NewDecWithPrec(1, 2),
			ExitFee: osmomath.ZeroDec(),
		}
		msg := balancer.NewMsgCreateBalancerPool(acc1, poolParams, poolAssets, defaultFutureGovernor)

		poolId, err := chain.GetOsmosisApp().PoolManagerKeeper.CreatePool(chain.GetContext(), msg)
		suite.Require().NoError(err)

		pool, err := chain.GetOsmosisApp().GAMMKeeper.GetPoolAndPoke(chain.GetContext(), poolId)
		suite.Require().NoError(err)

		pools = append(pools, pool)
	}

	return pools
}

func (suite *HooksTestSuite) SetupCrosschainSwaps(chainName Chain, setupForwarding bool) (sdk.AccAddress, sdk.AccAddress) {
	chain := suite.GetChain(chainName)
	owner := chain.SenderAccount.GetAddress()

	registryAddr, _, _, _ := suite.SetupCrosschainRegistry(chainName)
	suite.setChainChannelLinks(registryAddr, chainName)
	suite.setAllPrefixesToOsmo(registryAddr, chainName)
	if setupForwarding {
		suite.setForwardingOnAllChains(registryAddr)
	}

	// Fund the account with some uosmo and some stake
	bankKeeper := chain.GetOsmosisApp().BankKeeper
	i, ok := osmomath.NewIntFromString("20000000000000000000000")
	suite.Require().True(ok)
	amounts := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, i), sdk.NewCoin(sdk.DefaultBondDenom, i), sdk.NewCoin("token0", i), sdk.NewCoin("token1", i))
	err := bankKeeper.MintCoins(chain.GetContext(), minttypes.ModuleName, amounts)
	suite.Require().NoError(err)
	err = bankKeeper.SendCoinsFromModuleToAccount(chain.GetContext(), minttypes.ModuleName, owner, amounts)
	suite.Require().NoError(err)

	suite.SetupPools(chainName, []osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	// Setup contract
	chain.StoreContractCode(&suite.Suite, "./bytecode/swaprouter.wasm")
	swaprouterAddr := chain.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"owner": "%s"}`, owner), 2)
	chain.StoreContractCode(&suite.Suite, "./bytecode/crosschain_swaps.wasm")

	crosschainAddr := chain.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"swap_contract": "%s", "governor": "%s", "registry_contract":"%s"}`, swaprouterAddr, owner, registryAddr),
		3)

	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	ctx := chain.GetContext()

	msg := `{"set_route":{"input_denom":"token0","output_denom":"token1","pool_route":[{"pool_id":"1","token_out_denom":"stake"},{"pool_id":"2","token_out_denom":"token1"}]}}`
	_, err = contractKeeper.Execute(ctx, swaprouterAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	// Move forward one block
	chain.NextBlock()
	chain.Coordinator.IncrementTime()

	// Update both clients
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	return swaprouterAddr, crosschainAddr
}

func (suite *HooksTestSuite) fundAccount(chain *osmosisibctesting.TestChain, owner sdk.AccAddress) {
	// TODO: allow this function to fund with custom token names (calling them tokenA, tokenB, etc. would make tests easier to read, I think)
	bankKeeper := chain.GetOsmosisApp().BankKeeper
	i, ok := osmomath.NewIntFromString("20000000000000000000000")
	suite.Require().True(ok)
	amounts := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, i), sdk.NewCoin(sdk.DefaultBondDenom, i), sdk.NewCoin("token0", i), sdk.NewCoin("token1", i))
	err := bankKeeper.MintCoins(chain.GetContext(), minttypes.ModuleName, amounts)
	suite.Require().NoError(err)
	err = bankKeeper.SendCoinsFromModuleToAccount(chain.GetContext(), minttypes.ModuleName, owner, amounts)
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) SetupCrosschainRegistry(chainName Chain) (sdk.AccAddress, string, string, string) {
	chain := suite.GetChain(chainName)
	owner := chain.SenderAccount.GetAddress()

	// Fund the account with some uosmo and some stake.
	for _, ch := range []*osmosisibctesting.TestChain{suite.chainA, suite.chainB, suite.chainC} {
		suite.fundAccount(ch, ch.SenderAccount.GetAddress())
	}

	// Setup pools
	suite.SetupPools(chainName, []osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	// Setup contract
	chain.StoreContractCode(&suite.Suite, "./bytecode/crosschain_registry.wasm")
	registryAddr := chain.InstantiateContract(&suite.Suite, fmt.Sprintf(`{"owner": "%s"}`, owner), 1)
	_, err := sdk.Bech32ifyAddressBytes("symphony", registryAddr)
	suite.Require().NoError(err)

	// Send some token0 tokens from C to B
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(2000)), suite.chainC.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), "channel-0", "")
	_, _, _, err = suite.FullSend(transferMsg, CtoB)
	suite.Require().NoError(err)

	// Send some token0 tokens from B to A
	transferMsg = NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(2000)), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), "channel-0", "")
	_, _, _, err = suite.FullSend(transferMsg, BtoA)
	suite.Require().NoError(err)

	// Send some token0 tokens from C to B to A
	denomTrace0CB := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathBC.EndpointA.ChannelID, "token0"))
	token0CB := denomTrace0CB.IBCDenom()
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0CB, osmomath.NewInt(2000)), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), "channel-0", "")
	_, _, _, err = suite.FullSend(transferMsg, BtoA)
	suite.Require().NoError(err)

	// Denom traces
	CBAPath := fmt.Sprintf("transfer/%s/transfer/%s", suite.pathAB.EndpointA.ChannelID, suite.pathBC.EndpointA.ChannelID)
	denomTrace0CBA := transfertypes.DenomTrace{Path: CBAPath, BaseDenom: "token0"}
	token0CBA := denomTrace0CBA.IBCDenom()

	// Move forward one block
	chain.NextBlock()
	chain.Coordinator.IncrementTime()

	// Update both clients
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathBC.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	return registryAddr, token0CB, token0CBA, CBAPath
}

func (suite *HooksTestSuite) setChainChannelLinks(registryAddr sdk.AccAddress, chainName Chain) {
	chain := suite.GetChain(chainName)
	ctx := chain.GetContext()
	owner := chain.SenderAccount.GetAddress()
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	// Add all chain channel links in a single message
	msg := `{
		"modify_chain_channel_links": {
		  "operations": [
			{"operation": "set","source_chain": "chainB","destination_chain": "osmosis","channel_id": "channel-0"},
			{"operation": "set","source_chain": "osmosis","destination_chain": "chainB","channel_id": "channel-0"},
			{"operation": "set","source_chain": "chainB","destination_chain": "chainC","channel_id": "channel-1"},
			{"operation": "set","source_chain": "chainC","destination_chain": "chainB","channel_id": "channel-0"},
			{"operation": "set","source_chain": "osmosis","destination_chain": "chainC","channel_id": "channel-1"},
			{"operation": "set","source_chain": "chainC","destination_chain": "osmosis","channel_id": "channel-1"},
			{"operation": "set","source_chain": "osmosis","destination_chain": "chainB-cw20","channel_id": "channel-2"},
			{"operation": "set","source_chain": "chainB-cw20","destination_chain": "osmosis","channel_id": "channel-2"}
		  ]
		}
	  }
	  `
	_, err := contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

}

func (suite *HooksTestSuite) setAllPrefixesToOsmo(registryAddr sdk.AccAddress, chainName Chain) {
	chain := suite.GetChain(chainName)
	ctx := chain.GetContext()
	owner := chain.SenderAccount.GetAddress()
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	// Add all chain channel links in a single message
	msg := fmt.Sprintf(`{
		"modify_bech32_prefixes": {
		  "operations": [
			{"operation": "set", "chain_name": "osmosis", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainB", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainB-cw20", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainC", "prefix": "osmo"}
		  ]
		}
	  }
	  `)
	_, err := contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) setForwardingOnAllChains(registryAddr sdk.AccAddress) {
	suite.SetupAndTestPFM(ChainB, "chainB", registryAddr)
	suite.SetupAndTestPFM(ChainC, "chainC", registryAddr)
}

// modifyChainChannelLinks modifies the chain channel links in the crosschain registry utilizing set, remove, and change operations
func (suite *HooksTestSuite) modifyChainChannelLinks(registryAddr sdk.AccAddress, chainName Chain) {
	chain := suite.GetChain(chainName)
	ctx := chain.GetContext()
	owner := chain.SenderAccount.GetAddress()
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	msg := `{
		"modify_chain_channel_links": {
		  "operations": [
			{"operation": "remove","source_chain": "chainB","destination_chain": "chainC","channel_id": "channel-1"},
			{"operation": "set","source_chain": "chainD","destination_chain": "ChainC","channel_id": "channel-1"},
			{"operation": "remove","source_chain": "chainC","destination_chain": "chainB","channel_id": "channel-0"},
			{"operation": "set","source_chain": "ChainC","destination_chain": "chainD","channel_id": "channel-0"},
			{"operation": "change","source_chain": "chainB","destination_chain": "osmosis","new_source_chain": "chainD"},
			{"operation": "change","source_chain": "osmosis","destination_chain": "chainB","new_destination_chain": "chainD"}
		  ]
		}
	  }
	  `
	_, err := contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)
}

// modifyChainChannelLinks modifies the chain channel links in the crosschain registry utilizing set, remove, and change operations
func (suite *HooksTestSuite) setContractAlias(registryAddr sdk.AccAddress, contractAlias string, chainName Chain) {
	chain := suite.GetChain(chainName)
	ctx := chain.GetContext()
	owner := chain.SenderAccount.GetAddress()
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	msg := fmt.Sprintf(`{
		"modify_contract_alias": {
		  "operations": [
			{"operation": "set", "alias": "%s", "address": "%s"}
		  ]
		}
	  }
	  `, contractAlias, registryAddr)
	_, err := contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) TestCrosschainRegistry() {
	// Instantiate contract and set up three chains with funds sent between each
	registryAddr, _, token0CBA, CBAPath := suite.SetupCrosschainRegistry(ChainA)

	// Set the registry address to an alias
	contractAlias := "osmosis_registry_contract"
	suite.setContractAlias(registryAddr, contractAlias, ChainA)

	// Retrieve the registry address from the alias
	contractAddressFromAliasQuery := fmt.Sprintf(`{"get_address_from_alias": {"contract_alias": "%s"}}`, contractAlias)
	contractAddressFromAliasQueryResponse := suite.chainA.QueryContract(&suite.Suite, registryAddr, []byte(contractAddressFromAliasQuery))
	expectedAddressFromAliasQueryResponse := fmt.Sprintf(`{"address":"%s"}`, registryAddr)
	suite.Require().Equal(expectedAddressFromAliasQueryResponse, contractAddressFromAliasQueryResponse)

	// Add chain channel links to the registry on chain A
	suite.setChainChannelLinks(registryAddr, ChainA)

	// Query the denom trace of token0CB and check that it is as expected
	denomTraceQuery := fmt.Sprintf(`{"get_denom_trace": {"ibc_denom": "%s"}}`, token0CBA)
	denomTraceQueryResponse := suite.chainA.QueryContract(&suite.Suite, registryAddr, []byte(denomTraceQuery))
	expectedDenomTrace := fmt.Sprintf(`{"path":"%s","base_denom":"token0"}`, CBAPath)
	suite.Require().Equal(expectedDenomTrace, denomTraceQueryResponse)

	// Unwrap token0CB and check that it is as expected
	channelQuery := `{"get_channel_from_chain_pair": {"source_chain": "osmosis", "destination_chain": "chainB"}}`
	channelQueryResponse := suite.chainA.QueryContractJson(&suite.Suite, registryAddr, []byte(channelQuery))

	suite.Require().Equal("channel-0", channelQueryResponse.Str)

	// Remove, set, and change links on the registry on chain A
	suite.modifyChainChannelLinks(registryAddr, ChainA)

	_, err := suite.chainA.GetOsmosisApp().WasmKeeper.QuerySmart(suite.chainA.GetContext(), registryAddr, []byte(channelQuery))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "RegistryValue not found")

	// Unwrap token0CB and check that the path has changed
	channelQuery = `{"get_channel_from_chain_pair": {"source_chain": "osmosis", "destination_chain": "chainD"}}`
	channelQueryResponse = suite.chainA.QueryContractJson(&suite.Suite, registryAddr, []byte(channelQuery))
	suite.Require().Equal("channel-0", channelQueryResponse.Str)
}

func (suite *HooksTestSuite) TestUnwrapToken() {
	// Instantiate contract and set up three chains with funds sent between each
	registryAddr, _, token0CBA, _ := suite.SetupCrosschainRegistry(ChainA)
	suite.setChainChannelLinks(registryAddr, ChainA)
	suite.setAllPrefixesToOsmo(registryAddr, ChainA)
	suite.setForwardingOnAllChains(registryAddr)

	chain := suite.GetChain(ChainA)
	owner := chain.SenderAccount.GetAddress()
	receiver := chain.SenderAccounts[1].SenderAccount.GetAddress()
	osmosisApp := chain.GetOsmosisApp()

	// Check that the balances are correct: token0CB should be >100, token0CBA should be 0
	denomTrace0CA := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathAC.EndpointA.ChannelID, "token0"))
	token0CA := denomTrace0CA.IBCDenom()
	denomTrace0CB := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathBC.EndpointA.ChannelID, "token0"))
	token0CB := denomTrace0CB.IBCDenom()
	denomTrace0BA := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathAB.EndpointA.ChannelID, "token0"))
	token0BA := denomTrace0BA.IBCDenom()
	denomTrace0BC := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathBC.EndpointB.ChannelID, "token0"))
	token0BC := denomTrace0BC.IBCDenom()

	testCases := []struct {
		intoChain     Chain
		intoChainName string
		sentToken     string
		receivedToken string
		relayChain    []Direction
	}{
		{ChainA, "osmosis", token0CBA, token0CA, []Direction{AtoB, BtoC, CtoA}},
		{ChainB, "chainB", token0CBA, token0CB, []Direction{AtoB, BtoC, CtoB}},
		{ChainC, "chainC", token0BA, token0BC, []Direction{AtoB, BtoC}},
		{ChainC, "chainC", token0CBA, "token0", []Direction{AtoB, BtoC}},
	}

	for _, tc := range testCases {
		receiverChain := suite.GetChain(tc.intoChain)
		receiverApp := receiverChain.GetOsmosisApp()
		initialSenderBalance := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, tc.sentToken)
		sentAmount := osmomath.NewInt(100)
		suite.Require().Greater(initialSenderBalance.Amount.Int64(), sentAmount.Int64())
		initialReceiverBalance := receiverApp.BankKeeper.GetBalance(receiverChain.GetContext(), receiver, tc.receivedToken)
		suite.Require().Equal(osmomath.NewInt(0), initialReceiverBalance.Amount)

		msg := fmt.Sprintf(`{
		"unwrap_coin": {
			"receiver": "%s",
            "into_chain": "%s" 
    		}
	     }
	    `, receiver, tc.intoChainName)
		var exec sdk.Msg = &types.MsgExecuteContract{Contract: registryAddr.String(), Msg: []byte(msg), Sender: owner.String(), Funds: sdk.NewCoins(sdk.NewCoin(tc.sentToken, sentAmount))}
		res, err := chain.SendMsgsNoCheck(exec)
		suite.Require().NoError(err)

		for i, direction := range tc.relayChain {
			packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
			suite.Require().NoError(err)
			if i != len(tc.relayChain)-1 { // Only check the ack on the last hop
				res = suite.RelayPacketNoAck(packet, direction)
			} else {
				_, ack := suite.RelayPacket(packet, direction)
				suite.Require().Contains(string(ack), "result")
			}
		}

		// Check the balances are correct
		finalSenderBalance := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, tc.sentToken)
		suite.Require().Equal(initialSenderBalance.Amount.Sub(sentAmount), finalSenderBalance.Amount)
		finalReceiverBalance := receiverApp.BankKeeper.GetBalance(receiverChain.GetContext(), receiver, tc.receivedToken)
		suite.Require().Equal(sentAmount, finalReceiverBalance.Amount)
	}
}

func (suite *HooksTestSuite) TestCrosschainSwaps() {
	owner := suite.chainA.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)
	osmosisApp := suite.chainA.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	balanceSender := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, "token0")

	ctx := suite.chainA.GetContext()

	msg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s", "on_failed_delivery": "do_nothing"}}`,
		suite.chainB.SenderAccount.GetAddress(),
	)
	res, err := contractKeeper.Execute(ctx, crosschainAddr, owner, []byte(msg), sdk.NewCoins(sdk.NewCoin("token0", osmomath.NewInt(1000))))
	suite.Require().NoError(err)
	var responseJson map[string]interface{}
	err = json.Unmarshal(res, &responseJson)
	suite.Require().NoError(err)
	sentAmount, ok := responseJson["sent_amount"].(string)
	suite.Require().True(ok)
	suite.Require().Len(sentAmount, 3) // Not using exact amount in case calculations change

	denom, ok := responseJson["denom"].(string)
	suite.Require().True(ok)
	suite.Require().Equal(denom, "token1")

	channelID, ok := responseJson["channel_id"].(string)
	suite.Require().True(ok)
	suite.Require().Equal(channelID, "channel-0")

	receiver, ok := responseJson["receiver"].(string)
	suite.Require().True(ok)
	suite.Require().Equal(receiver, suite.chainB.SenderAccount.GetAddress().String())

	packetSequence, ok := responseJson["packet_sequence"].(float64)
	suite.Require().True(ok)
	suite.Require().Equal(packetSequence, 2.0)

	balanceSender2 := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, "token0")
	suite.Require().Equal(int64(1000), balanceSender.Amount.Sub(balanceSender2.Amount).Int64())
}

func (suite *HooksTestSuite) TestCrosschainSwapsViaIBCTest() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()
	denomTrace1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token1"))
	token1IBC := denomTrace1.IBCDenom()

	osmosisAppB := suite.chainB.GetOsmosisApp()
	balanceToken0 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	receiver := initializer
	balanceToken1 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), receiver, token1IBC)

	suite.Require().Equal(int64(0), balanceToken1.Amount.Int64())

	// Generate swap instructions for the contract
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s", "on_failed_delivery": "do_nothing", "next_memo":{}}}`,
		receiver,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, osmomath.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), "channel-0", msg)
	_, receiveResult, _, err := suite.FullSend(transferMsg, BtoA)

	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(receiveResult)

	// "Relay the packet" by executing the receive on chain B
	packet, err := ibctesting.ParsePacketFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)
	suite.RelayPacket(packet, AtoB)

	balanceToken0After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	suite.Require().Equal(int64(1000), balanceToken0.Amount.Sub(balanceToken0After.Amount).Int64())

	balanceToken1After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), receiver, token1IBC)
	suite.Require().Greater(balanceToken1After.Amount.Int64(), int64(0))
}

// This is a copy of the above to test bad acks. Lots of repetition here could be abstracted, but keeping as-is for
// now to avoid complexity
// The main difference between this test and the above one is that the receiver specified in the memo does not
// exist on chain B
func (suite *HooksTestSuite) TestCrosschainSwapsViaIBCBadAck() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()

	osmosisAppB := suite.chainB.GetOsmosisApp()
	balanceToken0 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	receiver := suite.chainB.SenderAccounts[5].SenderAccount.GetAddress().String()

	// Generate swap instructions for the contract. This will send correctly on chainA, but fail to be received on chainB
	recoverAddr := suite.chainA.SenderAccounts[8].SenderAccount.GetAddress()
	// we can no longer test by using a bad prefix as this is checked by the contracts. We will use a bad wasm memo to ensure the forward fails
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s","on_failed_delivery": {"local_recovery_addr": "%s"}, "next_memo": %s }}`,
		receiver,
		recoverAddr,
		`{"wasm": "bad wasm specifier"}`,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, osmomath.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), "channel-0", msg)
	_, receiveResult, _, err := suite.FullSend(transferMsg, BtoA)

	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(receiveResult)

	// "Relay the packet" by executing the receive on chain B
	packet, err := ibctesting.ParsePacketFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)
	receiveResult, ack2 := suite.RelayPacket(packet, AtoB)

	attrs := suite.ExtractAttributes(suite.FindEvent(receiveResult.GetEvents(), "ibccallbackerror-ibc-acknowledgement-error"))
	suite.Require().Contains(attrs["ibccallbackerror-error-context"], "wasm metadata is not a valid JSON map object")

	balanceToken0After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	suite.Require().Equal(int64(1000), balanceToken0.Amount.Sub(balanceToken0After.Amount).Int64())

	// The balance is stuck in the contract
	osmosisAppA := suite.chainA.GetOsmosisApp()
	balanceContract := osmosisAppA.BankKeeper.GetBalance(suite.chainA.GetContext(), crosschainAddr, "token1")
	suite.Require().Greater(balanceContract.Amount.Int64(), int64(0))

	// Send a second bad transfer from  with another recovery addr
	recoverAddr2 := suite.chainA.SenderAccounts[9].SenderAccount.GetAddress()
	swapMsg2 := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s","on_failed_delivery": {"local_recovery_addr": "%s"}, "next_memo": %s }}`,
		receiver,
		recoverAddr2,
		`{"wasm": "bad wasm specifier"}`,
	)
	// Generate full memo
	msg2 := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg2)
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, osmomath.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), "channel-0", msg2)
	_, receiveResult, _, err = suite.FullSend(transferMsg, BtoA)

	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(receiveResult)

	// "Relay the packet" by executing the receive on chain B
	packet, err = ibctesting.ParsePacketFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)
	_, ack2 = suite.RelayPacket(packet, AtoB)
	fmt.Println(string(ack2))

	balanceContract2 := osmosisAppA.BankKeeper.GetBalance(suite.chainA.GetContext(), crosschainAddr, "token1")
	suite.Require().Greater(balanceContract2.Amount.Int64(), balanceContract.Amount.Int64())

	// check that the contract knows this
	state := suite.chainA.QueryContract(
		&suite.Suite, crosschainAddr,
		[]byte(fmt.Sprintf(`{"recoverable": {"addr": "%s"}}`, recoverAddr)))
	suite.Require().Contains(state, "token1")
	suite.Require().Contains(state, `"sequence":3`)

	// Recover the stuck amount
	recoverMsg := `{"recover": {}}`
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisAppA.WasmKeeper)
	_, err = contractKeeper.Execute(suite.chainA.GetContext(), crosschainAddr, recoverAddr2, []byte(recoverMsg), sdk.NewCoins())
	suite.Require().NoError(err)

	balanceRecovery := osmosisAppA.BankKeeper.GetBalance(suite.chainA.GetContext(), recoverAddr2, "token1")
	suite.Require().Equal(balanceContract2.Sub(balanceContract).Amount.Int64(), balanceRecovery.Amount.Int64())

	// Calling recovery again should fail
	_, err = contractKeeper.Execute(suite.chainA.GetContext(), crosschainAddr, recoverAddr2, []byte(recoverMsg), sdk.NewCoins())
	suite.Require().Error(err)
}

// CrosschainSwapsViaIBCBadSwap tests that if the crosschain-swap fails, the tokens are returned to the sender
// This is very similar to the two tests above, but the swap is done incorrectly
func (suite *HooksTestSuite) TestCrosschainSwapsViaIBCBadSwap() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()
	denomTrace1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token1"))
	token1IBC := denomTrace1.IBCDenom()

	osmosisAppB := suite.chainB.GetOsmosisApp()
	balanceToken0 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	receiver := initializer
	balanceToken1 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), receiver, token1IBC)

	suite.Require().Equal(int64(0), balanceToken1.Amount.Int64())

	// Generate swap instructions for the contract. The min output amount here is too high, so the swap will fail
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"min_output_amount":"50000"},"receiver":"chainB/%s", "on_failed_delivery": "do_nothing"}}`,
		receiver,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, osmomath.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), "channel-0", msg)
	_, receiveResult, ack, err := suite.FullSend(transferMsg, BtoA)

	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(receiveResult)
	suite.Require().Contains(ack, "ABCI code: 6") // calculated amount is lesser than min output amount

	balanceToken0After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	suite.Require().Equal(balanceToken0.Amount, balanceToken0After.Amount)

	balanceToken1After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), receiver, token1IBC)
	suite.Require().Equal(balanceToken1After.Amount.Int64(), int64(0))
}

func (suite *HooksTestSuite) TestBadCrosschainSwapsNextMemoMessages() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(20000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()

	recoverAddr := suite.chainA.SenderAccounts[8].SenderAccount.GetAddress()
	receiver := initializer

	// next_memo is set to `%s` after the SprintF. It is then format replaced in each test case.
	innerMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s","on_failed_delivery": {"local_recovery_addr": "%s"},"next_memo":%%s}}`,
		receiver, // Note that this is the chain A account, which does not exist on chain B
		recoverAddr)

	testCases := []struct {
		memo    string
		expPass bool
	}{
		{fmt.Sprintf(innerMsg, `1`), false},
		{fmt.Sprintf(innerMsg, `""`), false},
		{fmt.Sprintf(innerMsg, `null`), true},
		{fmt.Sprintf(innerMsg, `"{\"ibc_callback\": \"something\"}"`), false},
		{fmt.Sprintf(innerMsg, `"{\"myKey\": \"myValue\"}"`), false}, // JSON memo should not be escaped
		{fmt.Sprintf(innerMsg, `"{}""`), true},                       // wasm not routed
		{fmt.Sprintf(innerMsg, `{}`), true},
		{fmt.Sprintf(innerMsg, `{"myKey": "myValue"}`), true},
	}

	for _, tc := range testCases {
		// Generate swap instructions for the contract. This will send correctly on chainA, but fail to be received on chainB
		// Generate full memo
		msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, tc.memo)
		// Send IBC transfer with the memo with crosschain-swap instructions
		fmt.Println(msg)
		transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, osmomath.NewInt(10)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), "channel-0", msg)
		_, _, ack, _ := suite.FullSend(transferMsg, BtoA)
		if tc.expPass {
			fmt.Println(ack)
			suite.Require().Contains(ack, "result", tc.memo)
		} else {
			suite.Require().Contains(ack, "error", tc.memo)
		}
	}
}

func (suite *HooksTestSuite) CreateIBCPoolOnChain(chainName Chain, denom1, denom2 string, amount1 osmomath.Int) uint64 {
	chain := suite.GetChain(chainName)
	acc1 := chain.SenderAccount.GetAddress()

	defaultFutureGovernor := ""

	// pool assets
	defaultFooAsset := balancer.PoolAsset{
		Weight: osmomath.NewInt(100),
		Token:  sdk.NewCoin(denom1, amount1),
	}
	defaultBarAsset := balancer.PoolAsset{
		Weight: osmomath.NewInt(100),
		Token:  sdk.NewCoin(denom2, osmomath.NewInt(defaultPoolAmount)),
	}

	poolAssets := []balancer.PoolAsset{defaultFooAsset, defaultBarAsset}

	poolParams := balancer.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.ZeroDec(),
	}
	msg := balancer.NewMsgCreateBalancerPool(acc1, poolParams, poolAssets, defaultFutureGovernor)
	poolId, err := chain.GetOsmosisApp().PoolManagerKeeper.CreatePool(chain.GetContext(), msg)
	suite.Require().NoError(err)

	_, err = chain.GetOsmosisApp().GAMMKeeper.GetPoolAndPoke(chain.GetContext(), poolId)
	suite.Require().NoError(err)
	return poolId
}

func (suite *HooksTestSuite) CreateIBCNativePoolOnChain(chainName Chain, denom string) uint64 {
	chain := suite.GetChain(chainName)
	bondDenom, err := chain.GetOsmosisApp().StakingKeeper.BondDenom(chain.GetContext())
	suite.Require().NoError(err)

	multiplier := osmomath.NewDec(20)

	uosmoAmount := gammtypes.InitPoolSharesSupply.ToLegacyDec().Mul(multiplier).RoundInt()

	return suite.CreateIBCPoolOnChain(chainName, bondDenom, denom, uosmoAmount)
}

func (suite *HooksTestSuite) SetupIBCRouteOnChain(swaprouterAddr, owner sdk.AccAddress, poolId uint64, chainName Chain, denom string) {
	chain := suite.GetChain(chainName)
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	msg := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"token0","pool_route":[{"pool_id":"%v","token_out_denom":"%s"},{"pool_id":"1","token_out_denom":"token0"}]}}`,
		denom, poolId, sdk.DefaultBondDenom)
	_, err := contractKeeper.Execute(chain.GetContext(), swaprouterAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	msg2 := fmt.Sprintf(`{"set_route":{"input_denom":"token0","output_denom":"%s","pool_route":[{"pool_id":"1","token_out_denom":"%s"},{"pool_id":"%v","token_out_denom":"%s"}]}}`,
		denom, sdk.DefaultBondDenom, poolId, denom)
	_, err = contractKeeper.Execute(chain.GetContext(), swaprouterAddr, owner, []byte(msg2), sdk.NewCoins())
	suite.Require().NoError(err)

	// Move forward one block
	chain.NextBlock()
	chain.Coordinator.IncrementTime()

	// Update both clients
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) SetupIBCSimpleRouteOnChain(swaprouterAddr, owner sdk.AccAddress, poolId uint64, chainName Chain, denom1, denom2 string) {
	chain := suite.GetChain(chainName)
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	msg := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"%s","pool_route":[{"pool_id":"%v","token_out_denom":"%s"}]}}`,
		denom1, denom2, poolId, denom2)
	_, err := contractKeeper.Execute(chain.GetContext(), swaprouterAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	msg2 := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"%s","pool_route":[{"pool_id":"%v","token_out_denom":"%s"}]}}`,
		denom2, denom1, poolId, denom1)
	_, err = contractKeeper.Execute(chain.GetContext(), swaprouterAddr, owner, []byte(msg2), sdk.NewCoins())
	suite.Require().NoError(err)

	// Move forward one block
	chain.NextBlock()
	chain.Coordinator.IncrementTime()

	// Update both clients
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
}

// TestCrosschainForwardWithMemo tests the that the next_memo field is correctly forwarded to the other chain on the IBC transfer.
// The second chain also has crosschain swaps setup and will execute a crosschain swap on receiving the response
func (suite *HooksTestSuite) TestCrosschainForwardWithMemo() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	receiver := suite.chainA.SenderAccounts[5].SenderAccount.GetAddress()

	_, crosschainAddrA := suite.SetupCrosschainSwaps(ChainA, true)
	swaprouterAddrB, crosschainAddrB := suite.SetupCrosschainSwaps(ChainB, false)
	// Send some token0 and token1 tokens to B so that there are ibc token0 to send to A and crosschain-swap, and token1 to create the pool
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(500000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)
	transferMsg1 := NewMsgTransfer(sdk.NewCoin("token1", osmomath.NewInt(500000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err = suite.FullSend(transferMsg1, AtoB)
	suite.Require().NoError(err)
	denom := suite.GetIBCDenom(ChainA, ChainB, "token1")
	poolId := suite.CreateIBCNativePoolOnChain(ChainB, denom)
	suite.SetupIBCRouteOnChain(swaprouterAddrB, suite.chainB.SenderAccount.GetAddress(), poolId, ChainB, denom)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()

	balanceToken0IBCBefore := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), receiver, token0IBC)
	fmt.Println("receiver now has: ", balanceToken0IBCBefore)
	suite.Require().Equal(int64(0), balanceToken0IBCBefore.Amount.Int64())

	// suite.Require().Equal(int64(0), balanceToken1.Amount.Int64())

	// Generate swap instructions for the contract
	//
	// Note: Both chains think of themselves as "osmosis" and the other as "chainB". That is, the registry
	// contracts on each test chain are not in sync. That's ok for this test, but a bit confusing.
	//
	// There is still an open question about how to handle verification and
	// forwarding if the user has manually specified the channel and/or memo that may
	// be relevant here
	nextMemo := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"osmosis_swap":{"output_denom":"token0","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"ibc:channel-0/%s", "on_failed_delivery": "do_nothing"}}}}`,
		crosschainAddrB,
		receiver,
	)
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"ibc:channel-0/%s", "on_failed_delivery": "do_nothing", "next_memo": %s}}`,
		crosschainAddrB,
		nextMemo,
	)
	fmt.Println(swapMsg)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddrA, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, osmomath.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddrA.String(), "channel-0", msg)
	_, receiveResult, _, err := suite.FullSend(transferMsg, BtoA)

	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(receiveResult)

	// "Relay the packet" by executing the receive on chain B
	packet, err := ibctesting.ParsePacketFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)
	relayResult, _ := suite.RelayPacket(packet, AtoB)

	// Now that chain B has processed it, it should be sending a message to chain A. Relay the response
	packet2, err := ibctesting.ParsePacketFromEvents(relayResult.GetEvents())
	suite.Require().NoError(err)
	suite.RelayPacket(packet2, BtoA)

	balanceToken0IBCAfter := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), receiver, token0IBC)
	fmt.Println("receiver now has: ", balanceToken0IBCAfter)
	suite.Require().Greater(balanceToken0IBCAfter.Amount.Int64(), int64(0))
}

// This test will send tokens from A->C->B so that B has token0CBA and then
// execute a crosschain swap on A to obtain B's native token0
func (suite *HooksTestSuite) TestCrosschainSwapsViaIBCMultiHop() {
	accountA := suite.chainA.SenderAccount.GetAddress()
	accountB := suite.chainB.SenderAccount.GetAddress()
	accountC := suite.chainC.SenderAccount.GetAddress()

	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)

	// Send A's token0 all the way to B (A->C->B)
	transferMsg := NewMsgTransfer(
		sdk.NewCoin("token0", osmomath.NewInt(2000)),
		accountA.String(),
		accountC.String(),
		suite.pathAC.EndpointA.ChannelID,
		"",
	)
	_, _, _, err := suite.FullSend(transferMsg, AtoC)
	suite.Require().NoError(err)

	token0AC := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathAC.EndpointB.ChannelID, "token0")).IBCDenom()
	transferMsg = NewMsgTransfer(
		sdk.NewCoin(token0AC, osmomath.NewInt(2000)),
		accountC.String(),
		accountB.String(),
		suite.pathBC.EndpointB.ChannelID,
		"",
	)
	_, _, _, err = suite.FullSend(transferMsg, CtoB)
	suite.Require().NoError(err)

	// Calculate the names of the tokens when sent via IBC
	ACBPath := fmt.Sprintf("transfer/%s/transfer/%s", suite.pathAC.EndpointB.ChannelID, suite.pathBC.EndpointA.ChannelID)
	denomTrace0ACB := transfertypes.DenomTrace{Path: ACBPath, BaseDenom: "token0"}
	token0ACB := denomTrace0ACB.IBCDenom()

	denomTrace1AB := transfertypes.DenomTrace{Path: fmt.Sprintf("transfer/%s", suite.pathAB.EndpointB.ChannelID), BaseDenom: "token1"}
	token1AB := denomTrace1AB.IBCDenom()

	osmosisAppB := suite.chainB.GetOsmosisApp()
	balanceToken0 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), accountB, token0ACB)
	balanceToken1 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), accountB, token1AB)

	suite.Require().Equal(int64(0), balanceToken1.Amount.Int64())

	// Generate swap instructions for the contract
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s", "on_failed_delivery": "do_nothing", "next_memo":{}}}`,
		accountB,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0ACB, osmomath.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), "channel-0", msg)
	_, res, _, err := suite.FullSend(transferMsg, BtoA)
	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	// Now that chain A has processed it, it should be sending a new packet to chain C with the proper forward memo
	// First to B
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	res = suite.RelayPacketNoAck(packet, AtoB)

	// B forwards to C
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	res = suite.RelayPacketNoAck(packet, BtoC)

	// C forwards to A
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	res = suite.RelayPacketNoAck(packet, CtoA)

	// Now the swwap can actually execute on A via the callback and generate a new packet with the swapped token to B
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	_ = suite.RelayPacketNoAck(packet, AtoB)

	balanceToken0After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), accountB, token0ACB)
	suite.Require().Equal(int64(1000), balanceToken0.Amount.Sub(balanceToken0After.Amount).Int64())

	balanceToken1After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), accountB, token1AB)
	suite.Require().Greater(balanceToken1After.Amount.Int64(), int64(0))
}

// simple transfer
func (suite *HooksTestSuite) SimpleNativeTransfer(token string, amount osmomath.Int, path []Chain) string {
	prev := path[0]
	prevPrefix := ""
	denom := token
	for i, path := range path {
		if i == 0 {
			continue
		}
		fromChain := prev
		toChain := path
		transferMsg := NewMsgTransfer(
			sdk.NewCoin(denom, amount),
			suite.GetChain(prev).SenderAccount.GetAddress().String(),
			suite.GetChain(path).SenderAccount.GetAddress().String(),
			suite.GetSenderChannel(fromChain, toChain),
			"",
		)
		_, _, _, err := suite.FullSend(transferMsg, suite.GetDirection(fromChain, toChain))
		suite.Require().NoError(err)
		receiveChannel := suite.GetSenderChannel(toChain, fromChain)
		prevPrefix += "/" + strings.TrimRight(transfertypes.GetDenomPrefix("transfer", receiveChannel), "/")
		prevPrefix = strings.TrimLeft(prevPrefix, "/")
		denom = transfertypes.DenomTrace{Path: prevPrefix, BaseDenom: token}.IBCDenom()
		prev = toChain
	}
	return denom
}

func (suite *HooksTestSuite) GetPath(chain1, chain2 Chain) string {
	return fmt.Sprintf("transfer/%s", suite.GetReceiverChannel(chain1, chain2))
}

func (suite *HooksTestSuite) GetIBCDenom(a Chain, b Chain, denom string) string {
	return transfertypes.DenomTrace{Path: suite.GetPath(a, b), BaseDenom: denom}.IBCDenom()
}

type ChainActorDefinition struct {
	Chain
	name    string
	address sdk.AccAddress
}

func (suite *HooksTestSuite) TestMultiHopXCS() {
	accountB := suite.chainB.SenderAccount.GetAddress()

	swapRouterAddr, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)

	sendAmount := osmomath.NewInt(100)

	actorChainB := ChainActorDefinition{
		Chain:   ChainB,
		name:    "chainB",
		address: accountB,
	}

	var customRoute string

	testCases := []struct {
		name              string
		sender            ChainActorDefinition
		swapFor           string
		receiver          ChainActorDefinition
		receivedToken     string
		setupInitialToken func() string
		relayChain        []Direction
		requireAck        []bool
	}{
		{
			name:          "A's token0 in B wrapped as B.C.A, send to A for unwrapping and then swap for A's token1, receive into B",
			sender:        actorChainB,
			swapFor:       "token1",
			receiver:      actorChainB,
			receivedToken: transfertypes.DenomTrace{Path: suite.GetPath(ChainA, ChainB), BaseDenom: "token1"}.IBCDenom(),
			setupInitialToken: func() string {
				return suite.SimpleNativeTransfer("token0", sendAmount, []Chain{ChainA, ChainC, ChainB})
			},
			relayChain: []Direction{AtoB, BtoC, CtoA, AtoB},
			requireAck: []bool{false, false, true, true},
		},

		{
			name:          "C's token0 in B wrapped as B.C, send to A for unwrapping and then swap for A's token0, receive into B",
			sender:        actorChainB,
			swapFor:       "token0",
			receiver:      actorChainB,
			receivedToken: suite.GetIBCDenom(ChainA, ChainB, "token0"),
			setupInitialToken: func() string {
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(defaultPoolAmount), []Chain{ChainC, ChainA})

				denom := suite.GetIBCDenom(ChainC, ChainA, "token0")
				poolId := suite.CreateIBCNativePoolOnChain(ChainA, denom)
				suite.SetupIBCRouteOnChain(swapRouterAddr, suite.chainA.SenderAccount.GetAddress(), poolId, ChainA, denom)

				return suite.SimpleNativeTransfer("token0", sendAmount, []Chain{ChainC, ChainB})
			},
			relayChain: []Direction{AtoB, BtoC, CtoA, AtoB},
			requireAck: []bool{false, false, true, true},
		},

		{
			name: "Native to OsmoNative into same chain",
			// This is currently failing when running all tests together but not individually. TODO: Figure out why
			sender:        actorChainB,
			swapFor:       "token0",
			receiver:      actorChainB,
			receivedToken: transfertypes.DenomTrace{Path: suite.GetPath(ChainA, ChainB), BaseDenom: "token0"}.IBCDenom(),
			setupInitialToken: func() string {
				suite.SimpleNativeTransfer("token1", osmomath.NewInt(defaultPoolAmount), []Chain{ChainB, ChainA})

				suite.SimpleNativeTransfer("token1", osmomath.NewInt(10000), []Chain{ChainB, ChainA})
				denom := suite.GetIBCDenom(ChainB, ChainA, "token1")
				poolId := suite.CreateIBCNativePoolOnChain(ChainA, denom)
				suite.SetupIBCRouteOnChain(swapRouterAddr, suite.chainA.SenderAccount.GetAddress(), poolId, ChainA, denom)

				return "token1"
			},
			relayChain: []Direction{AtoB},
			requireAck: []bool{true},
		},

		{
			name:          "OsmoNative to Native into same chain",
			sender:        actorChainB,
			swapFor:       transfertypes.DenomTrace{Path: suite.GetPath(ChainA, ChainB), BaseDenom: "token1"}.IBCDenom(),
			receiver:      actorChainB,
			receivedToken: "token1",
			setupInitialToken: func() string {
				// Send ibc token to chainB
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(500), []Chain{ChainA, ChainB})

				// Setup pool
				suite.SimpleNativeTransfer("token1", osmomath.NewInt(defaultPoolAmount), []Chain{ChainB, ChainA})
				denom := suite.GetIBCDenom(ChainB, ChainA, "token1")
				poolId := suite.CreateIBCNativePoolOnChain(ChainA, denom)
				suite.SetupIBCRouteOnChain(swapRouterAddr, suite.chainA.SenderAccount.GetAddress(), poolId, ChainA, denom)

				return suite.GetIBCDenom(ChainB, ChainA, "token0")
			},
			relayChain: []Direction{AtoB},
			requireAck: []bool{true},
		},

		{
			name:          "Swap two IBC'd tokens",
			sender:        actorChainB,
			swapFor:       suite.GetIBCDenom(ChainC, ChainA, "token0"),
			receiver:      actorChainB,
			receivedToken: suite.GetIBCDenom(ChainC, ChainB, "token0"),
			setupInitialToken: func() string {
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainC, ChainA})
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainB, ChainA})

				token0BA := suite.GetIBCDenom(ChainB, ChainA, "token0")
				token0CA := suite.GetIBCDenom(ChainC, ChainA, "token0")

				// Setup pool
				poolId := suite.CreateIBCPoolOnChain(ChainA, token0BA, token0CA, osmomath.NewInt(defaultPoolAmount))
				suite.SetupIBCSimpleRouteOnChain(swapRouterAddr, suite.chainA.SenderAccount.GetAddress(), poolId, ChainA, token0BA, token0CA)

				return "token0"
			},
			relayChain: []Direction{AtoC, CtoB},
			requireAck: []bool{false, true},
		},

		{
			name:          "Swap two IBC'd tokens with specified route",
			sender:        actorChainB,
			swapFor:       suite.GetIBCDenom(ChainC, ChainA, "token0"),
			receiver:      actorChainB,
			receivedToken: suite.GetIBCDenom(ChainC, ChainB, "token0"),
			setupInitialToken: func() string {
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainC, ChainA})
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainB, ChainA})

				token0BA := suite.GetIBCDenom(ChainB, ChainA, "token0")
				token0CA := suite.GetIBCDenom(ChainC, ChainA, "token0")

				// Setup pool
				poolId := suite.CreateIBCPoolOnChain(ChainA, token0BA, token0CA, osmomath.NewInt(defaultPoolAmount))

				customRoute = fmt.Sprintf(`[{"pool_id": "%d", "token_out_denom": "%s"}]`, poolId, token0CA)

				return "token0"
			},
			relayChain: []Direction{AtoC, CtoB},
			requireAck: []bool{false, true},
		},

		{
			name:    "Swap two IBC'd tokens with unwrapping before and after",
			sender:  actorChainB,
			swapFor: suite.GetIBCDenom(ChainB, ChainA, "token0"),
			receiver: ChainActorDefinition{
				Chain:   ChainC,
				name:    "chainC",
				address: accountB,
			},
			receivedToken: suite.GetIBCDenom(ChainB, ChainC, "token0"),
			setupInitialToken: func() string {
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainC, ChainA})
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainB, ChainA})
				suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainC, ChainB})

				token0BA := suite.GetIBCDenom(ChainB, ChainA, "token0")
				token0CB := suite.GetIBCDenom(ChainC, ChainB, "token0")
				token0CA := suite.GetIBCDenom(ChainC, ChainA, "token0")

				// Setup pool
				poolId := suite.CreateIBCPoolOnChain(ChainA, token0BA, token0CA, osmomath.NewInt(defaultPoolAmount))
				suite.SetupIBCSimpleRouteOnChain(swapRouterAddr, suite.chainA.SenderAccount.GetAddress(), poolId, ChainA, token0BA, token0CA)

				return token0CB
			},
			relayChain: []Direction{AtoB, BtoC, CtoA, AtoB, BtoC},
			requireAck: []bool{false, false, true, false, true},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			senderChain := suite.GetChain(tc.sender.Chain)
			receiverChain := suite.GetChain(tc.receiver.Chain)
			customRoute = "" // reset the custom route

			initialToken := tc.setupInitialToken()

			// Check that the balances are correct
			// check sender balance
			sentTokenBalance := senderChain.GetOsmosisApp().BankKeeper.GetBalance(senderChain.GetContext(), tc.sender.address, initialToken)
			fmt.Println("sentTokenBalance", sentTokenBalance)
			suite.Require().True(sentTokenBalance.Amount.GTE(sendAmount))
			// get receiver balance
			receivedTokenBalance := receiverChain.GetOsmosisApp().BankKeeper.GetBalance(receiverChain.GetContext(), tc.receiver.address, tc.receivedToken)

			// Generate swap instructions for the contract
			var extra string
			if customRoute != "" {
				extra = fmt.Sprintf(`,"route": %s`, customRoute)
			}

			swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"%s","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s/%s", "on_failed_delivery": "do_nothing", "next_memo":{}%s}}`,
				tc.swapFor, tc.receiver.name, tc.receiver.address, extra,
			)
			// Generate full memo
			msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
			// Send IBC transfer with the memo with crosschain-swap instructions
			transferMsg := NewMsgTransfer(sdk.NewCoin(initialToken, sendAmount), tc.sender.address.String(), crosschainAddr.String(), suite.GetSenderChannel(tc.sender.Chain, ChainA), msg)
			_, res, _, err := suite.FullSend(transferMsg, BtoA)
			// We use the receive result here because the receive adds another packet to be sent back
			suite.Require().NoError(err)
			suite.Require().NotNil(res)

			var ack []byte
			for i, direction := range tc.relayChain {
				packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
				suite.Require().NoError(err)
				if tc.requireAck[i] {
					res, ack = suite.RelayPacket(packet, direction)
					suite.Require().Contains(string(ack), "result")
				} else {
					res = suite.RelayPacketNoAck(packet, direction)
				}
			}

			sentTokenAfter := senderChain.GetOsmosisApp().BankKeeper.GetBalance(senderChain.GetContext(), tc.sender.address, initialToken)
			suite.Require().Equal(sendAmount.Int64(), sentTokenBalance.Amount.Sub(sentTokenAfter.Amount).Int64())

			fmt.Println(receiverChain.GetOsmosisApp().BankKeeper.GetAllBalances(receiverChain.GetContext(), tc.receiver.address))

			receivedTokenAfter := receiverChain.GetOsmosisApp().BankKeeper.GetBalance(receiverChain.GetContext(), tc.receiver.address, tc.receivedToken)
			suite.Require().True(receivedTokenAfter.Amount.GT(receivedTokenBalance.Amount))
		})
	}
}

// This sends a packet (setup to use PFM) through a path and ensures acks are returned to the sender
func (suite *HooksTestSuite) SendAndAckPacketThroughPath(packetPath []Direction, initialPacket channeltypes.Packet) {
	var res *abci.ExecTxResult
	var err error

	packetStack := make([]channeltypes.Packet, 0)
	packet := initialPacket

	for i, direction := range packetPath {
		packetStack = append(packetStack, packet)
		suite.Require().NoError(err)
		res = suite.RelayPacketNoAck(packet, direction)
		if i != len(packetPath)-1 {
			packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
			suite.Require().NoError(err)
		}

		senderEndpoint, receiverEndpoint := suite.GetEndpoints(direction)
		receiverEndpoint.Chain.NextBlock()
		err = receiverEndpoint.UpdateClient()
		suite.Require().NoError(err)
		senderEndpoint.Chain.NextBlock()
		err = senderEndpoint.UpdateClient()
		suite.Require().NoError(err)
	}
	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	for i := range packetPath {
		packet = packetStack[len(packetStack)-i-1]
		direction := packetPath[len(packetPath)-i-1]
		// sender Acknowledges
		senderEndpoint, receiverEndpoint := suite.GetEndpoints(direction)

		senderEndpoint.Chain.NextBlock()
		err = senderEndpoint.UpdateClient()
		suite.Require().NoError(err)
		receiverEndpoint.Chain.NextBlock()
		err = receiverEndpoint.UpdateClient()
		suite.Require().NoError(err)
		err = senderEndpoint.AcknowledgePacket(packet, ack)
		suite.Require().NoError(err)
	}

}

func (suite *HooksTestSuite) TestSwapErrorAfterPreSwapUnwind() {
	// setup
	accountB := suite.chainB.SenderAccount.GetAddress()

	swapRouterAddr, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)

	sendAmount := osmomath.NewInt(100)

	sender := ChainActorDefinition{
		Chain:   ChainB,
		name:    "chainB",
		address: accountB,
	}
	swapFor := suite.GetIBCDenom(ChainB, ChainA, "token0")
	receiver := ChainActorDefinition{
		Chain:   ChainC,
		name:    "chainC",
		address: accountB,
	}

	//setup initial tokens
	suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainC, ChainA})
	suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainB, ChainA})
	suite.SimpleNativeTransfer("token0", osmomath.NewInt(12000000), []Chain{ChainC, ChainB})

	token0BA := suite.GetIBCDenom(ChainB, ChainA, "token0")
	token0CB := suite.GetIBCDenom(ChainC, ChainB, "token0")
	token0CA := suite.GetIBCDenom(ChainC, ChainA, "token0")

	// Setup pool
	poolId := suite.CreateIBCPoolOnChain(ChainA, token0BA, token0CA, osmomath.NewInt(defaultPoolAmount))
	suite.SetupIBCSimpleRouteOnChain(swapRouterAddr, suite.chainA.SenderAccount.GetAddress(), poolId, ChainA, token0BA, token0CA)

	initialToken := token0CB

	// execute
	// In this test, we will send chainC's native from chain B to chain A. The XCS contract will then send it back
	// for unwinding before executing the swap. The swap is setup to fail at this step. The contract should then
	// allow users to recover these tokens when the ack of the unwind is received.
	senderChain := suite.GetChain(sender.Chain)

	sentTokenBalance := senderChain.GetOsmosisApp().BankKeeper.GetBalance(senderChain.GetContext(), sender.address, initialToken)
	suite.Require().True(sentTokenBalance.Amount.GTE(sendAmount))

	// Generate swap instructions for the contract
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"%s","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"0"}},"receiver":"%s/%s", "on_failed_delivery": {"local_recovery_addr": "%s"}, "next_memo":{}}}`,
		swapFor, receiver.name, receiver.address, sender.address,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg := NewMsgTransfer(sdk.NewCoin(initialToken, sendAmount), sender.address.String(), crosschainAddr.String(), suite.GetSenderChannel(sender.Chain, ChainA), msg)
	_, res, _, err := suite.FullSend(transferMsg, BtoA)
	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.SendAndAckPacketThroughPath([]Direction{AtoB, BtoC, CtoA}, packet)

	recoverableQuery := fmt.Sprintf(`{"recoverable": {"addr": "%s"}}`, sender.address)
	recoverableQueryResponse := suite.chainA.QueryContractJson(&suite.Suite, crosschainAddr, []byte(recoverableQuery))
	suite.Require().Equal(1, len(recoverableQueryResponse.Array()))
	suite.Require().Equal(sender.address.String(), recoverableQueryResponse.Get("0.recovery_addr").String())
	suite.Require().Equal(sendAmount.String(), recoverableQueryResponse.Get("0.amount").String())

}

func (suite *HooksTestSuite) ExecuteOutpostSwap(initializer, receiverAddr sdk.AccAddress, receiver string) {
	// Setup
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)
	// Store and instantiate the outpost on chainB
	suite.chainB.StoreContractCode(&suite.Suite, "./bytecode/outpost.wasm")
	outpostAddr := suite.chainB.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"crosschain_swaps_contract": "%s", "osmosis_channel": "channel-0"}`, crosschainAddr), 1)

	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", osmomath.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "channel-0", "")
	_, _, _, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()
	denomTrace1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token1"))
	token1IBC := denomTrace1.IBCDenom()

	osmosisAppB := suite.chainB.GetOsmosisApp()
	balanceToken0 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	balanceToken1 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, token1IBC)

	suite.Require().Equal(int64(0), balanceToken1.Amount.Int64())

	// Generate swap instructions for the contract
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s", "on_failed_delivery": "do_nothing"}}`,
		receiver,
	)

	// Call the outpost
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisAppB.WasmKeeper)
	ctxB := suite.chainB.GetContext()
	_, err = contractKeeper.Execute(ctxB, outpostAddr, initializer, []byte(swapMsg), sdk.NewCoins(sdk.NewCoin(token0IBC, osmomath.NewInt(1000))))
	suite.Require().NoError(err)
	suite.chainB.NextBlock()
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// "Relay the packet" by executing the receive on chain A
	packet, err := ibctesting.ParsePacketFromEvents(ctxB.EventManager().Events().ToABCIEvents())
	suite.Require().NoError(err)
	receiveResult, _ := suite.RelayPacket(packet, BtoA)

	suite.chainA.NextBlock()
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	// The chain A should execute the cross chain swaps and add a new packet
	// "Relay the packet" by executing the receive on chain B
	packet, err = ibctesting.ParsePacketFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)
	suite.RelayPacket(packet, AtoB)

	// The sender has 1000token0IBC less
	balanceToken0After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	suite.Require().Equal(int64(1000), balanceToken0.Amount.Sub(balanceToken0After.Amount).Int64())

	// But the receiver now has some token1IBC
	balanceToken1After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, token1IBC)
	// fmt.Println("receiver now has: ", balanceToken1After)
	suite.Require().Greater(balanceToken1After.Amount.Int64(), int64(0))
}

func (suite *HooksTestSuite) TestOutpostSimplified() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	suite.ExecuteOutpostSwap(initializer, initializer, fmt.Sprintf(`chainB/%s`, initializer.String()))
}

func (suite *HooksTestSuite) TestOutpostExplicit() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	suite.ExecuteOutpostSwap(initializer, initializer, fmt.Sprintf(`ibc:channel-0/%s`, initializer.String()))
}
