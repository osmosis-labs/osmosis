package wasmbindings

import sdk "github.com/cosmos/cosmos-sdk/types"

type OsmosisMsg struct {
	/// Contracts can create denoms, namespaced under the contract's address.
	//A contract may create any number of independent sub-denoms.
	CreateDenom *CreateDenom `json:"create_denom,omitempty"`
	/// Contracts can mint native tokens for an existing denom
	/// namespaced under the contract's address.
	MintTokens *MintTokens `json:"mint_tokens,omitempty"`
	/// Swap over one or more pools
	Swap *SwapMsg `json:"swap,omitempty"`
}

type CreateDenom struct {
	/// Subdenoms (nonces) are validated as part of the full denomination.
	/// Can be up to 128 - prefix length (currently 7) - bech32 address length (4 (osmo) + 39) - number of separators (2) =
	/// 76 "alphanumeric" (https://github.com/cosmos/cosmos-sdk/blob/2646b474c7beb0c93d4fafd395ef345f41afc251/types/coin.go#L677)
	/// characters long.
	/// Empty subdenoms are valid. The token will then be prefix + contract address, i.e. "factory/<bech32 address>/"
	Subdenom string `json:"subdenom"`
}

type MintTokens struct {
	Subdenom  string  `json:"subdenom"`
	Amount    sdk.Int `json:"amount"`
	Recipient string  `json:"recipient"`
}

type SwapMsg struct {
	First  Swap                `json:"first"`
	Route  []Step              `json:"route"`
	Amount SwapAmountWithLimit `json:"amount"`
}
