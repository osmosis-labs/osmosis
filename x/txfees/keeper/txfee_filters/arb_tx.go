package txfee_filters

import (
	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"

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

	for _, m := range msgs {
		// (4) Check that the tx doesn't have both JoinPool & ExitPool msgs
		lpMsg, isLpMsg := m.(gammtypes.LiquidityChangeMsg)
		if isLpMsg {
			lpTypesSeen[lpMsg.LiquidityChangeType()] = true
			if len(lpTypesSeen) > 1 {
				return true
			}
		}

		swapMsg, isSwapMsg := m.(gammtypes.SwapMsgRoute)
		if !isSwapMsg {
			continue
		}

		// (1) Check that swap denom in != swap denom out
		if swapMsg.TokenInDenom() == swapMsg.TokenOutDenom() {
			return true
		}

		// (2)
		if swapInDenom != "" && swapMsg.TokenInDenom() != swapInDenom {
			return true
		}
		swapInDenom = swapMsg.TokenInDenom()
	}

	return false
}
