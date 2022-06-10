package bindings

import sdk "github.com/cosmos/cosmos-sdk/types"

type OsmosisMsg struct {
	/// Contracts can create denoms, namespaced under the contract's address.
	/// A contract may create any number of independent sub-denoms.
	CreateDenom *CreateDenom `json:"create_denom,omitempty"`
	/// Contracts can change the admin of a denom that they are the admin of.
	ChangeAdmin *ChangeAdmin `json:"change_admin,omitempty"`
	/// Contracts can mint native tokens for an existing factory denom
	/// that they are the admin of.
	MintTokens *MintTokens `json:"mint_tokens,omitempty"`
	/// Contracts can burn native tokens for an existing factory denom
	/// that they are the admin of.
	/// Currently, the burn from address must be the admin contract.
	BurnTokens *BurnTokens `json:"burn_tokens,omitempty"`
	/// Swap over one or more pools
	Swap *SwapMsg `json:"swap,omitempty"`
	/// Remove liquidity from a specified pool with an **exact** amount of LP shares while specifying the **minimum** number of tokens willing to receive for said LP shares.
	ExitPool *ExitPool `json:"exit_pool,omitempty"`
	/// Add liquidity to a specified pool to get an exact amount of LP shares while specifying a maximum number tokens
	/// willing to swap to receive said LP shares.
	JoinPoolNoSwap *JoinPoolNoSwap `json:"join_pool_no_swap,omitempty"`
	/// Add liquidity to a specified pool with only one of the required assets (i.e. Join pool 1 (50/50 ATOM-OSMO) with just ATOM).
	JoinSwapExactAmountIn *JoinSwapExactAmountIn `json:"join_swap_exact_amount_in,omitempty"`
	/// Remove an **exact** amount of LP shares from a specified pool, swap the LP shares to
	/// one of the token pairs to receive a **minimum** of the specified token amount.
	ExitSwapShareAmountIn *ExitSwapShareAmountIn `json:"exit_swap_share_amount_in,omitempty"`
}

/// CreateDenom creates a new factory denom, of denomination:
/// factory/{creating contract address}/{Subdenom}
/// Subdenom can be of length at most 44 characters, in [0-9a-zA-Z./]
/// The (creating contract address, subdenom) pair must be unique.
/// The created denom's admin is the creating contract address,
/// but this admin can be changed using the ChangeAdmin binding.
type CreateDenom struct {
	Subdenom string `json:"subdenom"`
}

/// ChangeAdmin changes the admin for a factory denom.
/// If the NewAdminAddress is empty, the denom has no admin.
type ChangeAdmin struct {
	Denom           string `json:"denom"`
	NewAdminAddress string `json:"new_admin_address"`
}

type MintTokens struct {
	Denom         string  `json:"denom"`
	Amount        sdk.Int `json:"amount"`
	MintToAddress string  `json:"mint_to_address"`
}

type BurnTokens struct {
	Denom  string  `json:"denom"`
	Amount sdk.Int `json:"amount"`
	// BurnFromAddress must be set to "" for now.
	BurnFromAddress string `json:"burn_from_address"`
}

type SwapMsg struct {
	First  Swap                `json:"first"`
	Route  []Step              `json:"route"`
	Amount SwapAmountWithLimit `json:"amount"`
}

type ExitPool struct {
	PoolId        uint64    `json:"pool_id"`
	ShareInAmount sdk.Int   `json:"share_in_amount"`
	TokenOutMins  sdk.Coins `json:"token_out_mins"`
}

type JoinPoolNoSwap struct {
	PoolId            uint64    `json:"pool_id"`
	ShareOutAmount    sdk.Int   `json:"share_out_amount"`
	TokenInMaxs       sdk.Coins `json:"token_in_maxs"`
	ShareOutMinAmount sdk.Int   `json:"share_out_min_amount"`
}

type JoinSwapExactAmountIn struct {
	PoolId            uint64   `json:"pool_id"`
	ShareOutMinAmount sdk.Int  `json:"share_out_min_amount"`
	TokenIn           sdk.Coin `json:"token_in"`
}

type ExitSwapShareAmountIn struct {
	PoolId            uint64  `json:"pool_id"`
	TokenOutDenom     string  `json:"token_out_denom"`
	ShareInAmount     sdk.Int `json:"share_in_amount"`
	TokenOutMinAmount sdk.Int `json:"token_out_min_amount"`
}
