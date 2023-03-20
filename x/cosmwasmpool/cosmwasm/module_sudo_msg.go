// This file defines requests and responses for issuing
// sudo messages to the cosmwasm pool contract from the cosmwasm pool module.
package cosmwasm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SwapExactAmountIn
type SwapExactAmountIn struct {
	Sender        string   `json:"sender"`
	TokenIn       sdk.Coin `json:"token_in"`
	TokenOutDenom string   `json:"token_out_denom"`
	TokenOutMin   sdk.Int  `json:"token_out_min"`
	SwapFee       sdk.Dec  `json:"swap_fee"`
}

type SwapExactAmountInRequest struct {
	SwapExactAmountIn SwapExactAmountIn `json:"swap_exact_amount_in"`
}

func NewSwapExactAmountInRequest(sender string, tokenIn sdk.Coin, tokenOutDenom string, tokenOutMin sdk.Int, swapFee sdk.Dec) SwapExactAmountInRequest {
	return SwapExactAmountInRequest{
		SwapExactAmountIn: SwapExactAmountIn{
			Sender:        sender,
			TokenIn:       tokenIn,
			TokenOutDenom: tokenOutDenom,
			TokenOutMin:   tokenOutMin,
			SwapFee:       swapFee,
		},
	}
}

type SwapExactAmountInResponse struct {
	TokenOutAmount sdk.Int `json:"token_out_amount"`
}

// SwapExactAmountOut
type SwapExactAmountOutRequest struct {
	SwapExactAmountOut SwapExactAmountOut `json:"swap_exact_amount_out"`
}

type SwapExactAmountOut struct {
	Sender           string   `json:"sender"`
	TokenInDenom     string   `json:"token_in_denom"`
	TokenOut         sdk.Coin `json:"token_out"`
	TokenInMaxAmount sdk.Int  `json:"token_in_max_amount"`
	SwapFee          sdk.Dec  `json:"swap_fee"`
}

func NewSwapExactAmountOutRequest(sender string, tokenInDenom string, tokenOut sdk.Coin, tokenInMaxAmount sdk.Int, swapFee sdk.Dec) SwapExactAmountOutRequest {
	return SwapExactAmountOutRequest{
		SwapExactAmountOut: SwapExactAmountOut{
			Sender:           sender,
			TokenInDenom:     tokenInDenom,
			TokenOut:         tokenOut,
			TokenInMaxAmount: tokenInMaxAmount,
			SwapFee:          swapFee,
		},
	}
}

type SwapExactAmountOutResponse struct {
	TokenInAmount sdk.Int `json:"token_in_amount"`
}
