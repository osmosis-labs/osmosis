package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the oracle Querier
const (
	QuerySwap           = "swap"
	QueryTerraPoolDelta = "terra_pool_delta"
	QueryParameters     = "parameters"
)

// QuerySwapParams for query
// - 'custom/market/swap'
type QuerySwapParams struct {
	OfferCoin sdk.Coin `json:"offer_coin"`
	AskDenom  string   `json:"ask_denom"`
}

// NewQuerySwapParams returns param object for swap query
func NewQuerySwapParams(offerCoin sdk.Coin, askDenom string) QuerySwapParams {
	return QuerySwapParams{
		OfferCoin: offerCoin,
		AskDenom:  askDenom,
	}
}
