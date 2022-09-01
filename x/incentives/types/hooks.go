package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

type IncentiveHooks interface {
	AfterCreateGauge(ctx sdk.Context, gaugeId uint64) error
	AfterAddToGauge(ctx sdk.Context, gaugeId uint64) error
	AfterStartDistribution(ctx sdk.Context, gaugeId uint64) error
	AfterFinishDistribution(ctx sdk.Context, gaugeId uint64) error
	AfterEpochDistribution(ctx sdk.Context) error
}

var _ IncentiveHooks = MultiIncentiveHooks{}

// MultiIncentiveHooks combines multiple incentive hooks. All hook functions are run in array sequence.
type MultiIncentiveHooks []IncentiveHooks

// NewMultiIncentiveHooks combines multiple incentive hooks into a single IncentiveHooks array.
func NewMultiIncentiveHooks(hooks ...IncentiveHooks) MultiIncentiveHooks {
	return hooks
}

func (h MultiIncentiveHooks) AfterCreateGauge(ctx sdk.Context, gaugeId uint64) error {
	for i := range h {
		errorCatchingIncentiveHook(ctx, h[i].AfterCreateGauge, gaugeId)
	}
	return nil
}

func (h MultiIncentiveHooks) AfterAddToGauge(ctx sdk.Context, gaugeId uint64) error {
	for i := range h {
		errorCatchingIncentiveHook(ctx, h[i].AfterAddToGauge, gaugeId)
	}
	return nil
}

func (h MultiIncentiveHooks) AfterStartDistribution(ctx sdk.Context, gaugeId uint64) error {
	for i := range h {
		errorCatchingIncentiveHook(ctx, h[i].AfterStartDistribution, gaugeId)
	}
	return nil
}

func (h MultiIncentiveHooks) AfterFinishDistribution(ctx sdk.Context, gaugeId uint64) error {
	for i := range h {
		errorCatchingIncentiveHook(ctx, h[i].AfterFinishDistribution, gaugeId)
	}
	return nil
}

func (h MultiIncentiveHooks) AfterEpochDistribution(ctx sdk.Context) error {
	for i := range h {
		err := osmoutils.ApplyFuncIfNoError(ctx, h[i].AfterEpochDistribution)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("error in incentive hook %v", err))
		}
	}
	return nil
}

func errorCatchingIncentiveHook(
	ctx sdk.Context,
	hookFn func(ctx sdk.Context, gaugeId uint64) error,
	gaugeId uint64,
) {
	wrappedHookFn := func(ctx sdk.Context) error {
		return hookFn(ctx, gaugeId)
	}

	err := osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("error in incentive hook %v", err))
	}
}
