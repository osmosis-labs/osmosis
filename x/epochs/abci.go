package epochs

import (
	"fmt"
	"time"

	"github.com/c-osmosis/osmosis/x/epochs/keeper"
	"github.com/c-osmosis/osmosis/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of epochs module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// TODO: should we run epoch start on begin blocker?
}

// EndBlocker of epochs module
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.Epoch) (stop bool) {
		nextEpochTimeEst := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		if ctx.BlockTime().Before(nextEpochTimeEst) {
			return false
		}

		k.OnEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochEnd,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
			),
		)

		// TODO: epoch start should be always begin blocker? for now we do on endblocker
		epochInfo.CurrentEpoch = epochInfo.CurrentEpoch + 1
		epochInfo.CurrentEpochStartTime = ctx.BlockTime()
		k.OnEpochStart(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochStart,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
				sdk.NewAttribute(types.AttributeEpochStartTime, fmt.Sprintf("%d", epochInfo.CurrentEpochStartTime.Unix())),
			),
		)
		return false
	})
}
