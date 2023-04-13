package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

var (
	StakingSyntheticDenom   = stakingSyntheticDenom
	UnstakingSyntheticDenom = unstakingSyntheticDenom
)

func (k Keeper) ValidateLockForSFDelegate(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	return k.validateLockForSFDelegate(ctx, lock, sender)
}
