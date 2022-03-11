package cosmwasm

// OsmosisQuery contains osmosis custom queries.
// See https://github.com/confio/osmosis-bindings/blob/main/packages/bindings/src/query.rs
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
	EstimatePrice *EstimatePrice `json:"estimate_price,omitempty"`
}

type FullDenom struct {
	Contract string `json:"contract"`
	SubDenom string `json:"sub_denom"`
}

type PoolState struct {
	PoolId uint64 `json:"id"`
}

type SpotPrice struct {
	Swap        Swap `json:"swap"`
	WithSwapFee bool `json:"with_swap_fee"`
}

type EstimatePrice struct {
	First  Swap
	Route  []Step
	Amount SwapAmount
}
