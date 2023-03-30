// This file defines helpers for
// sudo messages issued to the cosmwasm pool contract from the cosmwasm pool module.
package msg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SwapExactAmountIn
func NewSwapExactAmountInSudoMsg(sender string, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMin sdk.Int, swapFee sdk.Dec) SwapExactAmountInSudoMsg {
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
func NewSwapExactAmountOutSudoMsg(sender string, tokenInDenom string, tokenOut sdk.Coin, tokenInMaxAmount sdk.Int, swapFee sdk.Dec) SwapExactAmountOutSudoMsg {
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
