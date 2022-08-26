package ibc_rate_limit_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/osmosis-labs/osmosis/v11/app"
	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/x/ibc-rate-limit/testutil"
	"github.com/osmosis-labs/osmosis/v11/x/ibc-rate-limit/types"
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

// NewValidMessage generates a new sdk.Msg of type MsgTransfer.
// forward=true means that the message will be a "send" message, while forward=false is for  a "receive" message.
// amount represents the amount transferred
func (suite *MiddlewareTestSuite) NewValidMessage(forward bool, amount sdk.Int) sdk.Msg {
	var coins sdk.Coin
	var port, channel, accountFrom, accountTo string

	if forward {
		coins = sdk.NewCoin(sdk.DefaultBondDenom, amount)
		port = suite.path.EndpointA.ChannelConfig.PortID
		channel = suite.path.EndpointA.ChannelID
		accountFrom = suite.chainA.SenderAccount.GetAddress().String()
		accountTo = suite.chainB.SenderAccount.GetAddress().String()
	} else {
		coins = transfertypes.GetTransferCoin(
			suite.path.EndpointB.ChannelConfig.PortID,
			suite.path.EndpointB.ChannelID,
			sdk.DefaultBondDenom,
			sdk.NewInt(1),
		)
		coins = sdk.NewCoin(sdk.DefaultBondDenom, amount)
		port = suite.path.EndpointB.ChannelConfig.PortID
		channel = suite.path.EndpointB.ChannelID
		accountFrom = suite.chainB.SenderAccount.GetAddress().String()
		accountTo = suite.chainA.SenderAccount.GetAddress().String()
	}

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
			"acknoledgment is an error")
	} else {
		suite.Require().Contains(string(ack), "error",
			"acknoledgment is not an error")
		suite.Require().Contains(string(ack), types.RateLimitExceededMsg,
			"acknoledgment error is not of the right type")
	}
	return ack, err
}

func (suite *MiddlewareTestSuite) AssertSend(success bool, msg sdk.Msg) (*sdk.Result, error) {
	r, err := suite.chainA.SendMsgsNoCheck(msg)
	if success {
		suite.Require().NoError(err, "IBC send failed. Expected success. %s", err)
	} else {
		suite.Require().Error(err, "IBC send succeeded. Expected failure")
		suite.ErrorContains(err, types.RateLimitExceededMsg, "Bad error type")
	}
	return r, err
}

func (suite *MiddlewareTestSuite) BuildChannelQuota(name string, duration, send_precentage, recv_percentage uint32) string {
	return fmt.Sprintf(`
          {"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"%s", "duration": %d, "send_recv":[%d, %d]}] }
    `, sdk.DefaultBondDenom, name, duration, send_precentage, recv_percentage)
}

// Tests

// Test that Sending IBC messages works when the middleware isn't configured
func (suite *MiddlewareTestSuite) TestSendTransferNoContract() {
	one := sdk.NewInt(1)
	suite.AssertSend(true, suite.NewValidMessage(true, one))
}

// Test that Receiving IBC messages works when the middleware isn't configured
func (suite *MiddlewareTestSuite) TestReceiveTransferNoContract() {
	one := sdk.NewInt(1)
	suite.AssertReceive(true, suite.NewValidMessage(false, one))
}

// Test rate limiting on sends
func (suite *MiddlewareTestSuite) TestSendTransferWithRateLimiting() map[string]string {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	quotas := suite.BuildChannelQuota("weekly", 604800, 5, 5)
	addr := suite.chainA.InstantiateContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// Setup sender chain's quota
	osmosisApp := suite.chainA.GetOsmosisApp()

	// Each user has 10% of the supply
	supply := osmosisApp.BankKeeper.GetSupplyWithOffset(suite.chainA.GetContext(), sdk.DefaultBondDenom)
	quota := supply.Amount.QuoRaw(20)
	half := quota.QuoRaw(2)

	// send 2.5% (quota is 5%)
	suite.AssertSend(true, suite.NewValidMessage(true, half))

	// send 2.5% (quota is 5%)
	r, _ := suite.AssertSend(true, suite.NewValidMessage(true, half))

	// Calculate remaining allowance in the quota
	attrs := suite.ExtractAttributes(suite.FindEvent(r.GetEvents(), "wasm"))
	used, ok := sdk.NewIntFromString(attrs["weekly_used_out"])
	suite.Require().True(ok)

	suite.Require().Equal(used, quota)

	// Sending above the quota should fail.
	suite.AssertSend(false, suite.NewValidMessage(true, sdk.NewInt(1)))
	return attrs
}

// Test rate limits are reset when the specified time period has passed
func (suite *MiddlewareTestSuite) TestSendTransferReset() {
	// Same test as above, but the quotas get reset after time passes
	attrs := suite.TestSendTransferWithRateLimiting()
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
	suite.AssertSend(true, suite.NewValidMessage(true, sdk.NewInt(1)))
}

// Test rate limiting on receives
func (suite *MiddlewareTestSuite) TestRecvTransferWithRateLimiting() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	quotas := suite.BuildChannelQuota("weekly", 604800, 5, 5)
	addr := suite.chainA.InstantiateContract(&suite.Suite, quotas)
	suite.chainA.RegisterRateLimitingContract(addr)

	// Setup receiver chain's quota
	osmosisApp := suite.chainA.GetOsmosisApp()

	// Each user has 10% of the supply
	supply := osmosisApp.BankKeeper.GetSupplyWithOffset(suite.chainA.GetContext(), sdk.DefaultBondDenom)
	quota := supply.Amount.QuoRaw(20)
	half := quota.QuoRaw(2)

	// receive 2.5% (quota is 5%)
	suite.AssertReceive(true, suite.NewValidMessage(false, half))

	// receive 2.5% (quota is 5%)
	suite.AssertReceive(true, suite.NewValidMessage(false, half))

	// Sending above the quota should fail. Adding some extra here because the cap is increasing. See test bellow.
	suite.AssertReceive(false, suite.NewValidMessage(false, sdk.NewInt(1)))
}

// Test no rate limiting occurs when the contract is set, but not quotas are condifured for the path
func (suite *MiddlewareTestSuite) TestSendTransferNoQuota() {
	// Setup contract
	suite.chainA.StoreContractCode(&suite.Suite)
	addr := suite.chainA.InstantiateContract(&suite.Suite, ``)
	suite.chainA.RegisterRateLimitingContract(addr)

	// send 1 token.
	// If the contract doesn't have a quota for the current channel, all transfers are allowed
	suite.AssertSend(true, suite.NewValidMessage(true, sdk.NewInt(1)))
}

// Test the contract used for these tests is the same contract used for E2E testing
func (s *MiddlewareTestSuite) TestRateLimitingE2ETestsSetupCorrectly() {
	// Checking the rate limiting e2e tests are setup correctly
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	f1, err := ioutil.ReadFile(fmt.Sprintf("%s/testdata/rate_limiter.wasm", dir))
	s.Require().NoError(err)
	f2, err := ioutil.ReadFile(fmt.Sprintf("%s/../../tests/e2e/scripts/rate_limiter.wasm", dir))
	s.Require().NoError(err)
	s.Require().True(bytes.Equal(f1, f2))
}
