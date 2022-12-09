package downtimedetector

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/downtime-detector/types"
)

func (k *Keeper) BeginBlock(ctx sdk.Context) {
	curTime := ctx.BlockTime()
	lastBlockTime, err := k.GetLastBlockTime(ctx)
	if err != nil {
		ctx.Logger().Error("Downtime-detector, could not get last block time, did initialization happen correctly. " + err.Error())
	}
	downtime := curTime.Sub(lastBlockTime)
	if downtime > 30*time.Second {
		k.saveDowntimeUpdates(ctx, downtime)
	}
	k.StoreLastBlockTime(ctx, curTime)
}

func (k *Keeper) saveDowntimeUpdates(ctx sdk.Context, downtime time.Duration) {
	types.DowntimeToDuration.Ascend(0, func(downType types.Downtime, duration time.Duration) bool {
		// if downtime < duration of this entry, stop iterating further, don't update this entry.
		if downtime < duration {
			return false
		}
		k.StoreLastDowntimeOfLength(ctx, downType, ctx.BlockTime())
		return true
	})
}
