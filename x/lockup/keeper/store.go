package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPeriodLockKey returns the key used for getting a set of period locks
// where unlockTime is after a specific time
func (k Keeper) GetPeriodLockKey(timestamp time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(types.KeyPrefixTimestamp)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], types.KeyPrefixTimestamp)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}

// GetLastLockID returns ID used last time
func (k Keeper) GetLastLockID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastLockID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastLockID save ID used by last lock
func (k Keeper) SetLastLockID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastLockID, sdk.Uint64ToBigEndian(ID))
}

// LockStoreKey returns action store key from ID
func LockStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodLock, sdk.Uint64ToBigEndian(ID))
}

// getLockTimestamp get lock IDs specified on the prefix and timestamp key
func (k Keeper) getLockTimestamp(ctx sdk.Context, key []byte) types.LockIDs {
	store := ctx.KVStore(k.storeKey)
	timeLock := types.LockIDs{}
	if store.Has(key) {
		bz := store.Get(key)
		err := json.Unmarshal(bz, &timeLock)
		if err != nil {
			panic(err)
		}
	}
	return timeLock
}

func (k Keeper) appendLockTimestamp(ctx sdk.Context, key []byte, lockID uint64) {
	// TODO: should add test for appendLockTimestamp, deleteLockTimestamp and getLockTimestamp
	store := ctx.KVStore(k.storeKey)
	timeLock := k.getLockTimestamp(ctx, key)
	timeLock.IDs = append(timeLock.IDs, lockID)
	bz, err := json.Marshal(timeLock)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

func (k Keeper) deleteLockTimestamp(ctx sdk.Context, key []byte, lockID uint64) {
	var index = -1
	store := ctx.KVStore(k.storeKey)
	timeLock := k.getLockTimestamp(ctx, key)
	timeLock.IDs, index = removeValue(timeLock.IDs, lockID)
	if index < 0 {
		panic(fmt.Sprintf("specific lock with ID %d not found", lockID))
	}
	if len(timeLock.IDs) == 0 {
		store.Delete(key)
	} else {
		bz, err := json.Marshal(timeLock)
		if err != nil {
			panic(err)
		}
		store.Set(key, bz)
	}
}

// Lock is a utility to lock coins into module account
func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock) error {
	if err := k.bk.SendCoinsFromAccountToModule(ctx, lock.Owner, types.ModuleName, lock.Coins); err != nil {
		return err
	}

	lockID := k.GetLastLockID(ctx) + 1
	store := ctx.KVStore(k.storeKey)
	store.Set(LockStoreKey(lockID), k.cdc.MustMarshalJSON(&lock))
	k.SetLastLockID(ctx, lockID)

	timeKey := k.GetPeriodLockKey(lock.EndTime)

	lockTimeKey := combineKeys(types.KeyPrefixLockTimestamp, timeKey)
	k.appendLockTimestamp(ctx, lockTimeKey, lockID)

	accLockTimeKey := combineKeys(types.KeyPrefixAccountLockTimestamp, lock.Owner, timeKey)
	k.appendLockTimestamp(ctx, accLockTimeKey, lockID)

	for _, coin := range lock.Coins {
		denomBz := []byte(coin.Denom)
		denomLockTimeKey := combineKeys(types.KeyPrefixDenomLockTimestamp, denomBz, timeKey)
		k.appendLockTimestamp(ctx, denomLockTimeKey, lockID)

		accDenomLockTimeKey := combineKeys(types.KeyPrefixAccountDenomLockTimestamp, lock.Owner, denomBz, timeKey)
		k.appendLockTimestamp(ctx, accDenomLockTimeKey, lockID)
	}
	return nil
}

// Unlock is a utility to unlock coins from module account
func (k Keeper) Unlock(ctx sdk.Context, lock types.PeriodLock) error {
	// validation for current time and unlock time
	curTime := ctx.BlockTime()
	if !curTime.After(lock.EndTime) {
		return fmt.Errorf("lock is not unlockable yet: %s >= %s", curTime.String(), lock.EndTime.String())
	}

	// send coins back to owner
	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, lock.Coins); err != nil {
		return err
	}

	lockID := lock.ID
	// remove lock from store
	store := ctx.KVStore(k.storeKey)
	store.Delete(LockStoreKey(lockID))

	// remove all timestamp query reference IDs
	timeKey := k.GetPeriodLockKey(lock.EndTime)
	lockTimeKey := combineKeys(types.KeyPrefixLockTimestamp, timeKey)
	k.deleteLockTimestamp(ctx, lockTimeKey, lockID)

	accLockTimeKey := combineKeys(types.KeyPrefixAccountLockTimestamp, lock.Owner, timeKey)
	k.deleteLockTimestamp(ctx, accLockTimeKey, lockID)

	for _, coin := range lock.Coins {
		denomBz := []byte(coin.Denom)
		denomLockTimeKey := combineKeys(types.KeyPrefixDenomLockTimestamp, denomBz, timeKey)
		k.deleteLockTimestamp(ctx, denomLockTimeKey, lockID)

		accDenomLockTimeKey := combineKeys(types.KeyPrefixAccountDenomLockTimestamp, lock.Owner, denomBz, timeKey)
		k.deleteLockTimestamp(ctx, accDenomLockTimeKey, lockID)
	}
	return nil
}
