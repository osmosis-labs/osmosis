package ibc_hooks_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
)

// Instantiate the cw20 and cw20-ics20 contract
func (suite *HooksTestSuite) SetupCW20(chainName Chain) (sdk.AccAddress, sdk.AccAddress) {
	// Instantiate the cw20 contract on chainB
	chain := suite.GetChain(chainName)
	cw20CodeId := chain.StoreContractCodeDirect(&suite.Suite, "./bytecode/cw20_base.wasm")
	// create contract with initial balances
	cw20Addr := chain.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"name": "CW20-token0", "symbol": "cwtoken", "decimals": 6, "initial_balances": [{"address": "%s", "amount": "1000000000000000000000000"}]}`,
			chain.SenderAccount.GetAddress().String()), cw20CodeId)
	// Store the cw20-ics20 contract on chainB
	cw20ics20CodeId := chain.StoreContractCodeDirect(&suite.Suite, "./bytecode/cw20_ics20.wasm")
	// Instantiate the cw20-ics20 contract on chainB
	cw20ics20Addr := chain.InstantiateContract(&suite.Suite,
		fmt.Sprintf(`{"default_timeout": 1200, "gov_contract": "%s", "allowlist": [], "default_gas_limit": 200000}`,
			chain.SenderAccount.GetAddress().String()), cw20ics20CodeId)
	return cw20Addr, cw20ics20Addr
}

// Function to easily transfer the created cw20 tokens to chainA
func (suite *HooksTestSuite) TransferCW20Tokens(path *ibctesting.Path, cw20Addr, cw20ics20Addr, receiver sdk.AccAddress, amount, memo string) (*sdk.Result, []byte) {
	chainB := suite.GetChain(ChainB)

	osmosisApp := chainB.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)

	if len(memo) == 0 {
		memo = "\"\""
	}

	transferMsg := fmt.Sprintf(`{"channel": "%s", "remote_address": "%s", "memo": %s}`, path.EndpointB.ChannelID, receiver.String(), memo)
	transferMsgBase64 := base64.StdEncoding.EncodeToString([]byte(transferMsg))

	ctx := chainB.GetContext()
	_, err := contractKeeper.Execute(
		ctx,
		cw20Addr,
		chainB.SenderAccount.GetAddress(),
		[]byte(fmt.Sprintf(`{"send": {"contract": "%s", "amount": "%s", "msg": "%s"}}`, cw20ics20Addr.String(), amount, transferMsgBase64)),
		nil,
	)
	suite.Require().NoError(err)

	// Move forward one block
	chainB.NextBlock()
	chainB.Coordinator.IncrementTime()

	// Update both clients
	err = path.EndpointB.UpdateClient()
	suite.Require().NoError(err)
	err = path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	events := ctx.EventManager().Events()
	packet, err := ibctesting.ParsePacketFromEvents(events)
	suite.Require().NoError(err)
	result, ack := suite.RelayPacket(packet, CW20toA)
	suite.Require().Contains(string(ack), "result")
	return result, ack
}

// Test that the cw20-ics20 contract can be instantiated and used to send tokens to chainA
func (suite *HooksTestSuite) TestCW20ICS20() {
	cw20Addr, cw20ics20Addr := suite.SetupCW20(ChainB)
	swaprouterAddr, crosschainAddr := suite.SetupCrosschainSwaps(ChainA)

	chainA := suite.GetChain(ChainA)
	chainB := suite.GetChain(ChainB)

	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = fmt.Sprintf("wasm.%s", cw20ics20Addr.String())
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	suite.pathCW20 = path

	suite.coordinator.Setup(path)

	// Send some cwtoken tokens from B to A via the  new path
	amount := sdk.NewInt(defaultPoolAmount)
	suite.TransferCW20Tokens(path, cw20Addr, cw20ics20Addr, chainA.SenderAccount.GetAddress(), amount.String(), "")

	// Check receiver's balance
	osmosisApp := chainA.GetOsmosisApp()
	// Hardcoding the ibc denom for simplicity
	ibcDenom := "ibc/134A49086C1164C78313D57E69E5A8656D8AE8CF6BB45B52F2DBFEFAE6EE30B8"
	balanceSender := osmosisApp.BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), ibcDenom)
	fmt.Println("balanceSender:", balanceSender)
	suite.Require().Equal(amount, balanceSender.Amount)

	// Create a pool for that token
	poolId := suite.CreateIBCPoolOnChain(ChainA, ibcDenom, "stake", amount)

	// create a swap route for that token / poolId
	msg := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"%s","pool_route":[{"pool_id":"%v","token_out_denom":"%s"}]}}`,
		ibcDenom, "stake", poolId, "stake")
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	_, err := contractKeeper.Execute(chainA.GetContext(), swaprouterAddr, suite.chainA.SenderAccount.GetAddress(), []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)

	// Transfer the tokens with a memo for XCS
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"stake","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB-cw20/%s", "on_failed_delivery": "do_nothing", "next_memo":{}}}`,
		chainB.SenderAccount.GetAddress(),
	)
	xcsMsg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	serializedMemo, _ := json.Marshal(xcsMsg)
	suite.TransferCW20Tokens(path, cw20Addr, cw20ics20Addr, crosschainAddr, "100", string(serializedMemo))

}
