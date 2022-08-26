package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

type EpochHooks interface {
	// the first block whose timestamp is after the duration is counted as the end of the epoch
	AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
	// new epoch is next block of epoch end block
	BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
}

var _ EpochHooks = MultiEpochHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiEpochHooks []EpochHooks

func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the number of epoch that is ending.
func (h MultiEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		panicCatchingEpochHook(ctx, h[i].AfterEpochEnd, epochIdentifier, epochNumber)
	}
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is the number of epoch that is starting.
func (h MultiEpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		panicCatchingEpochHook(ctx, h[i].BeforeEpochStart, epochIdentifier, epochNumber)
	}
}

func panicCatchingEpochHook(
	ctx sdk.Context,
	hookFn func(ctx sdk.Context, epochIdentifier string, epochNumber int64),
	epochIdentifier string,
	epochNumber int64,
) {
	defer func() {
		if recovErr := recover(); recovErr != nil {
			osmoutils.PrintPanicRecoveryError(ctx, recovErr)
		}
	}()
	cacheCtx, write := ctx.CacheContext()
	hookFn(cacheCtx, epochIdentifier, epochNumber)
	write()
	ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
}
