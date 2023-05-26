package ibc_hooks_test

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
	"github.com/osmosis-labs/osmosis/v15/app"
)

func (suite *HooksTestSuite) forceContractToEmitAckForPacket(osmosisApp *app.OsmosisApp, contractAddr sdk.AccAddress, packet channeltypes.Packet) ([]byte, error) {
	packetJson, err := json.Marshal(packet)
	suite.Require().NoError(err)

	msg := fmt.Sprintf(`{"force_emit_ibc_ack": {"packet": %s, "channel": "channel-0"}}`, packetJson)
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	return contractKeeper.Execute(suite.chainA.GetContext(), contractAddr, suite.chainA.SenderAccount.GetAddress(), []byte(msg), sdk.NewCoins())

}

func (suite *HooksTestSuite) TestWasmHooksAsyncAcks() {
	sender := suite.chainB.SenderAccount.GetAddress()
	osmosisApp := suite.chainA.GetOsmosisApp()

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
	allAcks := osmosisApp.IBCKeeper.ChannelKeeper.GetAllPacketAcks(suite.chainA.GetContext())
	suite.Require().Equal(1, len(allAcks))

	// Try to emit an ack for a packet that already has been acked. This should fail

	// we extract the packet that has been acked here to test later that our contract can't emit an ack for it
	alreadyAckedPacket, err := ibctesting.ParsePacketFromEvents(sendResult.GetEvents())
	suite.Require().NoError(err)

	_, err = suite.forceContractToEmitAckForPacket(osmosisApp, contractAddr, alreadyAckedPacket)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "no ack actor set for channel-0 packet 1")

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

	// Store a second contract and ask that one to emit an ack for the packet that the first contract sent
	contractAddr2 := suite.chainA.InstantiateContract(&suite.Suite, "{}", 1)
	_, err = suite.forceContractToEmitAckForPacket(osmosisApp, contractAddr2, packet)
	// This should fail because the new contract is not authorized to emit acks for that packet
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "is not allowed to send an ack for channel channel-0 packet 2")

	// only the contract that sent the packet can send an ack for that packet sequence
	_, err = suite.forceContractToEmitAckForPacket(osmosisApp, contractAddr, packet)
	suite.Require().NoError(err)

	allAcks = osmosisApp.IBCKeeper.ChannelKeeper.GetAllPacketAcks(suite.chainA.GetContext())
	fmt.Println(allAcks)
	suite.Require().Equal(2, len(allAcks))

}
