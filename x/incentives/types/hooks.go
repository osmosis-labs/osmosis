package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type IncentiveHooks interface {
	AfterCreateGauge(ctx sdk.Context, gaugeId uint64)
	AfterAddToGauge(ctx sdk.Context, gaugeId uint64)
	AfterStartDistribution(ctx sdk.Context, gaugeId uint64)
	AfterFinishDistribution(ctx sdk.Context, gaugeId uint64)
	AfterEpochDistribution(ctx sdk.Context)
}

var _ IncentiveHooks = MultiIncentiveHooks{}

// MultiIncentiveHooks combines multiple incentive hooks. All hook functions are run in array sequence.
type MultiIncentiveHooks []IncentiveHooks

// NewMultiIncentiveHooks combines multiple incentive hooks into a single IncentiveHooks array.
func NewMultiIncentiveHooks(hooks ...IncentiveHooks) MultiIncentiveHooks {
	return hooks
}

func (h MultiIncentiveHooks) AfterCreateGauge(ctx sdk.Context, gaugeId uint64) {
	for i := range h {
		h[i].AfterCreateGauge(ctx, gaugeId)
	}
}

func (h MultiIncentiveHooks) AfterAddToGauge(ctx sdk.Context, gaugeId uint64) {
	for i := range h {
		h[i].AfterAddToGauge(ctx, gaugeId)
	}
}

func (h MultiIncentiveHooks) AfterStartDistribution(ctx sdk.Context, gaugeId uint64) {
	for i := range h {
		h[i].AfterStartDistribution(ctx, gaugeId)
	}
}

func (h MultiIncentiveHooks) AfterFinishDistribution(ctx sdk.Context, gaugeId uint64) {
	for i := range h {
		h[i].AfterFinishDistribution(ctx, gaugeId)
	}
}

func (h MultiIncentiveHooks) AfterEpochDistribution(ctx sdk.Context) {
	for i := range h {
		h[i].AfterEpochDistribution(ctx)
	}
}
