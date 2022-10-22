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
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
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
	initMsgBz := []byte(fmt.Sprintf(`{"owner": "%v"}`, suite.chainA.SenderAccount.GetAddress().String()))
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

	// Todo: Check the type of metadata is the expected one inside the hook so this test passes
	//    ...or just remove this test cause it should be part of the hooks PR
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
		Denom:    "ujuno",
		Amount:   "10",
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

func (suite *MiddlewareTestSuite) PrepareSimplePools() {
	fundsChainB := sdk.NewCoins(
		sdk.NewCoin("ujuno", sdk.NewInt(100000000000000)),
	)
	err := suite.chainB.FundAcc(suite.chainB.SenderAccount.GetAddress(), fundsChainB)
	suite.Require().NoError(err)

	funds := sdk.NewCoins(
		sdk.NewCoin("ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097", sdk.NewInt(100000000000000)),
		sdk.NewCoin("uosmo", sdk.NewInt(100000000000000)),
		sdk.NewCoin("uion", sdk.NewInt(100000000000000)),
	)
	err = suite.chainA.FundAcc(suite.chainA.SenderAccount.GetAddress(), funds)
	suite.Require().NoError(err)

	poolAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097", sdk.NewInt(100_000_000_000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("uosmo", sdk.NewInt(100_000_000_000)),
		},
	}
	poolParams := balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	}
	msg := balancer.NewMsgCreateBalancerPool(suite.chainA.SenderAccount.GetAddress(), poolParams, poolAssets, "")
	_, err = suite.chainA.GetOsmosisApp().GAMMKeeper.CreatePool(suite.chainA.GetContext(), msg)
	suite.Require().NoError(err)

	poolAssets[0].Token.Denom = "uosmo"
	poolAssets[1].Token.Denom = "uion"
	msg = balancer.NewMsgCreateBalancerPool(suite.chainA.SenderAccount.GetAddress(), poolParams, poolAssets, "")
	_, err = suite.chainA.GetOsmosisApp().GAMMKeeper.CreatePool(suite.chainA.GetContext(), msg)

	suite.Require().NoError(err)
}

func (suite *MiddlewareTestSuite) TestRecvTransferWithSwap() {
	suite.PrepareSimplePools()
	err := suite.chainA.StoreContractCode("./testdata/swaprouter.wasm")
	//// Move forward one block
	//suite.chainA.NextBlock()
	//suite.chainA.SenderAccount.SetSequence(suite.chainA.SenderAccount.GetSequence() + 1)
	//suite.chainA.Coordinator.IncrementTime()
	//
	//// Update both clients
	//err = suite.path.EndpointA.UpdateClient()
	//suite.Require().NoError(err)
	//err = suite.path.EndpointB.UpdateClient()
	//suite.Require().NoError(err)

	suite.Require().NoError(err)
	addr := suite.InstantiateSwapRouterContract(suite.chainA)
	contractAddr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	suite.Require().NoError(err)
	// Define a route on the swaprouter
	setRouteMsg := `{"set_route": {"input_denom": "ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097", "output_denom": "uion",  "pool_route": [
	    {"pool_id": 1, "token_out_denom": "uosmo"},
	    {"pool_id": 2, "token_out_denom": "uion"}
	]}}`

	_, err = suite.chainA.ExecuteContract(addr, suite.chainA.SenderAccount.GetAddress(), []byte(setRouteMsg), nil)
	suite.Require().NoError(err)

	metadata := fmt.Sprintf(`{
"wasm": {
    "contract": "%s",
    "execute": {"swap": 
	  {"input_coin": {"amount": "%d", "denom": "ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097"}, 
	   "output_denom": "uion", 
	   "slipage": {"max_price_impact_percentage": "3"}}
    }
  }
}`, contractAddr, 10)

	before := suite.chainB.GetOsmosisApp().BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress())
	fmt.Println(before)

	ackBytes := suite.receivePacket([]byte(metadata))
	contractBalances := suite.chainA.GetOsmosisApp().BankKeeper.GetAllBalances(suite.chainA.GetContext(), sdk.AccAddress(contractAddr))
	fmt.Println("contract", contractBalances)

	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")

	// Move forward one block
	suite.chainA.NextBlock()
	suite.chainA.Coordinator.IncrementTime()
	suite.chainB.NextBlock()
	suite.chainB.Coordinator.IncrementTime()

	//suite.chainA.SendMsgs()

	// Update both clients
	//err = suite.path.EndpointA.UpdateClient()
	//suite.Require().NoError(err)
	err = suite.path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	result := suite.chainB.GetOsmosisApp().BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress())
	fmt.Println(result)
}
