package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

func (k Keeper) addLockRefs(ctx sdk.Context, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
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

func (k Keeper) deleteLockRefs(ctx sdk.Context, lockRefPrefix []byte, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.deleteLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID); err != nil {
			return err
		}
	}
	return nil
}

// make references for
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

func (k Keeper) deleteSyntheticLockRefs(ctx sdk.Context, lock types.PeriodLock, synthLock types.SyntheticLock) error {
	refKeys, err := syntheticLockRefKeys(lock, synthLock)
	if err != nil {
		return err
	}
	lockRefPrefix := unlockingPrefix(synthLock.IsUnlocking())
	for _, refKey := range refKeys {
		if err := k.deleteLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), synthLock.UnderlyingLockId); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) ClearAllLockRefKeys(ctx sdk.Context) {
	k.clearKeysByPrefix(ctx, types.KeyPrefixNotUnlocking)
	k.clearKeysByPrefix(ctx, types.KeyPrefixUnlocking)
}
