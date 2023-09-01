package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

<<<<<<< HEAD
	lockuptypes "github.com/osmosis-labs/osmosis/v18/x/lockup/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

type (
	ValSet = valSet
)

func (k Keeper) ValidateLockForForceUnlock(ctx sdk.Context, lockID uint64, delegatorAddr string) (*lockuptypes.PeriodLock, osmomath.Int, error) {
	return k.validateLockForForceUnlock(ctx, lockID, delegatorAddr)
}
