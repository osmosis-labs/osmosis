package keeper

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/epochs/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of epochs module.
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		logger := k.Logger(ctx)

		// If blocktime < initial epoch start time, return
		if ctx.BlockTime().Before(epochInfo.StartTime) {
			return
		}
		// if epoch counting hasn't started, signal we need to start.
		shouldInitialEpochStart := !epochInfo.EpochCountingStarted

		epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		shouldEpochStart := (ctx.BlockTime().After(epochEndTime)) || shouldInitialEpochStart

		if !shouldEpochStart {
			return false
		}
		epochInfo.CurrentEpochStartHeight = ctx.BlockHeight()

		if shouldInitialEpochStart {
			epochInfo.EpochCountingStarted = true
			epochInfo.CurrentEpoch = 1
			epochInfo.CurrentEpochStartTime = epochInfo.StartTime
			logger.Info(fmt.Sprintf("Starting new epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
		} else {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEpochEnd,
					sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
				),
			)
			k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
			epochInfo.CurrentEpoch += 1
			epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
			logger.Info(fmt.Sprintf("Starting epoch with identifier %s epoch number %d", epochInfo.Identifier, epochInfo.CurrentEpoch))
		}

		// emit new epoch start event, set epoch info, and run BeforeEpochStart hook
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochStart,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
				sdk.NewAttribute(types.AttributeEpochStartTime, fmt.Sprintf("%d", epochInfo.CurrentEpochStartTime.Unix())),
			),
		)
		k.setEpochInfo(ctx, epochInfo)
		k.BeforeEpochStart(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)

		return false
	})
}
