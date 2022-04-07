package wasmbindings

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// This returns SwapAmountWithLimit with the largest possible limits (that will never be hit)
func (s SwapAmount) Unlimited() SwapAmountWithLimit {
	if s.In != nil {
		return SwapAmountWithLimit{
			ExactIn: &ExactIn{
				Input:     *s.In,
				MinOutput: sdk.NewInt(1),
			},
		}
	}
	if s.Out != nil {
		return SwapAmountWithLimit{
			ExactOut: &ExactOut{
				Output:   *s.Out,
				MaxInput: sdk.NewInt(math.MaxInt64),
			},
		}
	}
	panic("Must define In or Out")
}

type SwapAmountWithLimit struct {
	ExactIn  *ExactIn  `json:"exact_in,omitempty"`
	ExactOut *ExactOut `json:"exact_out,omitempty"`
}

// This returns the amount without min/max to use as simpler argument
func (s SwapAmountWithLimit) RemoveLimit() SwapAmount {
	if s.ExactIn != nil {
		return SwapAmount{
			In: &s.ExactIn.Input,
		}
	}
	if s.ExactOut != nil {
		return SwapAmount{
			Out: &s.ExactOut.Output,
		}
	}
	panic("Must define ExactIn or ExactOut")
}

type ExactIn struct {
	Input     sdk.Int `json:"input"`
	MinOutput sdk.Int `json:"min_output"`
}

type ExactOut struct {
	MaxInput sdk.Int `json:"max_input"`
	Output   sdk.Int `json:"output"`
}
