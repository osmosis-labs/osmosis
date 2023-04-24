package ibc_hooks_test

import (
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
)

func (suite *HooksTestSuite) TestPathValidation() {
	owner := suite.chainA.SenderAccount.GetAddress()
	registryAddr, _, _, _ := suite.SetupCrosschainRegistry(ChainA)
	suite.setChainChannelLinks(registryAddr, ChainA)

	osmosisApp := suite.chainA.GetOsmosisApp()
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
	_, err := contractKeeper.Execute(suite.chainA.GetContext(), registryAddr, owner, []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	transferMsg := NewMsgTransfer(sdk.NewCoin("token0", sdk.NewInt(2000)), suite.chainB.SenderAccount.GetAddress().String(), owner.String(), suite.GetSenderChannel(ChainB, ChainA), "")
	suite.FullSend(transferMsg, BtoA)
	tonenBA := suite.GetIBCDenom(ChainB, ChainA, "token0")

	ctx := suite.chainA.GetContext()

	msg = `{"propose_pfm":{"chain": "chainB"}}`
	_, err = contractKeeper.Execute(ctx, registryAddr, owner, []byte(msg), sdk.NewCoins(sdk.NewCoin(tonenBA, sdk.NewInt(1))))
	suite.Require().NoError(err)

	// Move forward one block
	suite.chainA.NextBlock()
	suite.chainA.Coordinator.IncrementTime()

	// Update both clients
	err = suite.pathAB.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = suite.pathAB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	events := ctx.EventManager().Events()
	packet, err := ibctesting.ParsePacketFromEvents(events)
	suite.Require().NoError(err)
	result := suite.RelayPacketNoAck(packet, AtoB) // No ack because it's a forward

	packet, err = ibctesting.ParsePacketFromEvents(result.GetEvents())
	suite.Require().NoError(err)
	suite.RelayPacket(packet, BtoA)
}
