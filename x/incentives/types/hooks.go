package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type IncentiveHooks interface {
	AfterCreateGauge(ctx sdk.Context, gaugeID uint64)
	AfterAddToGauge(ctx sdk.Context, gaugeID uint64)
	AfterStartDistribution(ctx sdk.Context, gaugeID uint64)
	AfterFinishDistribution(ctx sdk.Context, gaugeID uint64)
	AfterEpochDistribution(ctx sdk.Context)
}

var _ IncentiveHooks = MultiIncentiveHooks{}

// combine multiple incentive hooks, all hook functions are run in array sequence.
type MultiIncentiveHooks []IncentiveHooks

func NewMultiIncentiveHooks(hooks ...IncentiveHooks) MultiIncentiveHooks {
	return hooks
}

func (h MultiIncentiveHooks) AfterCreateGauge(ctx sdk.Context, gaugeID uint64) {
	for i := range h {
		h[i].AfterCreateGauge(ctx, gaugeID)
	}
}

func (h MultiIncentiveHooks) AfterAddToGauge(ctx sdk.Context, gaugeID uint64) {
	for i := range h {
		h[i].AfterAddToGauge(ctx, gaugeID)
	}
}

func (h MultiIncentiveHooks) AfterStartDistribution(ctx sdk.Context, gaugeID uint64) {
	for i := range h {
		h[i].AfterStartDistribution(ctx, gaugeID)
	}
}

func (h MultiIncentiveHooks) AfterFinishDistribution(ctx sdk.Context, gaugeID uint64) {
	for i := range h {
		h[i].AfterFinishDistribution(ctx, gaugeID)
	}
}

func (h MultiIncentiveHooks) AfterEpochDistribution(ctx sdk.Context) {
	for i := range h {
		h[i].AfterEpochDistribution(ctx)
	}
}
