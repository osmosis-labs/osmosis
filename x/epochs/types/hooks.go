package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type EpochHooks interface {
	// the first block whose timestamp is after the duration is counted as the end of the epoch
	OnEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
	// new epoch is next block of epoch end block
	OnEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
}

var _ EpochHooks = MultiEpochHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence
type MultiEpochHooks []EpochHooks

func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

func (h MultiEpochHooks) OnEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		h[i].OnEpochEnd(ctx, epochIdentifier, epochNumber)
	}
}

func (h MultiEpochHooks) OnEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		h[i].OnEpochStart(ctx, epochIdentifier, epochNumber)
	}
}
