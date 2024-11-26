package ibc_hooks_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/tests/osmosisibctesting"
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
func (suite *HooksTestSuite) TransferCW20Tokens(path *ibctesting.Path, cw20Addr, cw20ics20Addr, receiver sdk.AccAddress, amount, memo string) (*abci.ExecTxResult, []byte) {
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
	packet, err := ibctesting.ParsePacketFromEvents(events.ToABCIEvents())
	suite.Require().NoError(err)
	result, ack := suite.RelayPacket(packet, CW20toA)
	suite.Require().Contains(string(ack), "result")
	return result, ack
}

func (suite *HooksTestSuite) setupCW20PoolAndRoutes(chain *osmosisibctesting.TestChain, swaprouterAddr sdk.AccAddress, cw20IbcDenom string, amount osmomath.Int) {
	osmosisAppA := chain.GetOsmosisApp()
	poolId := suite.CreateIBCPoolOnChain(ChainA, cw20IbcDenom, sdk.DefaultBondDenom, amount)

	// create a swap route for that token / poolId in both directions
	msg := fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"%s","pool_route":[{"pool_id":"%v","token_out_denom":"%s"}]}}`,
		cw20IbcDenom, sdk.DefaultBondDenom, poolId, sdk.DefaultBondDenom)
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisAppA.WasmKeeper)
	_, err := contractKeeper.Execute(chain.GetContext(), swaprouterAddr, chain.SenderAccount.GetAddress(), []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)
	msg = fmt.Sprintf(`{"set_route":{"input_denom":"%s","output_denom":"%s","pool_route":[{"pool_id":"%v","token_out_denom":"%s"}]}}`,
		sdk.DefaultBondDenom, cw20IbcDenom, poolId, cw20IbcDenom)
	_, err = contractKeeper.Execute(chain.GetContext(), swaprouterAddr, chain.SenderAccount.GetAddress(), []byte(msg), sdk.NewCoins())
	suite.Require().NoError(err)
}

func (suite *HooksTestSuite) getCW20Balance(chain *osmosisibctesting.TestChain, cw20Addr, addr sdk.AccAddress) osmomath.Int {
	queryMsg := fmt.Sprintf(`{"balance":{"address":"%s"}}`, addr)
	res := chain.QueryContractJson(&suite.Suite, cw20Addr, []byte(queryMsg))
	balance, ok := osmomath.NewIntFromString(res.Get("balance").String())
	if !ok {
		panic("could not parse balance")
	}
	return balance
}

// Test that the cw20-ics20 contract can be instantiated and used to send tokens to chainA
func (suite *HooksTestSuite) TestCW20ICS20() {
	// Hardcoding the cw20 ibc denom for simplicity
	cw20IbcDenom := "ibc/134A49086C1164C78313D57E69E5A8656D8AE8CF6BB45B52F2DBFEFAE6EE30B8"

	cw20Addr, cw20ics20Addr := suite.SetupCW20(ChainB)
	swaprouterAddr, crosschainAddr := suite.SetupCrosschainSwaps(ChainA, true)

	chainA := suite.GetChain(ChainA)
	chainB := suite.GetChain(ChainB)

	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = fmt.Sprintf("wasm.%s", cw20ics20Addr.String())
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	suite.pathCW20 = path

	suite.coordinator.Setup(path)

	osmosisAppB := chainB.GetOsmosisApp()

	// Send some cwtoken tokens from B to A via the new path
	amount := osmomath.NewInt(defaultPoolAmount)
	suite.TransferCW20Tokens(path, cw20Addr, cw20ics20Addr, chainA.SenderAccount.GetAddress(), amount.String(), "")

	// Create a pool for that token
	suite.setupCW20PoolAndRoutes(chainA, swaprouterAddr, cw20IbcDenom, amount)

	// Check that the receiver doesn't have any sdk.DefaultBondDenom
	stakeAB := suite.GetIBCDenom(ChainA, ChainB, sdk.DefaultBondDenom) // IBC denom for stake in B
	balanceStakeReceiver := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), stakeAB)
	suite.Require().Equal(int64(0), balanceStakeReceiver.Amount.Int64())

	// Transfer the tokens with a memo for XCS
	swapMsg := fmt.Sprintf(`{"osmosis_swap":{"output_denom":"%s","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB/%s", "on_failed_delivery": "do_nothing", "next_memo":{}}}`,
		sdk.DefaultBondDenom,
		chainB.SenderAccount.GetAddress(),
	)
	xcsMsg := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)
	serializedMemo, err := json.Marshal(xcsMsg)
	suite.Require().NoError(err)
	result, ack := suite.TransferCW20Tokens(path, cw20Addr, cw20ics20Addr, crosschainAddr, "100", string(serializedMemo))
	suite.Require().Contains(string(ack), "result")

	// Relay the packet created by the XCS contract back to the receiver
	packet, err := ibctesting.ParsePacketFromEvents(result.GetEvents())
	suite.Require().NoError(err)
	suite.RelayPacket(packet, AtoB)

	balanceStakeReceiver = osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), stakeAB)
	suite.Require().Greater(balanceStakeReceiver.Amount.Int64(), int64(0))

	// Now swap on the other direction
	receiver2 := chainB.SenderAccounts[1].SenderAccount.GetAddress() // Using a different receiver that has no cw20s
	cw20Balance := suite.getCW20Balance(chainB, cw20Addr, receiver2)

	swapMsg = fmt.Sprintf(`{"osmosis_swap":{"output_denom":"%s","slippage":{"twap": {"window_seconds": 1, "slippage_percentage":"20"}},"receiver":"chainB-cw20/%s", "on_failed_delivery": "do_nothing", "next_memo":{}}}`,
		cw20IbcDenom,
		receiver2,
	)
	xcsMsg = fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": %s } }`, crosschainAddr, swapMsg)

	transferMsg := NewMsgTransfer(sdk.NewCoin(stakeAB, osmomath.NewInt(10)), suite.chainB.SenderAccount.GetAddress().String(), crosschainAddr.String(), suite.pathAB.EndpointB.ChannelID, xcsMsg)
	_, recvResult, _, _ := suite.FullSend(transferMsg, BtoA)

	packet, err = ibctesting.ParsePacketFromEvents(recvResult.GetEvents())
	suite.Require().NoError(err)
	suite.RelayPacket(packet, AtoCW20)

	// Check that the receiver has 10 less ibc'd stake than before
	balanceStakeReceiver2 := osmosisAppB.BankKeeper.GetBalance(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress(), stakeAB)
	suite.Require().Equal(int64(10), balanceStakeReceiver.Amount.Sub(balanceStakeReceiver2.Amount).Int64())

	// Check that the receiver has more cw20 tokens than before
	newCw20Balance := suite.getCW20Balance(chainB, cw20Addr, receiver2)
	suite.Require().Greater(newCw20Balance.Int64(), cw20Balance.Int64())
}
