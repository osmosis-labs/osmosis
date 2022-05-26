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
	/// Add liquidity to a specified pool to get an exact amount of LP shares while specifying a maximum number tokens
	/// willing to swap to receive said LP shares.
	JoinPool *JoinPool `json:"join_pool,omitempty"`
}

type CreateDenom struct {
	/// Sub_denoms (nonces) are validated as part of the full denomination.
	/// Can be up to 128 - prefix length (currently 7) - bech32 address length (4 (osmo) + 39) - number of separators (2) =
	/// 76 "alphanumeric" (https://github.com/cosmos/cosmos-sdk/blob/2646b474c7beb0c93d4fafd395ef345f41afc251/types/coin.go#L677)
	/// characters long.
	/// Empty sub-denoms are valid. The token will then be prefix + contract address, i.e. "factory/<bech32 address>/"
	SubDenom string `json:"sub_denom"`
}

type MintTokens struct {
	SubDenom  string  `json:"sub_denom"`
	Amount    sdk.Int `json:"amount"`
	Recipient string  `json:"recipient"`
}

type SwapMsg struct {
	First  Swap                `json:"first"`
	Route  []Step              `json:"route"`
	Amount SwapAmountWithLimit `json:"amount"`
}

type JoinPool struct {
	PoolId         uint64    `json:"pool_id"`
	ShareOutAmount sdk.Int   `json:"share_out_amount"`
	TokenInMaxs    sdk.Coins `json:"token_in_maxs"`
}
