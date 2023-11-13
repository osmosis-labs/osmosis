package txfee_filters

import (
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"

	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
