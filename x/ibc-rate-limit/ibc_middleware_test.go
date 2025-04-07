package ibc_rate_limit_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	txfeetypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/tests/osmosisibctesting"
	"github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
)

type MiddlewareTestSuite struct {
	apptesting.KeeperTestHelper

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *osmosisibctesting.TestChain
	chainB *osmosisibctesting.TestChain
	chainC *osmosisibctesting.TestChain
	path   *ibctesting.Path
	pathAC *ibctesting.Path
}

var oldConsensusMinFee = txfeetypes.ConsensusMinFee

// Setup
func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func NewTransferPath(chainA, chainB *osmosisibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version
	return path
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.SkipIfWSL()
	// TODO: This needs to get removed. Waiting on https://github.com/cosmos/ibc-go/issues/3123
	txfeetypes.ConsensusMinFee = osmomath.ZeroDec()
	suite.Setup()
	ibctesting.DefaultTestingAppInit = osmosisibctesting.SetupTestingApp
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(1)),
	}
	// Remove epochs to prevent  minting
	err := suite.chainA.MoveEpochsToTheFuture()
	suite.Require().NoError(err)
	// Create second chain
	suite.chainB = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(2)),
	}
	suite.path = NewTransferPath(suite.chainA, suite.chainB)
	err = suite.chainB.MoveEpochsToTheFuture()
	suite.Require().NoError(err)
	suite.coordinator.Setup(suite.path)
	// setup a third chain
	suite.chainC = &osmosisibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(3)),
	}
	suite.pathAC = NewTransferPath(suite.chainA, suite.chainC)
	err = suite.chainC.MoveEpochsToTheFuture()
	suite.Require().NoError(err)
	suite.coordinator.Setup(suite.pathAC)
}

// TODO: This needs to get removed. Waiting on https://github.com/cosmos/ibc-go/issues/3123
func (suite *MiddlewareTestSuite) TearDownSuite() {
	txfeetypes.ConsensusMinFee = oldConsensusMinFee

	for _, dir := range osmosisibctesting.TestingDirectories {
		os.RemoveAll(dir)
	}
}

// Helpers
func (suite *MiddlewareTestSuite) MessageFromAToB(denom string, amount osmomath.Int) sdk.Msg {
	coin := sdk.NewCoin(denom, amount)
	port := suite.path.EndpointA.ChannelConfig.PortID
	channel := suite.path.EndpointA.ChannelID
	accountFrom := suite.chainA.SenderAccount.GetAddress().String()
	accountTo := suite.chainB.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(10, 100)
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		coin,
		accountFrom,
		accountTo,
		timeoutHeight,
		uint64(time.Now().UnixNano()),
		"",
	)
}

func (suite *MiddlewareTestSuite) MessageFromBToA(denom string, amount osmomath.Int) sdk.Msg {
	coin := sdk.NewCoin(denom, amount)
	port := suite.path.EndpointB.ChannelConfig.PortID
	channel := suite.path.EndpointB.ChannelID
	accountFrom := suite.chainB.SenderAccount.GetAddress().String()
	accountTo := suite.chainA.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(10, 100)
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		coin,
		accountFrom,
		accountTo,
		timeoutHeight,
		uint64(time.Now().UnixNano()),
		"",
	)
}

func (suite *MiddlewareTestSuite) MessageFromAToC(denom string, amount osmomath.Int) sdk.Msg {
	coin := sdk.NewCoin(denom, amount)
	port := suite.pathAC.EndpointA.ChannelConfig.PortID
	channel := suite.pathAC.EndpointA.ChannelID
	accountFrom := suite.chainA.SenderAccount.GetAddress().String()
	accountTo := suite.chainC.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(10, 100)
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		coin,
		accountFrom,
		accountTo,
		timeoutHeight,
		uint64(time.Now().UnixNano()),
		"",
	)
}

func CalculateChannelValue(ctx sdk.Context, denom string, bankKeeper bankkeeper.Keeper) osmomath.Int {
	return bankKeeper.GetSupplyWithOffset(ctx, denom).Amount

	// ToDo: The commented-out code below is what we want to happen, but we're temporarily
	//  using the whole supply for efficiency until there's a solution for
	//  https://github.com/cosmos/ibc-go/issues/2664

	// For non-native (ibc) tokens, return the supply if the token in osmosis
	//if strings.HasPrefix(denom, "ibc/") {
	//	return bankKeeper.GetSupplyWithOffset(ctx, denom).Amount
	//}
	//
	// For native tokens, obtain the balance held in escrow for all potential channels
	//channels := channelKeeper.GetAllChannels(ctx)
	//balance := osmomath.NewInt(0)
	//for _, channel := range channels {
	//	escrowAddress := transfertypes.GetEscrowAddress("transfer", channel.ChannelId)
	//	balance = balance.Add(bankKeeper.GetBalance(ctx, escrowAddress, denom).Amount)
	//
	//}
	//return balance
}

func (suite *MiddlewareTestSuite) FullSendBToA(msg sdk.Msg) (*abci.ExecTxResult, string, error) {
	sendResult, err := suite.chainB.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	suite.Require().NoError(err)

	err = suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	res, err := suite.path.EndpointA.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	return sendResult, string(ack), err
}

func (suite *MiddlewareTestSuite) FullSendAToB(msg sdk.Msg) (*abci.ExecTxResult, string, error) {
	sendResult, err := suite.chainA.SendMsgsNoCheck(msg)
	if err != nil {
		return nil, "", err
	}

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	if err != nil {
		return nil, "", err
	}

	err = suite.path.EndpointB.UpdateClient()
	if err != nil {
		return nil, "", err
	}

	res, err := suite.path.EndpointB.RecvPacketWithResult(packet)
	if err != nil {
		return nil, "", err
	}

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	if err != nil {
		return nil, "", err
	}

	err = suite.path.EndpointA.UpdateClient()
	if err != nil {
		return nil, "", err
	}
	err = suite.path.EndpointB.UpdateClient()
	if err != nil {
		return nil, "", err
	}

	return sendResult, string(ack), nil
}

func (suite *MiddlewareTestSuite) FullSendAToC(msg sdk.Msg) (*abci.ExecTxResult, string, error) {
	sendResult, err := suite.chainA.SendMsgsNoCheck(msg)
	if err != nil {
		return nil, "", err
	}

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	if err != nil {
		return nil, "", err
	}

	err = suite.pathAC.EndpointB.UpdateClient()
	if err != nil {
		return nil, "", err
	}

	res, err := suite.pathAC.EndpointB.RecvPacketWithResult(packet)
	if err != nil {
		return nil, "", err
	}

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	if err != nil {
		return nil, "", err
	}

	err = suite.pathAC.EndpointA.UpdateClient()
	if err != nil {
		return nil, "", err
	}
	err = suite.pathAC.EndpointB.UpdateClient()
	if err != nil {
		return nil, "", err
	}

	return sendResult, string(ack), nil
}

func (suite *MiddlewareTestSuite) AssertReceive(success bool, msg sdk.Msg) (string, error) {
	_, ack, err := suite.FullSendBToA(msg)
	if success {
		suite.Require().NoError(err)
		suite.Require().NotContains(ack, "error",
			"acknowledgment is an error")
	} else {
		suite.Require().Contains(ack, "error",
			"acknowledgment is not an error")
		suite.Require().Contains(ack, fmt.Sprintf("ABCI code: %d", types.ErrRateLimitExceeded.ABCICode()),
			"acknowledgment error is not of the right type")
	}
	return ack, err
}

func (suite *MiddlewareTestSuite) AssertSend(success bool, msg sdk.Msg) (*abci.ExecTxResult, error) {
	r, _, err := suite.FullSendAToB(msg)
	if success {
		suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)
	} else {
		suite.Require().Error(err, "IBC send succeeded. Expected failure")
		suite.ErrorContains(err, types.ErrRateLimitExceeded.Error(), "Bad error type")
	}
	return r, err
}

func (suite *MiddlewareTestSuite) BuildChannelQuota(name, channel, denom string, duration, send_percentage, recv_percentage uint32) string {
	return fmt.Sprintf(`
          {"channel_id": "%s", "denom": "%s", "quotas": [{"name":"%s", "duration": %d, "send_recv":[%d, %d]}] }
    `, channel, denom, name, duration, send_percentage, recv_percentage)
}

// Tests

// Test that Sending IBC messages works when the middleware isn't configured
func (suite *MiddlewareTestSuite) TestSendTransferNoContract() {
	one := osmomath.NewInt(1)
	_, err := suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, one))
	suite.Require().NoError(err)
}

// Test that Receiving IBC messages works when the middleware isn't configured
func (suite *MiddlewareTestSuite) TestReceiveTransferNoContract() {
	one := osmomath.NewInt(1)
	_, err := suite.AssertReceive(true, suite.MessageFromBToA(sdk.DefaultBondDenom, one))
	suite.Require().NoError(err)
}

func (suite *MiddlewareTestSuite) initializeEscrow() (totalEscrow, expectedSed osmomath.Int) {
	osmosisApp := suite.chainA.GetOsmosisApp()
	supply := osmosisApp.BankKeeper.GetSupplyWithOffset(suite.chainA.GetContext(), sdk.DefaultBondDenom)

	// Move some funds from chainA to chainB so that there is something in escrow
	// Each user has 10% of the supply, so we send most of the funds from one user to chainA
	transferAmount := supply.Amount.QuoRaw(20)

	// When sending, the amount we're sending goes into escrow before we enter the middleware and thus
	// it's used as part of the channel value in the rate limiting contract
	// To account for that, we subtract the amount we'll send first (2.5% of transferAmount) here
	sendAmount := transferAmount.QuoRaw(40)

	// Send from A to B
	_, _, err := suite.FullSendAToB(suite.MessageFromAToB(sdk.DefaultBondDenom, transferAmount.Sub(sendAmount)))
	suite.Require().NoError(err)
	// Send from A to B
	_, _, err = suite.FullSendBToA(suite.MessageFromBToA(sdk.DefaultBondDenom, transferAmount.Sub(sendAmount)))
	suite.Require().NoError(err)

	return transferAmount, sendAmount
}

func (suite *MiddlewareTestSuite) fullSendTest(native bool) map[string]string {
	quotaPercentage := 5
	suite.initializeEscrow()
	// Get the denom and amount to send
	denom := sdk.DefaultBondDenom
	channel := "channel-0"
	if !native {
		denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", denom))
		fmt.Println(denomTrace)
		denom = denomTrace.IBCDenom()
	}

	osmosisApp := suite.chainA.GetOsmosisApp()

	// This is the first one. Inside the tests. It works as expected.
	channelValue := CalculateChannelValue(suite.chainA.GetContext(), denom, osmosisApp.BankKeeper)

	// The amount to be sent is send 2.5% (quota is 5%)
	quota := channelValue.QuoRaw(int64(100 / quotaPercentage))
	sendAmount := quota.QuoRaw(2)

	fmt.Printf("Testing send rate limiting for denom=%s, channelValue=%s, quota=%s, sendAmount=%s\n", denom, channelValue, quota, sendAmount)

	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	quotas := suite.BuildChannelQuota("weekly", channel, denom, 604800, 5, 5)
	fmt.Println(quotas)
	addr := suite.chainA.InstantiateRLContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// send 2.5% (quota is 5%)
	fmt.Printf("Sending %s from A to B. Represented in chain A as wrapped? %v\n", denom, !native)
	_, err := suite.AssertSend(true, suite.MessageFromAToB(denom, sendAmount))
	suite.Require().NoError(err)

	// send 2.5% (quota is 5%)
	fmt.Println("trying to send ", sendAmount)
	r, _ := suite.AssertSend(true, suite.MessageFromAToB(denom, sendAmount))

	// Calculate remaining allowance in the quota
	attrs := suite.ExtractAttributes(suite.FindEvent(r.GetEvents(), "wasm"))

	used, ok := osmomath.NewIntFromString(attrs["weekly_used_out"])
	suite.Require().True(ok)

	suite.Require().Equal(used, sendAmount.MulRaw(2))

	// Sending above the quota should fail. We use 2 instead of 1 here to avoid rounding issues
	_, err = suite.AssertSend(false, suite.MessageFromAToB(denom, osmomath.NewInt(2)))
	suite.Require().Error(err)
	return attrs
}

// Test rate limiting on sends
func (suite *MiddlewareTestSuite) TestSendTransferWithRateLimitingNative() {
	// Sends denom=stake from A->B. Rate limit receives "stake" in the packet. Nothing to do in the contract
	suite.fullSendTest(true)
}

// Test rate limiting on sends
func (suite *MiddlewareTestSuite) TestSendTransferWithRateLimitingNonNative() {
	// Sends denom=ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878 from A->B.
	// Rate limit receives "transfer/channel-0/stake" in the packet (because transfer.relay.SendTransfer is called before the middleware)
	// and should hash it before calculating the value
	suite.fullSendTest(false)
}

// Test rate limits are reset when the specified time period has passed
func (suite *MiddlewareTestSuite) TestSendTransferReset() {
	// Same test as above, but the quotas get reset after time passes
	attrs := suite.fullSendTest(true)
	parts := strings.Split(attrs["weekly_period_end"], ".") // Splitting timestamp into secs and nanos
	secs, err := strconv.ParseInt(parts[0], 10, 64)
	suite.Require().NoError(err)
	nanos, err := strconv.ParseInt(parts[1], 10, 64)
	suite.Require().NoError(err)
	resetTime := time.Unix(secs, nanos)

	// Move chainA forward one block
	suite.chainA.NextBlock()

	// Reset time + one second
	oneSecAfterReset := resetTime.Add(time.Second)
	suite.coordinator.IncrementTimeBy(oneSecAfterReset.Sub(suite.coordinator.CurrentTime))

	// Sending should succeed again
	_, err = suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, osmomath.NewInt(1)))
	suite.Require().NoError(err)
}

// Test rate limiting on receives
func (suite *MiddlewareTestSuite) fullRecvTest(native bool) {
	quotaPercentage := 4
	suite.initializeEscrow()
	// Get the denom and amount to send
	sendDenom := sdk.DefaultBondDenom
	localDenom := sdk.DefaultBondDenom
	channel := "channel-0"
	if native {
		denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", localDenom))
		localDenom = denomTrace.IBCDenom()
	} else {
		denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", sendDenom))
		sendDenom = denomTrace.IBCDenom()
	}

	osmosisApp := suite.chainA.GetOsmosisApp()

	channelValue := CalculateChannelValue(suite.chainA.GetContext(), localDenom, osmosisApp.BankKeeper)

	// The amount to be sent is 2% (quota is 4%)
	quota := channelValue.QuoRaw(int64(100 / quotaPercentage))
	sendAmount := quota.QuoRaw(2)

	fmt.Printf("Testing recv rate limiting for denom=%s, channelValue=%s, quota=%s, sendAmount=%s\n", localDenom, channelValue, quota, sendAmount)

	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	quotas := suite.BuildChannelQuota("weekly", channel, localDenom, 604800, 4, 4)
	addr := suite.chainA.InstantiateRLContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// receive 2.5% (quota is 5%)
	fmt.Printf("Sending %s from B to A. Represented in chain A as wrapped? %v\n", sendDenom, native)
	_, err := suite.AssertReceive(true, suite.MessageFromBToA(sendDenom, sendAmount))
	suite.Require().NoError(err)

	// receive 2.5% (quota is 5%)
	_, err = suite.AssertReceive(true, suite.MessageFromBToA(sendDenom, sendAmount))
	suite.Require().NoError(err)

	// Sending above the quota should fail. We send 2 instead of 1 to account for rounding errors
	_, err = suite.AssertReceive(false, suite.MessageFromBToA(sendDenom, osmomath.NewInt(2)))
	suite.Require().NoError(err)
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithRateLimitingNative() {
	// Sends denom=stake from B->A.
	// Rate limit receives "stake" in the packet and should wrap it before calculating the value
	// types.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) should return false => Wrap the token
	suite.fullRecvTest(true)
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithRateLimitingNonNative() {
	// Sends denom=ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878 from B->A.
	// Rate limit receives "transfer/channel-0/stake" in the packet and should turn it into "stake"
	// types.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) should return true => unprefix. If unprefixed is not local, hash.
	suite.fullRecvTest(false)
}

// Test no rate limiting occurs when the contract is set, but no quotas are configured for the path
func (suite *MiddlewareTestSuite) TestSendTransferNoQuota() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	addr := suite.chainA.InstantiateRLContract(&suite.Suite, ``)
	suite.chainA.RegisterRateLimitingContract(addr)

	// send 1 token.
	// If the contract doesn't have a quota for the current channel, all transfers are allowed
	_, err := suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, osmomath.NewInt(1)))
	suite.Require().NoError(err)
}

// Test rate limits are reverted if a "send" fails
func (suite *MiddlewareTestSuite) TestFailedSendTransfer() {
	suite.initializeEscrow()
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	quotas := suite.BuildChannelQuota("weekly", "channel-0", sdk.DefaultBondDenom, 604800, 1, 1)
	addr := suite.chainA.InstantiateRLContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// Get the escrowed amount
	osmosisApp := suite.chainA.GetOsmosisApp()
	// ToDo: This is what we eventually want here, but using the full supply temporarily for performance reasons. See CalculateChannelValue
	// escrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	// escrowed := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), escrowAddress, sdk.DefaultBondDenom)
	escrowed := osmosisApp.BankKeeper.GetSupplyWithOffset(suite.chainA.GetContext(), sdk.DefaultBondDenom)
	quota := escrowed.Amount.QuoRaw(100) // 1% of the escrowed amount

	// Use the whole quota
	coins := sdk.NewCoin(sdk.DefaultBondDenom, quota)
	port := suite.path.EndpointA.ChannelConfig.PortID
	channel := suite.path.EndpointA.ChannelID
	accountFrom := suite.chainA.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(10, 100)
	msg := transfertypes.NewMsgTransfer(port, channel, coins, accountFrom, "INVALID", timeoutHeight, 0, "")

	// Sending the message manually because AssertSend updates both clients. We need to update the clients manually
	// for this test so that the failure to receive on chain B happens after the second packet is sent from chain A.
	// That way we validate that chain A is blocking as expected, but the flow is reverted after the receive failure is
	// acknowledged on chain A
	res, err := suite.chainA.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	// Sending again fails as the quota is filled
	_, err = suite.AssertSend(false, suite.MessageFromAToB(sdk.DefaultBondDenom, quota))
	suite.Require().Error(err)

	// Move forward one block
	suite.chainA.NextBlock()
	suite.chainA.Coordinator.IncrementTime()

	// Update both clients
	err = suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	// Execute the acknowledgement from chain B in chain A

	// extract the sent packet
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// recv in chain b
	newRes, err := suite.path.EndpointB.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	// get the ack from the chain b's response
	ack, err := ibctesting.ParseAckFromEvents(newRes.GetEvents())
	suite.Require().NoError(err)

	// manually relay it to chain a
	err = suite.path.EndpointA.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)

	// We should be able to send again because the packet that exceeded the quota failed and has been reverted
	_, err = suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, osmomath.NewInt(1)))
	suite.Require().NoError(err)
}

func (suite *MiddlewareTestSuite) TestUnsetRateLimitingContract() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	addr := suite.chainA.InstantiateRLContract(&suite.Suite, "")
	suite.chainA.RegisterRateLimitingContract(addr)

	// Unset the contract param
	params, err := types.NewParams("")
	suite.Require().NoError(err)
	osmosisApp := suite.chainA.GetOsmosisApp()
	paramSpace, ok := osmosisApp.AppKeepers.ParamsKeeper.GetSubspace(types.ModuleName)
	suite.Require().True(ok)
	// N.B.: this panics if validation fails.
	paramSpace.SetParamSet(suite.chainA.GetContext(), &params)
}

// Test rate limits are reverted if a "send" fails
func (suite *MiddlewareTestSuite) TestNonICS20() {
	suite.initializeEscrow()
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	quotas := suite.BuildChannelQuota("weekly", "channel-0", sdk.DefaultBondDenom, 604800, 1, 1)
	addr := suite.chainA.InstantiateRLContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	osmosisApp := suite.chainA.GetOsmosisApp()

	data := []byte("{}")
	_, err := osmosisApp.RateLimitingICS4Wrapper.SendPacket(suite.chainA.GetContext(), capabilitytypes.NewCapability(1), "wasm.osmo1873ls0d60tg7hk00976teq9ywhzv45u3hk2urw8t3eau9eusa4eqtun9xn", "channel-0", clienttypes.NewHeight(0, 0), 1, data)

	suite.Require().Error(err)
	// This will error out, but not because of rate limiting
	suite.Require().NotContains(err.Error(), "rate limit")
	suite.Require().Contains(err.Error(), "channel not found")
}

func (suite *MiddlewareTestSuite) TestDenomRestrictionFlow() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/rate_limiter.wasm")
	quotas := suite.BuildChannelQuota("weekly", "channel-0", sdk.DefaultBondDenom, 604800, 1, 1)
	contractAddr := suite.chainA.InstantiateRLContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(contractAddr)
	osmosisApp := suite.chainA.GetOsmosisApp()
	govModule := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	denom := sdk.DefaultBondDenom
	sendAmount := osmomath.NewInt(1000)
	acceptedChannel := suite.path.EndpointA.ChannelID

	// Successfully send a denom before any restrictions are added.
	_, err := suite.AssertSend(true, suite.MessageFromAToB(denom, sendAmount))
	suite.Require().NoError(err, "Send should succeed without restrictions")

	// Sending on a diff channel should work
	_, _, err = suite.FullSendAToC(suite.MessageFromAToC(denom, sendAmount))
	suite.Require().NoError(err, "Send on alternative channel should work")

	// Add a restriction that only allows sending on the accepted channel
	restrictionMsg := fmt.Sprintf(`{"set_denom_restrictions": {"denom":"%s","allowed_channels":["%s"]}}`, denom, acceptedChannel)
	_, err = suite.chainA.ExecuteContract(contractAddr, govModule, []byte(restrictionMsg), sdk.Coins{})
	suite.Require().NoError(err)

	// Sending on the accepted channel should succeed
	_, err = suite.AssertSend(true, suite.MessageFromAToB(denom, sendAmount))
	suite.Require().NoError(err, "Send on accepted channel should succeed")

	// Sending on any other channel should fail
	_, err = suite.AssertSend(false, suite.MessageFromAToC(denom, sendAmount))
	suite.Require().Error(err, "Send on blocked channel should fail")

	// Unset the restriction and verify that sending on other channels works again
	unsetMsg := fmt.Sprintf(`{"unset_denom_restrictions": {"denom":"%s"}}`, denom)
	_, err = suite.chainA.ExecuteContract(contractAddr, govModule, []byte(unsetMsg), sdk.Coins{})
	suite.Require().NoError(err, "Unsetting denom restriction should succeed")

	// Sending again on the previously blocked channel should now succeed
	_, _, err = suite.FullSendAToC(suite.MessageFromAToC(denom, sendAmount))
	suite.Require().NoError(err, "Send on previously blocked channel should succeed after unsetting restriction")

}
