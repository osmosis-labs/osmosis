package epochs

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs/keeper"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
)

// BeginBlocker of epochs module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// Note: epoch length must be more than 2 blocks, as the implementation requires epoch_endblock and epoch_startblock are separate
	// epoch_startblock(n+1) = epoch_endblock(n) + 1
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		epochStart := false
		logger := k.Logger(ctx)

		if !epochInfo.EpochCountingStarted { // epoch counting not started
			// Should start this epoch timer? (Is StartTime <= ctx.BlockTime)
			if !epochInfo.StartTime.After(ctx.BlockTime()) {
				epochStart = true
				epochInfo.EpochCountingStarted = true
				epochInfo.CurrentEpoch = 0
			}
		} else if epochInfo.CurrentEpochEnded { // epoch ended in last block
			epochStart = true
			epochInfo.CurrentEpoch = epochInfo.CurrentEpoch + 1
		}
		if epochStart {
			epochInfo.CurrentEpochEnded = false
			if epochInfo.CurrentEpochStartTime.Equal(time.Time{}) { // uninitialized genesis state, initialize
				epochInfo.StartTime = ctx.BlockTime()
				epochInfo.CurrentEpochStartTime = ctx.BlockTime()
				logger.Info(fmt.Sprintf("Starting new epoch with identifier %s", epochInfo.Identifier))
			} else {
				epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
			}
			k.SetEpochInfo(ctx, epochInfo)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEpochStart,
					sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
					sdk.NewAttribute(types.AttributeEpochStartTime, fmt.Sprintf("%d", epochInfo.CurrentEpochStartTime.Unix())),
				),
			)
		}
		return false
	})
}

// EndBlocker of epochs module
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		// if epoch counting not started, continue to next counter
		if !epochInfo.EpochCountingStarted {
			return false
		}

		// check epoch duration pass and set the current epoch ended
		nextEpochTimeEst := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		if ctx.BlockTime().Before(nextEpochTimeEst) {
			return false
		}

		k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)
		k.BeforeEpochStart(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch)

		epochInfo.CurrentEpochEnded = true
		k.SetEpochInfo(ctx, epochInfo)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochEnd,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
			),
		)
		return false
	})
}
