package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// check if pool is whitelisted for unpool
func (k Keeper) checkUnpoolWhitelisted(ctx sdk.Context, poolId uint64) error {
	allowedPools := k.GetUnpoolAllowedPools(ctx)
	allowed := false

	for _, allowedPoolId := range allowedPools {
		if poolId == allowedPoolId {
			allowed = true
			break
		}
	}

	if !allowed {
		return types.ErrPoolNotWhitelisted
	}

	return nil
}

// check if pool is whitelisted for unpool
func (k Keeper) validateLockForUnpool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, lockId uint64) (*lockuptypes.PeriodLock, error) {
	lock, err := k.lk.GetLockByID(ctx, lockId)
	if err != nil {
		return lock, err
	}

	// consistency check: validate lock owner
	// However, we expect this to be guaranteed by caller though.
	if lock.Owner != sender.String() {
		return lock, lockuptypes.ErrNotLockOwner
	}

	gammShare := lock.Coins[0]
	if gammShare.Denom != gammtypes.GetPoolShareDenom(poolId) {
		return lock, types.ErrLockUnpoolNotAllowed
	}

	return lock, nil
}

// Returns a list of newly created lockIDs, or an error.
func (k Keeper) UnpoolAllowedPools(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, lockId uint64) ([]uint64, error) {
	// Steps for unpooling for a (sender, poolID, lockID) triplet.
	// 0) Check if its for a whitelisted unpooling poolID
	// 1) Consistency check that lockID corresponds to sender, and contains correct LP shares. (Should also be validated by caller)
	// 2) If superfluid delegated, superfluid undelegate
	// 3) Break underlying lock. This will clear any metadata if things are superfluid unbonding
	// 3) Get duration from {} (Consider if we can handle complexity for already unbonding Locks)
	// 4) ExitPool with these unlocked LP shares
	// 5) Make 1 new lock for every asset in collateral. Many code paths need this assumption to hold
	// 6) Make new lock begin unlocking

	// 0) check if pool is whitelisted for unpool
	err := k.checkUnpoolWhitelisted(ctx, poolId)
	if err != nil {
		return []uint64{}, err
	}

	// 1) Consistency check that lockID corresponds to sender, and contains correct LP shares.
	// These are expected to be true by the caller, but good to double check
	// TODO: Try to minimize dependence on lock here
	lock, err := k.validateLockForUnpool(ctx, sender, poolId, lockId)
	if err != nil {
		return []uint64{}, err
	}

	// check if the lock is superfluid delegated
	_, found := k.GetIntermediaryAccountFromLockId(ctx, lockId)
	if found {
		// superfluid undelegate first
		// this undelegates delegation, breaks synthetic locks and
		// create a new synthetic lock representing unstaking
		err = k.SuperfluidUndelegate(ctx, sender.String(), lockId)
		if err != nil {
			return []uint64{}, err
		}
		// we don't need to call `SuperfluidUnbondLock` here as we would unlock break the lock anyways
	}

	// finish unlocking directly for locked locks
	// this also unlocks locks that were in the unlocking queue
	err = k.lk.BreakLockForUnpool(ctx, *lock)
	if err != nil {
		return []uint64{}, err
	}

	// 4) ExitPool with these unlocked LP shares
	gammShares := lock.Coins[0]
	minOutCoins := sdk.NewCoins()
	exitedCoins, err := k.gk.ExitPool(ctx, sender, poolId, gammShares.Amount, minOutCoins)
	if err != nil {
		return []uint64{}, err
	}

	newLock, err := k.lk.LockTokens(ctx, sender, exitedCoins, lock.Duration)
	if err != nil {
		return []uint64{}, err
	}

	// lock.EndTime is initialized to time.Time{} at `LockTokens` by default
	// lock.EndTime has value when the lock started unlocking
	// check if the lock was unlocking, run separate logic to preserve lock endTime
	defaultInitializedTime := time.Time{}
	if lock.EndTime != defaultInitializedTime {
		err = k.lk.BeginForceUnlockWithEndTime(ctx, newLock.ID, lock.EndTime)
		if err != nil {
			return []uint64{}, err
		}
	} else {
		err = k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
		if err != nil {
			return []uint64{}, err
		}
	}

	return []uint64{newLock.ID}, nil
}

func (k Keeper) GetUnpoolAllowedPools(ctx sdk.Context) []uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyUnpoolAllowedPools)
	if len(bz) == 0 {
		return []uint64{}
	}

	allowedPools := types.UnpoolWhitelistedPools{}
	k.cdc.MustUnmarshal(bz, &allowedPools)
	return allowedPools.Ids
}

func (k Keeper) SetUnpoolAllowedPools(ctx sdk.Context, poolIds []uint64) {
	store := ctx.KVStore(k.storeKey)

	allowedPools := types.UnpoolWhitelistedPools{
		Ids: poolIds,
	}

	bz := k.cdc.MustMarshal(&allowedPools)
	store.Set(types.KeyUnpoolAllowedPools, bz)
}
