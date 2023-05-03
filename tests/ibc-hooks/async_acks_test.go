package ibc_hooks_test

import (
	"fmt"
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

	// We store that only that contract can send an ack for that packet sequence

}
