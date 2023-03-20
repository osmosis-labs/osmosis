// This file defines requests and responses for querying
// the cosmwasm pool contract from the cosmwasm pool model.
package cosmwasm

import sdk "github.com/cosmos/cosmos-sdk/types"

// GetSwapFee
type GetSwapFee struct {
	GetSwapFee struct{} `json:"get_swap_fee"`
}

type GetSwapFeeResponse struct {
	SwapFee sdk.Dec `json:"swap_fee"`
}

// GetExitFee
type GetExitFee struct {
	GetExitFee struct{} `json:"get_exit_fee"`
}

type GetExitFeeResponse struct {
	ExitFee sdk.Dec `json:"exit_fee"`
}

// SpotPrice
type SpotPrice struct {
	SpotPrice struct {
		QuoteAssetDenom string `json:"quote_asset_denom"`
		BaseAssetDenom  string `json:"base_asset_denom"`
	} `json:"spot_price"`
}

type SpotPriceResponse struct {
	SpotPrice string `json:"spot_price"`
}

// GetTotalPoolLiquidity
type GetTotalPoolLiquidity struct {
	GetTotalPoolLiquidity struct{} `json:"get_total_pool_liquidity"`
}

type GetTotalPoolLiquidityResponse struct {
	TotalPoolLiquidity sdk.Coins `json:"total_pool_liquidity"`
}
