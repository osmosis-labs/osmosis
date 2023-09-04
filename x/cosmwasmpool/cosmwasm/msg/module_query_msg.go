// This file defines helpers for querying
// the cosmwasm pool contract from the cosmwasm pool module.
package msg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// CalcOutAmtGivenIn
func NewCalcOutAmtGivenInRequest(tokenIn sdk.Coin, tokenOutDenom string, swapFee osmomath.Dec) CalcOutAmtGivenInRequest {
	return CalcOutAmtGivenInRequest{
		CalcOutAmtGivenIn: CalcOutAmtGivenIn{
			TokenIn:       tokenIn,
			TokenOutDenom: tokenOutDenom,
			SwapFee:       swapFee,
		},
	}
}

// CalcInAmtGivenOut
func NewCalcInAmtGivenOutRequest(tokenInDenom string, tokenOut sdk.Coin, swapFee osmomath.Dec) CalcInAmtGivenOutRequest {
	return CalcInAmtGivenOutRequest{
		CalcInAmtGivenOut: CalcInAmtGivenOut{
			TokenInDenom: tokenInDenom,
			TokenOut:     tokenOut,
			SwapFee:      swapFee,
		},
	}
}
