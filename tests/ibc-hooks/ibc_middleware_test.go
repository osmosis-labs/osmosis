package ibc_hooks_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/types"

	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	minttypes "github.com/osmosis-labs/osmosis/v15/x/mint/types"
	txfeetypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"

	"github.com/osmosis-labs/osmosis/v15/tests/osmosisibctesting"

	"github.com/osmosis-labs/osmosis/v15/tests/ibc-hooks/testutils"
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
	pathBA *ibctesting.Path
	pathCA *ibctesting.Path
	pathCB *ibctesting.Path
}

var oldConsensusMinFee = txfeetypes.ConsensusMinFee

func (suite *HooksTestSuite) SetupTest() {
	// TODO: This needs to get removed. Waiting on https://github.com/cosmos/ibc-go/issues/3123
	txfeetypes.ConsensusMinFee = sdk.ZeroDec()
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

// TODO: This needs to get removed. Waiting on https://github.com/cosmos/ibc-go/issues/3123
func (suite *HooksTestSuite) TearDownSuite() {
	txfeetypes.ConsensusMinFee = oldConsensusMinFee
}

func TestIBCHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

// ToDo: Move this to osmosistesting to avoid repetition
func NewTransferPath(chainA, chainB *osmosisibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

func (suite *HooksTestSuite) TestOnRecvPacketHooks() {
	var (
		trace    transfertypes.DenomTrace
		amount   sdk.Int
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

			amount = sdk.NewInt(100) // must be explicitly changed in malleate
			seq := uint64(1)

			trace = transfertypes.ParseDenomTrace(sdk.DefaultBondDenom)

			// send coin from chainA to chainB
			transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.NewCoin(trace.IBCDenom(), amount), suite.chainA.SenderAccount.GetAddress().String(), receiver, clienttypes.NewHeight(1, 110), 0)
			_, err := suite.chainA.SendMsgs(transferMsg)
			suite.Require().NoError(err) // message committed

			tc.malleate(&status)

			data := transfertypes.NewFungibleTokenPacketData(trace.GetFullDenomPath(), amount.String(), suite.chainA.SenderAccount.GetAddress().String(), receiver)
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
		clienttypes.NewHeight(0, 100),
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

	err := suite.chainB.GetOsmosisApp().HooksICS4Wrapper.SendPacket(
		suite.chainB.GetContext(), channelCap, packet)
	suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)

	// Update both clients
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// recv in chain a
	res, err := suite.pathAB.EndpointA.RecvPacketWithResult(packet)

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
	suite.Require().Equal(sdk.NewInt(0), balance.Amount)

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
	suite.Require().Equal(sdk.NewInt(1), balance.Amount)
}

// If the wasm call wails, the contract acknowledgement should be an error and the funds returned
func (suite *HooksTestSuite) TestFundsAreReturnedOnFailedContractExec() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/echo.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, "{}", 1)

	// Check that the contract has no funds
	localDenom := osmoutils.MustExtractDenomFromPacketOnRecv(suite.makeMockPacket("", "", 0))
	balance := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), addr, localDenom)
	suite.Require().Equal(sdk.NewInt(0), balance.Amount)

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
	suite.Require().Equal(sdk.NewInt(0), balance.Amount)
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
	suite.Require().Equal(sdk.NewInt(0), balance.Amount)

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
	suite.Require().Equal(sdk.NewInt(2), balance.Amount)
}

// custom MsgTransfer constructor that supports Memo
func NewMsgTransfer(
	token sdk.Coin, sender, receiver string, memo string,
) *transfertypes.MsgTransfer {
	return &transfertypes.MsgTransfer{
		SourcePort:       "transfer",
		SourceChannel:    "channel-0",
		Token:            token,
		Sender:           sender,
		Receiver:         receiver,
		TimeoutHeight:    clienttypes.NewHeight(0, 100),
		TimeoutTimestamp: 0,
		Memo:             memo,
	}
}

type Direction int64

const (
	AtoB Direction = iota
	BtoA
	AtoC
	CtoA
	BtoC
	CtoB
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
	default:
		panic("invalid direction")
	}
	return sender, receiver
}

func (suite *HooksTestSuite) RelayPacket(packet channeltypes.Packet, direction Direction) (*sdk.Result, []byte) {
	sender, receiver := suite.GetEndpoints(direction)

	err := receiver.UpdateClient()
	suite.Require().NoError(err)

	// receiver Receives
	receiveResult, err := receiver.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)

	// sender Acknowledges
	err = sender.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)

	err = sender.UpdateClient()
	suite.Require().NoError(err)
	err = receiver.UpdateClient()
	suite.Require().NoError(err)

	return receiveResult, ack
}

func (suite *HooksTestSuite) RelayPacketNoAck(packet channeltypes.Packet, direction Direction) *sdk.Result {
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

func (suite *HooksTestSuite) FullSend(msg sdk.Msg, direction Direction) (*sdk.Result, *sdk.Result, string, error) {
	var sender *osmosisibctesting.TestChain
	switch direction {
	case AtoB:
		sender = suite.chainA
	case BtoA:
		sender = suite.chainB
	case CtoB:
		sender = suite.chainC
	}
	sendResult, err := sender.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	suite.Require().NoError(err)

	receiveResult, ack := suite.RelayPacket(packet, direction)

	return sendResult, receiveResult, string(ack), err
}

func (suite *HooksTestSuite) TestAcks() {
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/counter.wasm")
	addr := suite.chainA.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)

	// Generate swap instructions for the contract
	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), addr.String(), callbackMemo)
	suite.FullSend(transferMsg, AtoB)

	// The test contract will increment the counter for itself every time it receives an ack
	state := suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":1}`, state)

	suite.FullSend(transferMsg, AtoB)
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
	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), addr.String(), callbackMemo)
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
	transferMsg := NewMsgTransfer(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)), suite.chainA.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), "")
	_, _, ack, err := suite.FullSend(transferMsg, AtoB)
	suite.Require().NoError(err)
	suite.Require().Contains(ack, "result")
}

type Chain int64

const (
	ChainA Chain = iota
	ChainB
	ChainC
)

func (suite *HooksTestSuite) GetChain(name Chain) *osmosisibctesting.TestChain {
	if name == ChainA {
		return suite.chainA
	} else {
		return suite.chainB
	}
}

// This is a copy of the SetupGammPoolsWithBondDenomMultiplier from the  test helpers, but using chainA instead of the default
func (suite *HooksTestSuite) SetupPools(chainName Chain, multipliers []sdk.Dec) []gammtypes.CFMMPoolI {
	chain := suite.GetChain(chainName)
	acc1 := chain.SenderAccount.GetAddress()
	bondDenom := chain.GetOsmosisApp().StakingKeeper.BondDenom(chain.GetContext())

	pools := []gammtypes.CFMMPoolI{}
	for index, multiplier := range multipliers {
		token := fmt.Sprintf("token%d", index)
		uosmoAmount := gammtypes.InitPoolSharesSupply.ToDec().Mul(multiplier).RoundInt()

		var (
			defaultFutureGovernor = ""

			// pool assets
			defaultFooAsset = balancer.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(bondDenom, uosmoAmount),
			}
			defaultBarAsset = balancer.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(token, sdk.NewInt(10000)),
			}

			poolAssets = []balancer.PoolAsset{defaultFooAsset, defaultBarAsset}
		)

		poolParams := balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
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

func (suite *HooksTestSuite) SetupCrosschainSwaps(chainName Chain) (sdk.AccAddress, sdk.AccAddress) {
	chain := suite.GetChain(chainName)
	owner := chain.SenderAccount.GetAddress()

	// Fund the account with some uosmo and some stake
	bankKeeper := chain.GetOsmosisApp().BankKeeper
	i, ok := sdk.NewIntFromString("20000000000000000000000")
	suite.Require().True(ok)
	amounts := sdk.NewCoins(sdk.NewCoin("uosmo", i), sdk.NewCoin("stake", i), sdk.NewCoin("token0", i), sdk.NewCoin("token1", i))
	err := bankKeeper.MintCoins(chain.GetContext(), minttypes.ModuleName, amounts)
	suite.Require().NoError(err)
	err = bankKeeper.SendCoinsFromModuleToAccount(chain.GetContext(), minttypes.ModuleName, owner, amounts)
	suite.Require().NoError(err)

	suite.SetupPools(chainName, []sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// Setup contract
	chain.StoreContractCode(&suite.Suite, "./bytecode/swaprouter.wasm")
	swaprouterAddr := chain.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"owner": "%s"}`, owner), 1)
	chain.StoreContractCode(&suite.Suite, "./bytecode/crosschain_swaps.wasm")

	// Configuring two prefixes for the same channel here. This is so that we can test bad acks when the receiver can't handle the receiving addr
	channels := `[["osmo", "channel-0"],["juno", "channel-0"]]`
	crosschainAddr := chain.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"swap_contract": "%s", "channels": %s, "governor": "%s"}`, swaprouterAddr, channels, owner),
		2)

	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	ctx := chain.GetContext()

	// ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins
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
	i, ok := sdk.NewIntFromString("20000000000000000000000")
	suite.Require().True(ok)
	amounts := sdk.NewCoins(sdk.NewCoin("uosmo", i), sdk.NewCoin("stake", i), sdk.NewCoin("token0", i), sdk.NewCoin("token1", i))
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
	suite.SetupPools(chainName, []sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/crosschain_registry.wasm")
	registryAddr := chain.InstantiateContract(&suite.Suite, fmt.Sprintf(`{"owner": "%s"}`, owner), 1)
	_, err := sdk.Bech32ifyAddressBytes("osmo", registryAddr)
	suite.Require().NoError(err)

	// Send some token0 tokens from C to B
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), suite.chainC.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), "")
	suite.FullSend(transferMsg, CtoB)

	// Send some token0 tokens from B to A
	denomTrace0CB := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathBC.EndpointA.ChannelID, "token0"))
	token0CB := denomTrace0CB.IBCDenom()
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0CB, sdk.NewInt(2000)), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), "")
	suite.FullSend(transferMsg, BtoA)

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
			{"operation": "set","source_chain": "chainC","destination_chain": "osmosis","channel_id": "channel-1"}
		  ]
		}
	  }
	  `
	_, err := contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)
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
			{"operation": "set","source_chain": "ChainB","destination_chain": "ChainC","channel_id": "channel-1"},
			{"operation": "remove","source_chain": "chainC","destination_chain": "chainB","channel_id": "channel-0"},
			{"operation": "set","source_chain": "ChainC","destination_chain": "ChainB","channel_id": "channel-0"},
			{"operation": "change","source_chain": "chainB","destination_chain": "osmosis","new_source_chain": "ChainB"},
			{"operation": "change","source_chain": "osmosis","destination_chain": "chainB","new_destination_chain": "ChainB"}
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
	registryAddr, token0CB, token0CBA, CBAPath := suite.SetupCrosschainRegistry(ChainA)

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
	unwrapDenomQuery := fmt.Sprintf(`{"unwrap_denom": {"ibc_denom": "%s"}}`, token0CBA)
	unwrapDenomQueryResponse := suite.chainA.QueryContract(&suite.Suite, registryAddr, []byte(unwrapDenomQuery))
	expectedUnwrappedDenom := fmt.Sprintf(`[{"local_denom":"%s","on":"osmosis","via":"channel-0"},{"local_denom":"%s","on":"chainB","via":"channel-1"},{"local_denom":"token0","on":"chainC","via":null}]`, token0CBA, token0CB)

	suite.Require().Equal(expectedUnwrappedDenom, unwrapDenomQueryResponse)

	// Remove, set, and change links on the registry on chain A
	suite.modifyChainChannelLinks(registryAddr, ChainA)

	// Unwrap token0CB and check that the path has changed
	unwrapDenomQuery = fmt.Sprintf(`{"unwrap_denom": {"ibc_denom": "%s"}}`, token0CBA)
	unwrapDenomQueryResponse = suite.chainA.QueryContract(&suite.Suite, registryAddr, []byte(unwrapDenomQuery))
	expectedUnwrappedDenom = fmt.Sprintf(`[{"local_denom":"%s","on":"osmosis","via":"channel-0"},{"local_denom":"%s","on":"ChainB","via":"channel-1"},{"local_denom":"token0","on":"ChainC","via":null}]`, token0CBA, token0CB)
	suite.Require().Equal(expectedUnwrappedDenom, unwrapDenomQueryResponse)
}

func (suite *HooksTestSuite) TestUnwrapToken() {
	// Instantiate contract and set up three chains with funds sent between each
	registryAddr, _, token0CBA, _ := suite.SetupCrosschainRegistry(ChainA)
	suite.setChainChannelLinks(registryAddr, ChainA)

	chain := suite.GetChain(ChainA)
	ctx := chain.GetContext()
	owner := chain.SenderAccount.GetAddress()
	osmosisApp := chain.GetOsmosisApp()

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	msg := fmt.Sprintf(`{
		"modify_bech32_prefixes": {
		  "operations": [
			{"operation": "set", "chain_name": "osmosis", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainA", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainB", "prefix": "osmo"},
			{"operation": "set", "chain_name": "chainC", "prefix": "osmo"}
		  ]
		}
	  }
	  `)
	_, err := contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	// Check that the balances are correct: token0CB should be >100, token0CBA should be 0
	denomTrace0CA := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", suite.pathAC.EndpointA.ChannelID, "token0"))
	token0CA := denomTrace0CA.IBCDenom()
	initialWrappedBalance := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, token0CBA)
	suite.Require().Greater(initialWrappedBalance.Amount.Int64(), int64(100))
	initialBalance := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), registryAddr, token0CA)
	suite.Require().Equal(sdk.NewInt(0), initialBalance.Amount)

	msg = fmt.Sprintf(`{
		"unwrap_coin": {
			"receiver": "%s"
		}
	  }
	  `, registryAddr)
	var exec sdk.Msg = &types.MsgExecuteContract{Contract: registryAddr.String(), Msg: []byte(msg), Sender: owner.String(), Funds: sdk.NewCoins(sdk.NewCoin(token0CBA, sdk.NewInt(100)))}
	res, err := chain.SendMsgsNoCheck(exec)
	suite.Require().NoError(err)

	// "Relay the packet" by executing the receive  on chain B
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	res = suite.RelayPacketNoAck(packet, AtoB)

	// "Relay the packet" by executing the receive on chain C
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	res = suite.RelayPacketNoAck(packet, BtoC)

	// "Relay the packet" by executing the receive on chain A
	packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)
	res, ack := suite.RelayPacket(packet, CtoA)
	suite.Require().Contains(string(ack), "result")
	// Check th
	finalWrappedBalance := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, token0CBA)
	suite.Require().Equal(initialWrappedBalance.Amount.Sub(sdk.NewInt(100)), finalWrappedBalance.Amount)
	finalBalance := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), registryAddr, token0CA)
	suite.Require().Equal(sdk.NewInt(100), finalBalance.Amount)
}

func (suite *HooksTestSuite) TestCrosschainSwaps() {
	owner := suite.chainA.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)
	osmosisApp := suite.chainA.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	balanceSender := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, "token0")

	ctx := suite.chainA.GetContext()

	msg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s", "on_failed_delivery": "do_nothing"}}`,
		suite.chainB.SenderAccount.GetAddress(),
	)
	res, err := contractKeeper.Execute(ctx, crosschainAddr, owner, []byte(msg), sdk.NewCoins(sdk.NewCoin("token0", sdk.NewInt(1000))))
	suite.Require().NoError(err)
	var responseJson map[string]interface{}
	err = json.Unmarshal(res, &responseJson)
	suite.Require().NoError(err)
	suite.Require().Len(responseJson["sent_amount"].(string), 3) // Not using exact amount in case calculations change
	suite.Require().Equal(responseJson["denom"].(string), "token1")
	suite.Require().Equal(responseJson["channel_id"].(string), "channel-0")
	suite.Require().Equal(responseJson["receiver"].(string), suite.chainB.SenderAccount.GetAddress().String())
	suite.Require().Equal(responseJson["packet_sequence"].(float64), 1.0)

	balanceSender2 := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), owner, "token0")
	suite.Require().Equal(int64(1000), balanceSender.Amount.Sub(balanceSender2.Amount).Int64())
}

func (suite *HooksTestSuite) TestCrosschainSwapsViaIBCTest() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg, AtoB)

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
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s", "on_failed_delivery": "do_nothing", "next_memo":{}}}`,
		receiver,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, sdk.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), msg)
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
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg, AtoB)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()

	osmosisAppB := suite.chainB.GetOsmosisApp()
	balanceToken0 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	receiver := "juno1ka8v934kgrw6679fs9cuu0kesyl0ljjy4tmycx" // Will not exist on chainB

	// Generate swap instructions for the contract. This will send correctly on chainA, but fail to be received on chainB
	recoverAddr := suite.chainA.SenderAccounts[8].SenderAccount.GetAddress()
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s","on_failed_delivery": {"local_recovery_addr": "%s"}}}`,
		receiver, // Note that this is the chain A account, which does not exist on chain B
		recoverAddr,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, sdk.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), msg)
	_, receiveResult, _, err := suite.FullSend(transferMsg, BtoA)

	// We use the receive result here because the receive adds another packet to be sent back
	suite.Require().NoError(err)
	suite.Require().NotNil(receiveResult)

	// "Relay the packet" by executing the receive on chain B
	packet, err := ibctesting.ParsePacketFromEvents(receiveResult.GetEvents())
	suite.Require().NoError(err)
	_, ack2 := suite.RelayPacket(packet, AtoB)
	fmt.Println(string(ack2))

	balanceToken0After := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), initializer, token0IBC)
	suite.Require().Equal(int64(1000), balanceToken0.Amount.Sub(balanceToken0After.Amount).Int64())

	// The balance is stuck in the contract
	osmosisAppA := suite.chainA.GetOsmosisApp()
	balanceContract := osmosisAppA.BankKeeper.GetBalance(suite.chainA.GetContext(), crosschainAddr, "token1")
	suite.Require().Greater(balanceContract.Amount.Int64(), int64(0))

	// check that the contract knows this
	state := suite.chainA.QueryContract(
		&suite.Suite, crosschainAddr,
		[]byte(fmt.Sprintf(`{"recoverable": {"addr": "%s"}}`, recoverAddr)))
	suite.Require().Contains(state, "token1")
	suite.Require().Contains(state, `"sequence":2`)

	// Recover the stuck amount
	recoverMsg := `{"recover": {}}`
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisAppA.WasmKeeper)
	_, err = contractKeeper.Execute(suite.chainA.GetContext(), crosschainAddr, recoverAddr, []byte(recoverMsg), sdk.NewCoins())
	suite.Require().NoError(err)

	balanceRecovery := osmosisAppA.BankKeeper.GetBalance(suite.chainA.GetContext(), recoverAddr, "token1")
	suite.Require().Greater(balanceRecovery.Amount.Int64(), int64(0))
}

// CrosschainSwapsViaIBCBadSwap tests that if the crosschain-swap fails, the tokens are returned to the sender
// This is very similar to the two tests above, but the swap is done incorrectly
func (suite *HooksTestSuite) TestCrosschainSwapsViaIBCBadSwap() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg, AtoB)

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
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"min_output_amount":"50000"},"receiver":"%s", "on_failed_delivery": "do_nothing"}}`,
		receiver,
	)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, sdk.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), msg)
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
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)
	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(20000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg, AtoB)

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()

	recoverAddr := suite.chainA.SenderAccounts[8].SenderAccount.GetAddress()
	receiver := initializer

	// next_memo is set to `%s` after the SprintF. It is then format replaced in each test case.
	innerMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s","on_failed_delivery": {"local_recovery_addr": "%s"},"next_memo":%%s}}`,
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
		transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, sdk.NewInt(10)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), msg)
		_, _, ack, _ := suite.FullSend(transferMsg, BtoA)
		if tc.expPass {
			fmt.Println(ack)
			suite.Require().Contains(ack, "result", tc.memo)
		} else {
			suite.Require().Contains(ack, "error", tc.memo)
		}
	}
}

func (suite *HooksTestSuite) CreateIBCPoolOnChainB() {
	chain := suite.GetChain(ChainB)
	acc1 := chain.SenderAccount.GetAddress()
	bondDenom := chain.GetOsmosisApp().StakingKeeper.BondDenom(chain.GetContext())

	multiplier := sdk.NewDec(20)
	denomTrace1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token1"))
	token1IBC := denomTrace1.IBCDenom()

	uosmoAmount := gammtypes.InitPoolSharesSupply.ToDec().Mul(multiplier).RoundInt()

	defaultFutureGovernor := ""

	// pool assets
	defaultFooAsset := balancer.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin(bondDenom, uosmoAmount),
	}
	defaultBarAsset := balancer.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin(token1IBC, sdk.NewInt(10000)),
	}

	poolAssets := []balancer.PoolAsset{defaultFooAsset, defaultBarAsset}

	poolParams := balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}
	msg := balancer.NewMsgCreateBalancerPool(acc1, poolParams, poolAssets, defaultFutureGovernor)

	poolId, err := chain.GetOsmosisApp().PoolManagerKeeper.CreatePool(chain.GetContext(), msg)
	suite.Require().NoError(err)

	_, err = chain.GetOsmosisApp().GAMMKeeper.GetPoolAndPoke(chain.GetContext(), poolId)
	suite.Require().NoError(err)

}

func (suite *HooksTestSuite) SetupIBCRouteOnChainB(poolmanagerAddr, owner sdk.AccAddress) {
	chain := suite.GetChain(ChainB)
	denomTrace1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token1"))
	token1IBC := denomTrace1.IBCDenom()

	msg := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"token0","pool_route":[{"pool_id":"3","token_out_denom":"stake"},{"pool_id":"1","token_out_denom":"token0"}]}}`,
		token1IBC)
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	_, err := contractKeeper.Execute(chain.GetContext(), poolmanagerAddr, owner, []byte(msg), sdk.NewCoins())
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
	receiver := suite.chainA.SenderAccount.GetAddress()

	_, crosschainAddrA := suite.SetupCrosschainSwaps(ChainA)
	poolManagerAddrB, crosschainAddrB := suite.SetupCrosschainSwaps(ChainB)
	// Send some token0 and token1 tokens to B so that there are ibc token0 to send to A and crosschain-swap, and token1 to create the pool
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(500000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg, AtoB)
	transferMsg1 := NewMsgTransfer(sdk.NewCoin("token1", sdk.NewInt(500000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg1, AtoB)
	suite.CreateIBCPoolOnChainB()
	suite.SetupIBCRouteOnChainB(poolManagerAddrB, suite.chainB.SenderAccount.GetAddress())

	// Calculate the names of the tokens when swapped via IBC
	denomTrace0 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "token0"))
	token0IBC := denomTrace0.IBCDenom()

	balanceToken0IBCBefore := suite.chainA.GetOsmosisApp().BankKeeper.GetBalance(suite.chainA.GetContext(), receiver, token0IBC)
	fmt.Println("receiver now has: ", balanceToken0IBCBefore)
	suite.Require().Equal(int64(0), balanceToken0IBCBefore.Amount.Int64())

	//suite.Require().Equal(int64(0), balanceToken1.Amount.Int64())

	// Generate swap instructions for the contract
	nextMemo := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"osmosis_swap":{"output_denom":"token0","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s", "on_failed_delivery": "do_nothing"}}}}`,
		crosschainAddrB,
		receiver,
	)
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"token1","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"%s", "on_failed_delivery": "do_nothing", "next_memo": %s}}`,
		crosschainAddrB,
		nextMemo,
	)
	fmt.Println(swapMsg)
	// Generate full memo
	msg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddrA, swapMsg)
	// Send IBC transfer with the memo with crosschain-swap instructions
	transferMsg = NewMsgTransfer(sdk.NewCoin(token0IBC, sdk.NewInt(1000)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddrA.String(), msg)
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

func (suite *HooksTestSuite) ExecuteOutpostSwap(initializer, receiverAddr sdk.AccAddress, receiver string) {
	// Setup
	_, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)
	// Store and instantiate the outpost on chainB
	suite.chainB.StoreContractCode(&suite.Suite, "./bytecode/outpost.wasm")
	outpostAddr := suite.chainB.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"crosschain_swaps_contract": "%s", "osmosis_channel": "channel-0"}`, crosschainAddr), 1)

	// Send some token0 tokens to B so that there are ibc tokens to send to A and crosschain-swap
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), suite.chainA.SenderAccount.GetAddress().String(), initializer.String(), "")
	suite.FullSend(transferMsg, AtoB)

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
	_, err := contractKeeper.Execute(ctxB, outpostAddr, initializer, []byte(swapMsg), sdk.NewCoins(sdk.NewCoin(token0IBC, sdk.NewInt(1000))))
	suite.Require().NoError(err)
	suite.chainB.NextBlock()
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	// "Relay the packet" by executing the receive on chain A
	packet, err := ibctesting.ParsePacketFromEvents(ctxB.EventManager().Events())
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
	//fmt.Println("receiver now has: ", balanceToken1After)
	suite.Require().Greater(balanceToken1After.Amount.Int64(), int64(0))
}

func (suite *HooksTestSuite) TestOutpostSimplified() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	suite.ExecuteOutpostSwap(initializer, initializer, initializer.String())
}

func (suite *HooksTestSuite) TestOutpostExplicit() {
	initializer := suite.chainB.SenderAccount.GetAddress()
	suite.ExecuteOutpostSwap(initializer, initializer, fmt.Sprintf(`ibc:channel-0/%s`, initializer.String()))
}
