package txfee_filters

import (
	"encoding/json"

	authztypes "github.com/cosmos/cosmos-sdk/x/authz"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// See this for reference: https://github.com/osmosis-labs/affiliate-swap
type Swap struct {
	Routes            []poolmanagertypes.SwapAmountInRoute `json:"routes"`
	TokenOutMinAmount sdk.Coin                             `json:"token_out_min_amount"`
	FeePercentage     osmomath.Dec                         `json:"fee_percentage"`
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
	if len(m.Routes) == 0 {
		return "no-token-out"
	}
	lastPoolInRoute := m.Routes[len(m.Routes)-1]
	return lastPoolInRoute.TokenOutDenom
}

var _ poolmanagertypes.SwapMsgRoute = AffiliateSwapMsg{}

type InputCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Slippage struct {
	MinOutputAmount string `json:"min_output_amount"`
}

type ContractSwap struct {
	InputCoin   InputCoin `json:"input_coin"`
	OutputDenom string    `json:"output_denom"`
	Slippage    Slippage  `json:"slippage"`
}

type ContractSwapMsg struct {
	ContractSwap `json:"swap"`
}

// TokenDenomsOnPath implements types.SwapMsgRoute.
func (c ContractSwapMsg) TokenDenomsOnPath() []string {
	return []string{c.InputCoin.Denom, c.OutputDenom}
}

// TokenInDenom implements types.SwapMsgRoute.
func (c ContractSwapMsg) TokenInDenom() string {
	return c.InputCoin.Denom
}

// TokenOutDenom implements types.SwapMsgRoute.
func (c ContractSwapMsg) TokenOutDenom() string {
	return c.OutputDenom
}

var _ poolmanagertypes.SwapMsgRoute = ContractSwapMsg{}

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

		// Check that the contract message is an affiliate swap message
		isAffliliateSwap := isAffiliateSwapMsg(contractMessage)
		isContractSwap := isContractSwapContractMsg(contractMessage)

		if !isAffliliateSwap && !isContractSwap {
			return swapInDenom, false
		}

		if isAffliliateSwap {
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
		}

		if isContractSwap {
			var contractSwapMsg ContractSwapMsg
			if err := json.Unmarshal(contractMessage, &contractSwapMsg); err != nil {
				// If we can't unmarshal it, it's not a contract swap message
				return swapInDenom, false
			}

			// Otherwise, we have a contract swap message, so we check if it's an arb
			swapInDenom, isArb := isArbTxLooseSwapMsg(contractSwapMsg, swapInDenom)
			if isArb {
				return swapInDenom, true
			}

			// Also, check sent tokenIn just in case.
			contractSwapMsg.InputCoin.Denom = tokenIn.Denom
			swapInDenom, isArb = isArbTxLooseSwapMsg(contractSwapMsg, swapInDenom)
			if isArb {
				return swapInDenom, true
			}
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

// TODO: Make this generic by using isJsonSuperset from https://github.com/osmosis-labs/osmosis/blob/d56de7365428f0282eeab05c1cc75787370ef997/x/authenticator/authenticator/message_filter.go#L173C6-L173C12
func isAffiliateSwapMsg(msg []byte) bool {
	// Check that the contract message is a valid JSON object
	jsonObject := make(map[string]interface{})
	err := json.Unmarshal(msg, &jsonObject)
	if err != nil {
		return false
	}

	// check the main key is "swap"
	swap, ok := jsonObject["swap"].(map[string]interface{})
	if !ok {
		return false
	}

	if routes, ok := swap["routes"].([]interface{}); !ok || len(routes) == 0 {
		return false
	}

	if tokenOutMinAmount, ok := swap["token_out_min_amount"].(map[string]interface{}); !ok || len(tokenOutMinAmount) == 0 {
		return false
	}

	return true
}

// check if this: https://celatone.osmosis.zone/osmosis-1/txs/8D20755D4E009CB72C763963A76886BCCCC5C2EBFC3F57266332710216A0D10D
func isContractSwapContractMsg(msg []byte) bool {
	// Check that the contract message is a valid JSON object
	jsonObject := make(map[string]interface{})
	err := json.Unmarshal(msg, &jsonObject)
	if err != nil {
		return false
	}

	// check the main key is "swap"
	swap, ok := jsonObject["swap"].(map[string]interface{})
	if !ok {
		return false
	}

	if input_coin, ok := swap["input_coin"].(map[string]interface{}); !ok || len(input_coin) == 0 {
		return false
	}

	if outputDenom, ok := swap["output_denom"].(string); !ok || len(outputDenom) == 0 {
		return false
	}

	return true
}
