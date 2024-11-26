package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

// Returns a list of newly created lockIDs, or an error.
func (k Keeper) UnpoolAllowedPools(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, lockId uint64) ([]uint64, error) {
	// Steps for unpooling for a (sender, poolID, lockID) triplet.
	// 1) Check if its for a whitelisted unpooling poolID
	// 2) Consistency check that lockID corresponds to sender, and contains correct LP shares. (Should also be validated by caller)
	// 3) Get remaining duration on the lock.
	// 4) If superfluid delegated, superfluid undelegate
	// 5) Break underlying lock. This will clear any metadata if things are superfluid unbonding
	// 6) ExitPool with these unlocked LP shares
	// 7) Make 1 new lock for every asset in collateral. Many code paths need 1 coin / lock assumption to hold
	// 8) Make new lock begin unlocking

	// 1) check if pool is whitelisted for unpool
	err := k.checkUnpoolWhitelisted(ctx, poolId)
	if err != nil {
		return []uint64{}, err
	}

	// 2) Consistency check that lockID corresponds to sender, and contains correct LP shares.
	// These are expected to be true by the caller, but good to double check
	// TODO: Try to minimize dependence on lock here
	lock, err := k.validateGammLockForSuperfluidStaking(ctx, sender, poolId, lockId)
	if err != nil {
		return []uint64{}, err
	}
	gammSharesInLock := lock.Coins[0]

	// 3) Get remaining duration on the lock. Handle if the lock was unbonding.
	lockRemainingDuration, err := k.getExistingLockRemainingDuration(ctx, lock)
	if err != nil {
		return []uint64{}, err
	}

	// 4) If superfluid delegated, superfluid undelegate
	err = k.unbondSuperfluidIfExists(ctx, sender, lockId)
	if err != nil {
		return []uint64{}, err
	}

	// 5) finish unlocking directly for locked locks
	// this also unlocks locks that were in the unlocking queue
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return []uint64{}, err
	}

	// 6) ExitPool with these unlocked LP shares
	// minOutCoins is set to 0 for now, because no sandwiching can really be done atm for UST pools
	minOutCoins := sdk.NewCoins()
	exitedCoins, err := k.gk.ExitPool(ctx, sender, poolId, gammSharesInLock.Amount, minOutCoins)
	if err != nil {
		return []uint64{}, err
	}

	// 7) Make one new lock for every coin exited from the pool.
	newLocks := make([]lockuptypes.PeriodLock, 0, len(exitedCoins))
	newLockIds := make([]uint64, 0, len(exitedCoins))
	for _, exitedCoin := range exitedCoins {
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(exitedCoin), lockRemainingDuration)
		if err != nil {
			return []uint64{}, err
		}
		newLocks = append(newLocks, newLock)
		newLockIds = append(newLockIds, newLock.ID)
	}

	// 8) Begin unlocking every new lock
	for _, newLock := range newLocks {
		_, err = k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
		if err != nil {
			return []uint64{}, err
		}
	}

	return newLockIds, nil
}

// check if pool is whitelisted for unpool
func (k Keeper) checkUnpoolWhitelisted(ctx sdk.Context, poolId uint64) error {
	allowedPools := k.GetUnpoolAllowedPools(ctx)

	for _, allowedPoolId := range allowedPools {
		if poolId == allowedPoolId {
			return nil
		}
	}

	return types.ErrPoolNotWhitelisted
}

// validateGammLockForSuperfluidStaking checks if the provided lock:
// 1) is owned by the provided sender
// 2) contains only 1 coin
// 3) contains the gamm LP shares associated with the provided poolId
func (k Keeper) validateGammLockForSuperfluidStaking(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, lockId uint64) (*lockuptypes.PeriodLock, error) {
	lock, err := k.lk.GetLockByID(ctx, lockId)
	if err != nil {
		return &lockuptypes.PeriodLock{}, err
	}

	// Validate lock owner.
	// We expect this to be guaranteed by caller, though.
	if lock.Owner != sender.String() {
		return &lockuptypes.PeriodLock{}, lockuptypes.ErrNotLockOwner
	}

	if lock.Coins.Len() != 1 {
		return &lockuptypes.PeriodLock{}, types.ErrMultipleCoinsLockupNotSupported
	}

	gammShare := lock.Coins[0]
	if gammShare.Denom != gammtypes.GetPoolShareDenom(poolId) {
		return &lockuptypes.PeriodLock{}, types.UnexpectedDenomError{ExpectedDenom: gammtypes.GetPoolShareDenom(poolId), ProvidedDenom: gammShare.Denom}
	}

	return lock, nil
}

// getExistingLockRemainingDuration returns the time remaining until the lock is finished unlocking.
// If the lock is not unlocking, then the duration field of the lock is returned.
func (k Keeper) getExistingLockRemainingDuration(ctx sdk.Context, lock *lockuptypes.PeriodLock) (time.Duration, error) {
	if lock.IsUnlocking() {
		// Lock is unlocking, so remaining duration equals lock.EndTime - ctx.BlockTime.
		remainingDuration := lock.EndTime.Sub(ctx.BlockTime())
		// Defense in depth, ensure the duration is not negative.
		if remainingDuration < 0 {
			return 0, types.NegativeDurationError{Duration: remainingDuration}
		}
		return remainingDuration, nil
	}
	// Lock is not unlocking, thus the time it should take to unlock is the locks duration.
	return lock.Duration, nil
}

// TODO: Review this in more depth
func (k Keeper) unbondSuperfluidIfExists(ctx sdk.Context, sender sdk.AccAddress, lockId uint64) error {
	// Proxy for determining if a lock is superfluid delegated. This is because, every lock that is superfluid
	// delegated, has a state entry mapping the lock ID, to an intermediary account.
	// This state entry is deleted in Superfluid undelegate, hence detects if undelegating.
	_, found := k.GetIntermediaryAccountFromLockId(ctx, lockId)
	if found {
		// superfluid undelegate first
		// this undelegates delegation, breaks synthetic locks and
		// create a new synthetic lock representing unstaking
		err := k.SuperfluidUndelegate(ctx, sender.String(), lockId)
		if err != nil {
			return err
		}
		// we don't need to call `SuperfluidUnbondLock` here as we would unlock break the lock anyways
	}
	return nil
}

func (k Keeper) GetUnpoolAllowedPools(ctx sdk.Context) []uint64 {
	store := ctx.KVStore(k.storeKey)
	allowedPools := types.UnpoolWhitelistedPools{}
	found, err := osmoutils.Get(store, types.KeyUnpoolAllowedPools, &allowedPools)
	if err != nil {
		panic(err)
	}
	if !found {
		return []uint64{}
	}
	return allowedPools.Ids
}

func (k Keeper) SetUnpoolAllowedPools(ctx sdk.Context, poolIds []uint64) {
	store := ctx.KVStore(k.storeKey)
	allowedPools := types.UnpoolWhitelistedPools{
		Ids: poolIds,
	}
	osmoutils.MustSet(store, types.KeyUnpoolAllowedPools, &allowedPools)
}
