package cosmwasm

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
	In  string `json:"in,omitempty"`
	Out string `json:"out,omitempty"`
}

type SwapAmountWithLimit struct {
	ExactIn  *ExactIn  `json:"exact_in,omitempty"`
	ExactOut *ExactOut `json:"exact_out,omitempty"`
}

type ExactIn struct {
	Input     string `json:"input"`
	MinOutput string `json:"min_output"`
}

type ExactOut struct {
	MaxInput string `json:"max_input"`
	Output   string `json:"output"`
}
