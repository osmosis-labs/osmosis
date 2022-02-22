package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

func (k Keeper) AddUnlockBlocker(ctx sdk.Context, lockID uint64, blockerKey string) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	for _, existingBlocker := range lock.UnlockBlockers {
		if existingBlocker == blockerKey {
			return types.ErrUnlockBlockerAlreadyAdded
		}
	}

	lock.UnlockBlockers = append(lock.UnlockBlockers, blockerKey)
	err = k.setLock(ctx, *lock)
	return err
}

func (k Keeper) RemoveUnlockBlocker(ctx sdk.Context, lockID uint64, blockerKey string) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	for i, existingBlocker := range lock.UnlockBlockers {
		if existingBlocker == blockerKey {
			lock.UnlockBlockers = append(lock.UnlockBlockers[:i], lock.UnlockBlockers[i+1:]...)
			err = k.setLock(ctx, *lock)
			return err
		}
	}
	return nil
}
