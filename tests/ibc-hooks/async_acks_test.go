package ibc_hooks_test

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
)

func (suite *HooksTestSuite) TestWasmHooksAsyncAcks() {
	sender := suite.chainB.SenderAccount.GetAddress()

	// Instantiate a contract that knows how to send async Acks
	suite.chainA.StoreContractCode(&suite.Suite, "./bytecode/echo.wasm")
	contractAddr := suite.chainA.InstantiateContract(&suite.Suite, "{}", 1)

	// Calls that don't specify async acks work as expected
	memo := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"async": {"use_async": false}}}}`, contractAddr)
	suite.fundAccount(suite.chainB, sender)
	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), sender.String(), contractAddr.String(), "channel-0", memo)
	sendResult, receiveResult, ack, err := suite.FullSend(transferMsg, BtoA)
	suite.Require().NoError(err)
	suite.Require().NotNil(sendResult)
	suite.Require().NotNil(receiveResult)
	suite.Require().NotNil(ack)

	// the ack has been written
	osmosisApp := suite.chainA.GetOsmosisApp()
	allAcks := osmosisApp.IBCKeeper.ChannelKeeper.GetAllPacketAcks(suite.chainA.GetContext())
	suite.Require().Equal(1, len(allAcks))

	// Calls that specify async Acks work and no Acks are sent
	memo = fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"async": {"use_async": true}}}}`, contractAddr)
	suite.fundAccount(suite.chainB, sender)
	transferMsg = NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), sender.String(), contractAddr.String(), "channel-0", memo)

	sendResult, err = suite.chainB.SendMsgsNoCheck(transferMsg)
	suite.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	suite.Require().NoError(err)

	receiveResult = suite.RelayPacketNoAck(packet, BtoA)
	newAck, err := ibctesting.ParseAckFromEvents(receiveResult.GetEvents())
	suite.Require().Error(err) // No ack!
	suite.Require().Nil(newAck)

	// No new ack has been written
	allAcks = osmosisApp.IBCKeeper.ChannelKeeper.GetAllPacketAcks(suite.chainA.GetContext())
	suite.Require().Equal(1, len(allAcks))

	// TODO: add test of packet 1 not being found for emit ack (as it's already been acked)

	// We store that only that contract can send an ack for that packet sequence
	packetJson, err := json.Marshal(packet)
	suite.Require().NoError(err)

	msg := fmt.Sprintf(`{"force_emit_ibc_ack": {"packet": %s, "channel": "channel-0"}}`, packetJson)
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	_, err = contractKeeper.Execute(suite.chainA.GetContext(), contractAddr, suite.chainA.SenderAccount.GetAddress(), []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	allAcks = osmosisApp.IBCKeeper.ChannelKeeper.GetAllPacketAcks(suite.chainA.GetContext())
	fmt.Println(allAcks)
	suite.Require().Equal(2, len(allAcks))

}
