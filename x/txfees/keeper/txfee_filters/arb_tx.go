package txfee_filters

import (
	"encoding/json"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	authztypes "github.com/cosmos/cosmos-sdk/x/authz"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// See this for reference: https://github.com/osmosis-labs/affiliate-swap
type Swap struct {
	Routes            []poolmanagertypes.SwapAmountInRoute `json:"routes"`
	TokenOutMinAmount sdk.Coin                             `json:"token_out_min_amount"`
	FeePercentage     sdk.Dec                              `json:"fee_percentage"`
	FeeCollector      string                               `json:"fee_collector"`
	TokenIn           string                               `json:"token_in,omitempty"`
}

type AffiliateSwapMsg struct {
	Swap `json:"swap"`
}

// TokenDenomsOnPath implements types.SwapMsgRoute.
func (m AffiliateSwapMsg) TokenDenomsOnPath() []string {
	denoms := make([]string, 0, len(m.Routes)+1)
	denoms = append(denoms, m.TokenInDenom())
	for i := 0; i < len(m.Routes); i++ {
		denoms = append(denoms, m.Routes[i].TokenOutDenom)
	}
	return denoms
}

// TokenInDenom implements types.SwapMsgRoute.
func (m AffiliateSwapMsg) TokenInDenom() string {
	return m.TokenIn
}

// TokenOutDenom implements types.SwapMsgRoute.
func (m AffiliateSwapMsg) TokenOutDenom() string {
	lastPoolInRoute := m.Routes[len(m.Routes)-1]
	return lastPoolInRoute.TokenOutDenom
}

var _ poolmanagertypes.SwapMsgRoute = AffiliateSwapMsg{}

type CrosschainSwap struct {
}

// We check if a tx is an arbitrage for the mempool right now by seeing:
// 1) does start token of a msg = final token of msg (definitionally correct)
// 2) does it have multiple swap messages, with different tx ins. If so, we assume its an arb.
//   - This has false positives, but is intended to avoid the obvious solution of splitting
//     an arb into multiple messages.
//
// 3) We record all denoms seen across all swaps, and see if any duplicates. (TODO)
// 4) Contains both JoinPool and ExitPool messages in one tx.
//   - Has some false positives, but they seem relatively contrived.
//
// TODO: Move the first component to a future router module.
func IsArbTxLoose(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()

	swapInDenom := ""
	lpTypesSeen := make(map[gammtypes.LiquidityChangeType]bool, 2)
	isArb := false

	for _, m := range msgs {
		swapInDenom, isArb = isArbTxLooseAuthz(m, swapInDenom, lpTypesSeen)
		if isArb {
			return true
		}
	}

	return false
}

func isArbTxLooseAuthz(msg sdk.Msg, swapInDenom string, lpTypesSeen map[gammtypes.LiquidityChangeType]bool) (string, bool) {
	if authzMsg, ok := msg.(*authztypes.MsgExec); ok {
		msgs, _ := authzMsg.GetMessages()
		for _, m := range msgs {
			swapInDenom, isAuthz := isArbTxLooseAuthz(m, swapInDenom, lpTypesSeen)
			if isAuthz {
				return swapInDenom, true
			}
		}
		return swapInDenom, false
	}

	// Detects the affiliate swap message from the CosmWasm contract
	// See an example here:
	// // https://celatone.osmosis.zone/osmosis-1/txs/315EB6284778EBB5BAC0F94CC740F5D7E35DDA5BBE4EC9EC79F012548589C6E5
	if msgExecuteContract, ok := msg.(*wasmtypes.MsgExecuteContract); ok {
		// Grab token in from the funds sent to the contract
		tokensIn := msgExecuteContract.GetFunds()
		if len(tokensIn) != 1 {
			return swapInDenom, false
		}
		tokenIn := tokensIn[0]

		// Get the contract message and attempt to unmarshal it into the affiliate swap message
		contractMessage := msgExecuteContract.GetMsg()
		var affiliateSwapMsg AffiliateSwapMsg
		if err := json.Unmarshal(contractMessage, &affiliateSwapMsg); err != nil {
			// If we can't unmarshal it, it's not an affiliate swap message
			return swapInDenom, false
		}

		// Otherwise, we have an affiliate swap message, so we check if it's an arb
		affiliateSwapMsg.TokenIn = tokenIn.Denom
		swapInDenom, isArb := isArbTxLooseSwapMsg(affiliateSwapMsg, swapInDenom)
		if isArb {
			return swapInDenom, true
		}

		return swapInDenom, false
	}

	// Detect a contract execution via IBC hooks
	if msgIBCRecv, ok := msg.(*channeltypes.MsgRecvPacket); ok {
		var transferPacket transfertypes.FungibleTokenPacketData
		err := json.Unmarshal(msgIBCRecv.Packet.Data, &transferPacket)
		if err != nil {
			return swapInDenom, false
		}

		swapInDenom = getLoalIBCDenom(msgIBCRecv.Packet, transferPacket.Denom)

		payload, valid := isWasmHooksExecutePayload(transferPacket.Memo)
		if !valid {
			return swapInDenom, false
		}
		if checkWasmHookPayload(payload, swapInDenom) {
			return swapInDenom, true
		}
		return swapInDenom, false
	}

	// (4) Check that the tx doesn't have both JoinPool & ExitPool msgs
	lpMsg, isLpMsg := msg.(gammtypes.LiquidityChangeMsg)
	if isLpMsg {
		lpTypesSeen[lpMsg.LiquidityChangeType()] = true
		if len(lpTypesSeen) > 1 {
			return swapInDenom, true
		}
	}

	multiSwapMsg, isMultiSwapMsg := msg.(poolmanagertypes.MultiSwapMsgRoute)
	if isMultiSwapMsg {
		for _, swapMsg := range multiSwapMsg.GetSwapMsgs() {
			// TODO: Fix this later
			swapInDenom, isArb := isArbTxLooseSwapMsg(swapMsg, swapInDenom)
			if isArb {
				return swapInDenom, true
			}
		}
		return swapInDenom, false
	}

	swapMsg, isSwapMsg := msg.(poolmanagertypes.SwapMsgRoute)
	if !isSwapMsg {
		return swapInDenom, false
	}

	return isArbTxLooseSwapMsg(swapMsg, swapInDenom)
}

func isArbTxLooseSwapMsg(swapMsg poolmanagertypes.SwapMsgRoute, swapInDenom string) (string, bool) {
	// (1) Check that swap denom in != swap denom out
	if swapMsg.TokenInDenom() == swapMsg.TokenOutDenom() {
		return swapInDenom, true
	}

	// (2)
	if swapInDenom != "" && swapMsg.TokenInDenom() != swapInDenom {
		return swapInDenom, true
	}
	swapInDenom = swapMsg.TokenInDenom()
	return swapInDenom, false
}

func isWasmHooksExecutePayload(memo string) (map[string]interface{}, bool) {
	jsonObject := make(map[string]interface{})
	err := json.Unmarshal([]byte(memo), &jsonObject)
	if err != nil {
		return nil, false
	}
	wasm, ok := jsonObject["wasm"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	_, ok = wasm["contract"].(string)
	if !ok {
		return nil, false
	}
	msg, ok := wasm["msg"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	return msg, true
}

func getLoalIBCDenom(packet channeltypes.Packet, denom string) string {
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), denom) {
		// sender chain is not the source, unescrow tokens

		// remove prefix added by sender chain
		voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := denom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom := unprefixedDenom

		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
		if !denomTrace.IsNativeDenom() {
			denom = denomTrace.IBCDenom()
		}
		return denom
	}

	// sender chain is the source, mint vouchers

	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom

	// construct the denomination trace from the full raw denomination
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	voucherDenom := denomTrace.IBCDenom()
	return voucherDenom
}

func checkWasmHookPayload(payload map[string]interface{}, swapInDenom string) bool {
	// CrossChainSwaps message
	if swapMsg, ok := payload["osmosis_swap"].(map[string]interface{}); ok {
		if outputDenom, ok := swapMsg["output_denom"].(string); ok {
			return outputDenom == swapInDenom
		}
	}

	return false
}
