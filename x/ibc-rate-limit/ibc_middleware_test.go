package ibc_rate_limit_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/osmosis-labs/osmosis/v12/app"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/testutil"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"
	"github.com/stretchr/testify/suite"
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

func NewTransferPath(chainA, chainB *osmosisibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version
	return path
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
	suite.path = NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(suite.path)
}

// Helpers
func (suite *MiddlewareTestSuite) MessageFromAToB(denom string, amount sdk.Int, wrapDenom bool) sdk.Msg {
	var coins sdk.Coin
	var port, channel, accountFrom, accountTo string

	coins = sdk.NewCoin(denom, amount)
	if wrapDenom {
		coins = transfertypes.GetTransferCoin("transfer", "channel-0", denom, amount)
	}
	port = suite.path.EndpointA.ChannelConfig.PortID
	channel = suite.path.EndpointA.ChannelID
	accountFrom = suite.chainA.SenderAccount.GetAddress().String()
	accountTo = suite.chainB.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(0, 100)
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		coins,
		accountFrom,
		accountTo,
		timeoutHeight,
		0,
	)
}

func (suite *MiddlewareTestSuite) MessageFromBToA(denom string, amount sdk.Int, wrapDenom bool) sdk.Msg {
	coins := sdk.NewCoin(denom, amount)
	if wrapDenom {
		coins = transfertypes.GetTransferCoin("transfer", "channel-0", denom, amount)
	}
	port := suite.path.EndpointB.ChannelConfig.PortID
	channel := suite.path.EndpointB.ChannelID
	accountFrom := suite.chainB.SenderAccount.GetAddress().String()
	accountTo := suite.chainA.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(0, 100)
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		coins,
		accountFrom,
		accountTo,
		timeoutHeight,
		0,
	)
}

// Tests that a receiver address longer than 4096 is not accepted
func (suite *MiddlewareTestSuite) TestInvalidReceiver() {
	msg := transfertypes.NewMsgTransfer(
		suite.path.EndpointB.ChannelConfig.PortID,
		suite.path.EndpointB.ChannelID,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)),
		suite.chainB.SenderAccount.GetAddress().String(),
		strings.Repeat("x", 4097),
		clienttypes.NewHeight(0, 100),
		0,
	)
	ack, _ := suite.ExecuteReceive(msg)
	suite.Require().Contains(string(ack), "error",
		"acknowledgment is not an error")
	suite.Require().Contains(string(ack), sdkerrors.ErrInvalidAddress.Error(),
		"acknowledgment error is not of the right type")
}

func (suite *MiddlewareTestSuite) FullSendBToA(msg sdk.Msg) (string, error) {
	res, err := suite.chainB.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	res, err = suite.path.EndpointA.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	return string(ack), err
}

func (suite *MiddlewareTestSuite) FullSendAToB(msg sdk.Msg) (string, error) {
	res, err := suite.chainA.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = suite.path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	res, err = suite.path.EndpointB.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	return string(ack), err
}

func (suite *MiddlewareTestSuite) ExecuteReceive(msg sdk.Msg) (string, error) {
	res, err := suite.chainB.SendMsgsNoCheck(msg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	res, err = suite.path.EndpointA.RecvPacketWithResult(packet)
	suite.Require().NoError(err)

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	return string(ack), err
}

func (suite *MiddlewareTestSuite) AssertReceive(success bool, msg sdk.Msg) (string, error) {
	ack, err := suite.ExecuteReceive(msg)
	if success {
		suite.Require().NoError(err)
		suite.Require().NotContains(string(ack), "error",
			"acknowledgment is an error")
	} else {
		suite.Require().Contains(string(ack), "error",
			"acknowledgment is not an error")
		suite.Require().Contains(string(ack), types.ErrRateLimitExceeded.Error(),
			"acknowledgment error is not of the right type")
	}
	return ack, err
}

func (suite *MiddlewareTestSuite) AssertSend(success bool, msg sdk.Msg) (*sdk.Result, error) {
	r, err := suite.chainA.SendMsgsNoCheck(msg)
	if success {
		suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)
	} else {
		suite.Require().Error(err, "IBC send succeeded. Expected failure")
		suite.ErrorContains(err, types.ErrRateLimitExceeded.Error(), "Bad error type")
	}
	return r, err
}

func (suite *MiddlewareTestSuite) BuildChannelQuota(name, denom string, duration, send_precentage, recv_percentage uint32) string {
	return fmt.Sprintf(`
          {"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"%s", "duration": %d, "send_recv":[%d, %d]}] }
    `, denom, name, duration, send_precentage, recv_percentage)
}

// Tests

// Test that Sending IBC messages works when the middleware isn't configured
func (suite *MiddlewareTestSuite) TestSendTransferNoContract() {
	one := sdk.NewInt(1)
	suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, one, false))
}

// Test that Receiving IBC messages works when the middleware isn't configured
func (suite *MiddlewareTestSuite) TestReceiveTransferNoContract() {
	one := sdk.NewInt(1)
	suite.AssertReceive(true, suite.MessageFromBToA(sdk.DefaultBondDenom, one, false))
}

func (suite *MiddlewareTestSuite) initializeEscrow() (totalEscrow, expectedSed sdk.Int) {
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
	_, err := suite.FullSendAToB(suite.MessageFromAToB(sdk.DefaultBondDenom, transferAmount.Sub(sendAmount), false))
	suite.Require().NoError(err)
	// Send from A to B
	_, err = suite.FullSendBToA(suite.MessageFromBToA(sdk.DefaultBondDenom, transferAmount.Sub(sendAmount), false))
	suite.Require().NoError(err)

	return transferAmount, sendAmount
}

func (suite *MiddlewareTestSuite) fullSendTest(native bool) map[string]string {
	_, escrowAmount := suite.initializeEscrow()

	sendAmount := escrowAmount
	// Get the denom and amount to send
	denom := sdk.DefaultBondDenom
	if !native {
		denom = "transfer/channel-0/" + sdk.DefaultBondDenom
		osmosisApp := suite.chainA.GetOsmosisApp()

		sendAmount = osmosisApp.BankKeeper.GetSupply(suite.chainA.GetContext(), transfertypes.ParseDenomTrace(denom).IBCDenom()).Amount
	}

	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	quotas := suite.BuildChannelQuota("weekly", denom, 604800, 5, 5)
	addr := suite.chainA.InstantiateContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	//// Each user has 10% of the supply
	//escrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	//escrowed := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), escrowAddress, sdk.DefaultBondDenom)
	//
	////supply := osmosisApp.BankKeeper.GetSupplyWithOffset(suite.chainA.GetContext(), sdk.DefaultBondDenom)
	//quota := escrowed.Amount.QuoRaw(20)
	//half := quota.QuoRaw(2)
	//
	//fmt.Println("escrowed", escrowed)
	//fmt.Println("quota", quota)
	//fmt.Println("half", half)

	// send 2.5% (quota is 5%)
	suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, sendAmount, !native))

	// send 2.5% (quota is 5%)
	r, _ := suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, sendAmount, !native))

	// Calculate remaining allowance in the quota
	attrs := suite.ExtractAttributes(suite.FindEvent(r.GetEvents(), "wasm"))
	used, ok := sdk.NewIntFromString(attrs["weekly_used_out"])
	suite.Require().True(ok)

	suite.Require().Equal(used, sendAmount.MulRaw(2))

	// Sending above the quota should fail.
	suite.AssertSend(false, suite.MessageFromAToB(sdk.DefaultBondDenom, sdk.NewInt(1), !native))
	return attrs
}

// Test rate limiting on sends
func (suite *MiddlewareTestSuite) TestSendTransferWithRateLimitingNative() {
	suite.fullSendTest(true)
}

// Test rate limiting on sends
func (suite *MiddlewareTestSuite) TestSendTransferWithRateLimitingNonNative() {
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

	// Move both chains one block
	suite.chainA.NextBlock()
	suite.chainA.SenderAccount.SetSequence(suite.chainA.SenderAccount.GetSequence() + 1)
	suite.chainB.NextBlock()
	suite.chainB.SenderAccount.SetSequence(suite.chainB.SenderAccount.GetSequence() + 1)

	// Reset time + one second
	oneSecAfterReset := resetTime.Add(time.Second)
	suite.coordinator.IncrementTimeBy(oneSecAfterReset.Sub(suite.coordinator.CurrentTime))

	// Sending should succeed again
	suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, sdk.NewInt(1), false))
}

// Test rate limiting on receives
func (suite *MiddlewareTestSuite) fullRecvTest(native bool) {
	_, escrowAmount := suite.initializeEscrow()

	transferAmount := escrowAmount
	// Get the denom and amount to send
	denom := sdk.DefaultBondDenom
	if !native {
		denom = "transfer/channel-0/" + sdk.DefaultBondDenom
		osmosisApp := suite.chainA.GetOsmosisApp()

		transferAmount = osmosisApp.BankKeeper.GetSupply(suite.chainA.GetContext(), transfertypes.ParseDenomTrace(denom).IBCDenom()).Amount
	}

	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	quotas := suite.BuildChannelQuota("weekly", denom, 604800, 5, 5)
	addr := suite.chainA.InstantiateContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// Setup receiver chain's quota
	quota := transferAmount.QuoRaw(20)
	half := quota.QuoRaw(2)
	// receive 2.5% (quota is 5%)
	suite.AssertReceive(true, suite.MessageFromBToA(sdk.DefaultBondDenom, half, !native))

	// receive 2.5% (quota is 5%)
	suite.AssertReceive(true, suite.MessageFromBToA(sdk.DefaultBondDenom, half, !native))

	// Sending above the quota should fail.
	suite.AssertReceive(false, suite.MessageFromBToA(sdk.DefaultBondDenom, half, !native))
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithRateLimitingNative() {
	suite.fullRecvTest(true)
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithRateLimitingNonNative() {
	suite.fullRecvTest(false)
}

// Test no rate limiting occurs when the contract is set, but not quotas are condifured for the path
func (suite *MiddlewareTestSuite) TestSendTransferNoQuota() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	addr := suite.chainA.InstantiateContract(&suite.Suite, ``)
	suite.chainA.RegisterRateLimitingContract(addr)

	// send 1 token.
	// If the contract doesn't have a quota for the current channel, all transfers are allowed
	suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, sdk.NewInt(1), false))
}

// Test rate limits are reverted if a "send" fails
func (suite *MiddlewareTestSuite) TestFailedSendTransfer() {
	suite.initializeEscrow()
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	quotas := suite.BuildChannelQuota("weekly", sdk.DefaultBondDenom, 604800, 1, 1)
	addr := suite.chainA.InstantiateContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// Get the escrowed amount
	osmosisApp := suite.chainA.GetOsmosisApp()
	escrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	escrowed := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), escrowAddress, sdk.DefaultBondDenom)

	quota := escrowed.Amount.QuoRaw(100) // 1% of the escrowed amount

	// Use the whole quota
	coins := sdk.NewCoin(sdk.DefaultBondDenom, quota)
	port := suite.path.EndpointA.ChannelConfig.PortID
	channel := suite.path.EndpointA.ChannelID
	accountFrom := suite.chainA.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(0, 100)
	msg := transfertypes.NewMsgTransfer(port, channel, coins, accountFrom, "INVALID", timeoutHeight, 0)

	res, _ := suite.AssertSend(true, msg)

	// Sending again fails as the quota is filled
	suite.AssertSend(false, suite.MessageFromAToB(sdk.DefaultBondDenom, quota, false))

	// Move forward one block
	suite.chainA.NextBlock()
	suite.chainA.SenderAccount.SetSequence(suite.chainA.SenderAccount.GetSequence() + 1)
	suite.chainA.Coordinator.IncrementTime()

	// Update both clients
	err := suite.path.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = suite.path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	// Execute the acknowledgement from chain B in chain A

	// extract the sent packet
	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// recv in chain b
	res, err = suite.path.EndpointB.RecvPacketWithResult(packet)

	// get the ack from the chain b's response
	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	// manually relay it to chain a
	err = suite.path.EndpointA.AcknowledgePacket(packet, ack)
	suite.Require().NoError(err)

	// We should be able to send again because the packet that exceeded the quota failed and has been reverted
	suite.AssertSend(true, suite.MessageFromAToB(sdk.DefaultBondDenom, sdk.NewInt(1), false))
}
