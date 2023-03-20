// This file defines requests and responses for querying
// the cosmwasm pool contract from the cosmwasm pool module.
package cosmwasm

import sdk "github.com/cosmos/cosmos-sdk/types"

// GetPoolDenoms
type GetPoolDenoms struct {
	GetPoolDenoms struct{} `json:"get_pool_denoms"`
}

type GetPoolDenomsResponse struct {
	PoolDenoms []string `json:"pool_denoms"`
}

// CalcOutAmtGivenIn
type CalcOutGivenIn struct {
	TokenIn       sdk.Coin `json:"token_in"`
	TokenOutDenom string   `json:"token_out_denom"`
	SwapFee       sdk.Dec  `json:"swap_fee"`
}

type CalcOutAmtGivenInRequest struct {
	CalcOutGivenIn CalcOutGivenIn `json:"calc_out_given_in"`
}

func NewCalcOutAmtGivenInRequest(tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec) CalcOutAmtGivenInRequest {
	return CalcOutAmtGivenInRequest{
		CalcOutGivenIn: CalcOutGivenIn{
			TokenIn:       tokenIn,
			TokenOutDenom: tokenOutDenom,
			SwapFee:       swapFee,
		},
	}
}

type CalcOutAmtGivenInResponse struct {
	TokenOut sdk.Coin
}

// CalcInAmtGivenOut
type CalcInGivenOut struct {
	TokenInDenom string   `json:"token_in_denom"`
	TokenOut     sdk.Coin `json:"token_out"`
	SwapFee      sdk.Dec  `json:"swap_fee"`
}

type CalcInAmtGivenOutRequest struct {
	CalcInGivenOut CalcInGivenOut `json:"calc_in_given_out"`
}

func NewCalcInAmtGivenOutRequest(tokenInDenom string, tokenOut sdk.Coin, swapFee sdk.Dec) CalcInAmtGivenOutRequest {
	return CalcInAmtGivenOutRequest{
		CalcInGivenOut: CalcInGivenOut{
			TokenInDenom: tokenInDenom,
			TokenOut:     tokenOut,
			SwapFee:      swapFee,
		},
	}
}

type CalcInAmtGivenOutResponse struct {
	TokenIn sdk.Coin
}
