package downtimedetector

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

func (k *Keeper) BeginBlock(ctx sdk.Context) {
	curTime := ctx.BlockTime()
	lastBlockTime, err := k.GetLastBlockTime(ctx)
	if err != nil {
		ctx.Logger().Error("Downtime-detector, could not get last block time, did initialization happen correctly. " + err.Error())
	}
	downtime := curTime.Sub(lastBlockTime)
	k.saveDowntimeUpdates(ctx, downtime)
	k.StoreLastBlockTime(ctx, curTime)
}

// saveDowntimeUpdates saves the current block time as the
// last time the chain was down for all downtime lengths that are LTE the provided downtime.
func (k *Keeper) saveDowntimeUpdates(ctx sdk.Context, downtime time.Duration) {
	// minimum stored downtime is 30S, so if downtime is less than that, don't update anything.
	if downtime < 30*time.Second {
		return
	}
	types.DowntimeToDuration.Ascend(0, func(downType types.Downtime, duration time.Duration) bool {
		// if downtime < duration of this entry, stop iterating further, don't update this entry.
		if downtime < duration {
			return false
		}
		k.StoreLastDowntimeOfLength(ctx, downType, ctx.BlockTime())
		return true
	})
}
