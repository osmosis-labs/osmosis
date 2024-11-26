package keeper

import (
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// addLockRefs adds appropriate reference keys preceded by a prefix.
// A prefix indicates whether the lock is unlocking or not.
func (k Keeper) addLockRefs(ctx sdk.Context, lock types.PeriodLock) error {
	refKeys, err := durationLockRefKeys(lock)
	if lock.IsUnlocking() {
		refKeys, err = lockRefKeys(lock)
	}
	if err != nil {
		return err
	}
	lockRefPrefix := unlockingPrefix(lock.IsUnlocking())
	for _, refKey := range refKeys {
		if err := k.addLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID); err != nil {
			return err
		}
	}
	return nil
}

// deleteLockRefs deletes all the lock references of the lock with the given lock prefix.
func (k Keeper) deleteLockRefs(ctx sdk.Context, lockRefPrefix []byte, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		k.deleteLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID)
	}
	return nil
}

// addSyntheticLockRefs adds lock refs for the synthetic lock object.
func (k Keeper) addSyntheticLockRefs(ctx sdk.Context, lock types.PeriodLock, synthLock types.SyntheticLock) error {
	refKeys, err := syntheticLockRefKeys(lock, synthLock)
	if err != nil {
		return err
	}
	lockRefPrefix := unlockingPrefix(synthLock.IsUnlocking())
	for _, refKey := range refKeys {
		if err := k.addLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), synthLock.UnderlyingLockId); err != nil {
			return err
		}
	}
	return nil
}

// deleteSyntheticLockRefs deletes all lock refs for the synthetic lock object.
func (k Keeper) deleteSyntheticLockRefs(ctx sdk.Context, lock types.PeriodLock, synthLock types.SyntheticLock) error {
	refKeys, err := syntheticLockRefKeys(lock, synthLock)
	if err != nil {
		return err
	}
	lockRefPrefix := unlockingPrefix(synthLock.IsUnlocking())
	for _, refKey := range refKeys {
		k.deleteLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), synthLock.UnderlyingLockId)
	}
	return nil
}
