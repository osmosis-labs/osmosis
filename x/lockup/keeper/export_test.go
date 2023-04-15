package keeper

import (
	"github.com/osmosis-labs/osmosis/v19/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	return k.addLockRefByKey(ctx, key, lockID)
}

func (k Keeper) DeleteLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) {
	k.deleteLockRefByKey(ctx, key, lockID)
}

func (k Keeper) GetLockRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getLockRefs(ctx, key)
}

func (k Keeper) GetCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	return k.getCoinsFromLocks(locks)
}

func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock, tokensToLock sdk.Coins) error {
	return k.lock(ctx, lock, tokensToLock)
}

func (k Keeper) UnlockMaturedLockInternalLogic(ctx sdk.Context, lock types.PeriodLock) error {
	return k.unlockMaturedLockInternalLogic(ctx, lock)
}

func DurationLockRefKeys(lock types.PeriodLock) ([][]byte, error) {
	return durationLockRefKeys(lock)
}

func LockRefKeys(lock types.PeriodLock) ([][]byte, error) {
	return lockRefKeys(lock)
}

func CombineKeys(keys ...[]byte) []byte {
	return combineKeys(keys...)
}

func UnlockingPrefix(unlocking bool) []byte {
	return unlockingPrefix(unlocking)
}
