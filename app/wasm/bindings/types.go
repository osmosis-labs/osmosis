package cosmwasm

import sdk "github.com/cosmos/cosmos-sdk/types"

type Swap struct {
	PoolId   uint64 `json:"pool_id"`
	DenomIn  string `json:"denom_in"`
	DenomOut string `json:"denom_out"`
}

type Step struct {
	PoolId   uint64 `json:"pool_id"`
	DenomOut string `json:"denom_out"`
}

type SwapAmount struct {
	In  *sdk.Int `json:"in,omitempty"`
	Out *sdk.Int `json:"out,omitempty"`
}

type SwapAmountWithLimit struct {
	ExactIn  *ExactIn  `json:"exact_in,omitempty"`
	ExactOut *ExactOut `json:"exact_out,omitempty"`
}

type ExactIn struct {
	Input     sdk.Int `json:"input"`
	MinOutput sdk.Int `json:"min_output"`
}

type ExactOut struct {
	MaxInput sdk.Int `json:"max_input"`
	Output   sdk.Int `json:"output"`
}
