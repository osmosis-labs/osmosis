package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
)

type (
	ValSet = valSet
)

func (k Keeper) ValidateLockForForceUnlock(ctx sdk.Context, lockID uint64, delegatorAddr string) (*lockuptypes.PeriodLock, sdk.Int, error) {
	return k.validateLockForForceUnlock(ctx, lockID, delegatorAddr)
}
