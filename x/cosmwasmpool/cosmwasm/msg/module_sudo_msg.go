// This file defines helpers for
// sudo messages issued to the cosmwasm pool contract from the cosmwasm pool module.
package msg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// SwapExactAmountIn
func NewSwapExactAmountInSudoMsg(sender string, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMin osmomath.Int, swapFee osmomath.Dec) SwapExactAmountInSudoMsg {
	return SwapExactAmountInSudoMsg{
		SwapExactAmountIn: SwapExactAmountIn{
			Sender:            sender,
			TokenIn:           tokenIn,
			TokenOutDenom:     tokenOutDenom,
			TokenOutMinAmount: tokenOutMin,
			SwapFee:           swapFee,
		},
	}
}

// SwapExactAmountOut
func NewSwapExactAmountOutSudoMsg(sender string, tokenInDenom string, tokenOut sdk.Coin, tokenInMaxAmount osmomath.Int, swapFee osmomath.Dec) SwapExactAmountOutSudoMsg {
	return SwapExactAmountOutSudoMsg{
		SwapExactAmountOut: SwapExactAmountOut{
			Sender:           sender,
			TokenInDenom:     tokenInDenom,
			TokenOut:         tokenOut,
			TokenInMaxAmount: tokenInMaxAmount,
			SwapFee:          swapFee,
		},
	}
}
