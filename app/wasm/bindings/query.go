package wasmbindings

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

// OsmosisQuery contains osmosis custom queries.
// See https://github.com/osmosis-labs/osmosis-bindings/blob/main/packages/bindings/src/query.rs
type OsmosisQuery struct {
	/// Given a subdenom minted by a contract via `OsmosisMsg::MintTokens`,
	/// returns the full denom as used by `BankMsg::Send`.
	FullDenom *FullDenom `json:"full_denom,omitempty"`
	/// For a given pool ID, list all tokens traded on it with current liquidity (spot).
	/// As well as the total number of LP shares and their denom.
	PoolState *PoolState `json:"pool_state,omitempty"`
	/// Return current spot price swapping In for Out on given pool ID.
	/// Warning: this can easily be manipulated via sandwich attacks, do not use as price oracle.
	/// We will add TWAP for more robust price feed.
	SpotPrice *SpotPrice `json:"spot_price,omitempty"`
	/// Return current spot price swapping In for Out on given pool ID.
	EstimateSwap *EstimateSwap `json:"estimate_swap,omitempty"`
	/// Returns the admin of a denom, if the denom is a Token Factory denom.
	DenomAdmin *DenomAdmin `json:"denom_admin,omitempty"`
}

type FullDenom struct {
	CreatorAddr string `json:"creator_addr"`
	Subdenom    string `json:"subdenom"`
}

type DenomAdmin struct {
	Subdenom string `json:"subdenom"`
}

type DenomAdminResponse struct {
	Admin string `json:"admin"`
}

type PoolState struct {
	PoolId uint64 `json:"id"`
}

type SpotPrice struct {
	Swap        Swap `json:"swap"`
	WithSwapFee bool `json:"with_swap_fee"`
}

type EstimateSwap struct {
	Sender string     `json:"sender"`
	First  Swap       `json:"first"`
	Route  []Step     `json:"route"`
	Amount SwapAmount `json:"amount"`
}

func (e *EstimateSwap) ToSwapMsg() *SwapMsg {
	return &SwapMsg{
		First:  e.First,
		Route:  e.Route,
		Amount: e.Amount.Unlimited(),
	}
}

type FullDenomResponse struct {
	Denom string `json:"denom"`
}

type PoolStateResponse struct {
	/// The various assets that be swapped. Including current liquidity.
	Assets []wasmvmtypes.Coin `json:"assets"`
	/// The number of LP shares and their amount
	Shares wasmvmtypes.Coin `json:"shares"`
}

type SpotPriceResponse struct {
	/// How many output we would get for 1 input
	Price string `json:"price"`
}

type EstimatePriceResponse struct {
	// If you query with SwapAmount::Input, this is SwapAmount::Output.
	// If you query with SwapAmount::Output, this is SwapAmount::Input.
	Amount SwapAmount `json:"swap_amount"`
}
