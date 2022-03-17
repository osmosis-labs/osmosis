package wasmbindings

import sdk "github.com/cosmos/cosmos-sdk/types"

type OsmosisMsg struct {
	/// Contracts can mint native tokens that have an auto-generated denom
	/// namespaced under the contract's address. A contract may create any number
	/// of independent sub-denoms.
	MintTokens *MintTokens `json:"mint_tokens,omitempty"`
	/// Swap over one or more pools
	Swap *SwapMsg `json:"swap,omitempty"`
}

type MintTokens struct {
	/// Must be 2-32 alphanumeric characters
	SubDenom  string  `json:"sub_denom"`
	Amount    sdk.Int `json:"amount"`
	Recipient string  `json:"recipient"`
}

type SwapMsg struct {
	First  Swap                `json:"first"`
	Route  []Step              `json:"route"`
	Amount SwapAmountWithLimit `json:"amount"`
}
