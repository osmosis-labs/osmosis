package downtimedetector

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/types"
)

func (k *Keeper) RecoveredSinceDowntimeOfLength(ctx sdk.Context, downtime types.Downtime, recoveryDuration time.Duration) (bool, error) {
	lastDowntime, err := k.GetLastDowntimeOfLength(ctx, downtime)
	if err != nil {
		return false, err
	}
	if recoveryDuration == time.Duration(0) {
		return false, errors.New("invalid recovery duration of 0")
	}
	// Check if current time < lastDowntime + recovery duration
	// if LTE, then we have not waited recovery duration.
	if ctx.BlockTime().Before(lastDowntime.Add(recoveryDuration)) {
		return false, nil
	}
	return true, nil
}
