package keeper

import (
	"time"

	db "github.com/cosmos/cosmos-db"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func unlockingPrefix(isUnlocking bool) []byte {
	if isUnlocking {
		return types.KeyPrefixUnlocking
	}
	return types.KeyPrefixNotUnlocking
}

// iteratorAfterTime iterates through keys between that use prefix, and have a time.
func (k Keeper) iteratorAfterTime(ctx sdk.Context, prefix []byte, time time.Time) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	// If it’s unlockTime, then it should count as unlocked
	// inclusive end bytes = key + 1, next iterator
	return store.Iterator(storetypes.PrefixEndBytes(key), storetypes.PrefixEndBytes(prefix))
}

// iteratorBeforeTime iterates through keys between that use prefix, and have a time LTE max time.
func (k Keeper) iteratorBeforeTime(ctx sdk.Context, prefix []byte, maxTime time.Time) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(maxTime)
	key := combineKeys(prefix, timeKey)
	// If it’s unlockTime, then it should count as unlocked
	// inclusive end bytes = key + 1, next iterator
	return store.Iterator(prefix, storetypes.PrefixEndBytes(key))
}

// iteratorDuration iterates over a domain of keys for a specified duration.
func (k Keeper) iteratorDuration(ctx sdk.Context, prefix []byte, duration time.Duration) storetypes.Iterator {
	durationKey := getDurationKey(duration)
	key := combineKeys(prefix, durationKey)
	store := ctx.KVStore(k.storeKey)
	return storetypes.KVStorePrefixIterator(store, key)
}

// iteratorLongerDuration iterates over a domain of keys for longer than a specified duration.
func (k Keeper) iteratorLongerDuration(ctx sdk.Context, prefix []byte, duration time.Duration) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	durationKey := getDurationKey(duration)
	key := combineKeys(prefix, durationKey)
	// inclusive on longer side, means >= (longer or equal)
	return store.Iterator(key, storetypes.PrefixEndBytes(prefix))
}

// iteratorShorterDuration iterates over a domain of keys for shorter than a specified duration.
func (k Keeper) iteratorShorterDuration(ctx sdk.Context, prefix []byte, duration time.Duration) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	durationKey := getDurationKey(duration)
	key := combineKeys(prefix, durationKey)
	// inclusive on longer side, shorter means < (lower)
	return store.Iterator(prefix, key)
}

// iterator iterates over a domain of keys.
func (k Keeper) iterator(ctx sdk.Context, prefix []byte) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	return storetypes.KVStorePrefixIterator(store, prefix)
}

// LockIteratorAfterTime returns the iterator to get locked coins.
func (k Keeper) LockIteratorAfterTime(ctx sdk.Context, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorAfterTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixLockTimestamp), time)
}

// LockIteratorBeforeTime returns the iterator to get unlockable coins.
func (k Keeper) LockIteratorBeforeTime(ctx sdk.Context, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorBeforeTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixLockTimestamp), time)
}

// LockIterator returns the iterator used for getting all locks.
func (k Keeper) LockIterator(ctx sdk.Context, isUnlocking bool) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iterator(ctx, combineKeys(unlockingPrefix, types.KeyPrefixLockDuration))
}

// LockIteratorAfterTimeDenom returns the iterator to get locked coins by denom.
func (k Keeper) LockIteratorAfterTimeDenom(ctx sdk.Context, denom string, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorAfterTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixDenomLockTimestamp, []byte(denom)), time)
}

// LockIteratorBeforeTimeDenom returns the iterator to get unlockable coins by denom.
func (k Keeper) LockIteratorBeforeTimeDenom(ctx sdk.Context, denom string, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorBeforeTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixDenomLockTimestamp, []byte(denom)), time)
}

// LockIteratorLongerThanDurationDenom returns the iterator to get locked locks by denom.
func (k Keeper) LockIteratorLongerThanDurationDenom(ctx sdk.Context, isUnlocking bool, denom string, duration time.Duration) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iteratorLongerDuration(ctx, combineKeys(unlockingPrefix, types.KeyPrefixDenomLockDuration, []byte(denom)), duration)
}

// LockIteratorDenom returns the iterator used for getting all locks by denom.
func (k Keeper) LockIteratorDenom(ctx sdk.Context, isUnlocking bool, denom string) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iterator(ctx, combineKeys(unlockingPrefix, types.KeyPrefixDenomLockDuration, []byte(denom)))
}

// AccountLockIteratorAfterTime returns the iterator to get locked coins by account.
func (k Keeper) AccountLockIteratorAfterTime(ctx sdk.Context, addr sdk.AccAddress, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorAfterTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountLockTimestamp, addr), time)
}

// AccountLockIteratorBeforeTime returns the iterator to get unlockable coins by account.
func (k Keeper) AccountLockIteratorBeforeTime(ctx sdk.Context, addr sdk.AccAddress, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorBeforeTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountLockTimestamp, addr), time)
}

// AccountLockIterator returns the iterator used for getting all locks by account.
func (k Keeper) AccountLockIterator(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iterator(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountLockDuration, addr))
}

// AccountLockIteratorAfterTimeDenom returns the iterator to get locked coins by account and denom.
func (k Keeper) AccountLockIteratorAfterTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorAfterTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountDenomLockTimestamp, addr, []byte(denom)), time)
}

// AccountLockIteratorBeforeTimeDenom returns the iterator to get unlockable coins by account and denom.
func (k Keeper) AccountLockIteratorBeforeTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, time time.Time) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(true)
	return k.iteratorBeforeTime(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountDenomLockTimestamp, addr, []byte(denom)), time)
}

// AccountLockIteratorDenom returns the iterator used for getting all locks by account and denom.
func (k Keeper) AccountLockIteratorDenom(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress, denom string) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iterator(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountDenomLockDuration, addr, []byte(denom)))
}

// AccountLockIteratorLongerDuration returns iterator used for getting all locks by account longer than duration.
func (k Keeper) AccountLockIteratorLongerDuration(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress, duration time.Duration) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iteratorLongerDuration(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountLockDuration, addr), duration)
}

// AccountLockIteratorDuration returns an iterator used for getting all locks for a given account, isUnlocking, and specific duration.
func (k Keeper) AccountLockIteratorDuration(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress, duration time.Duration) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iteratorDuration(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountLockDuration, addr), duration)
}

// AccountLockIteratorShorterThanDuration returns an iterator used for getting all locks by account shorter than the specified duration.
func (k Keeper) AccountLockIteratorShorterThanDuration(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress, duration time.Duration) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iteratorShorterDuration(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountLockDuration, addr), duration)
}

// AccountLockIteratorLongerDurationDenom returns iterator used for getting all locks by account and denom longer than duration.
func (k Keeper) AccountLockIteratorLongerDurationDenom(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress, denom string, duration time.Duration) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iteratorLongerDuration(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountDenomLockDuration, addr, []byte(denom)), duration)
}

// AccountLockIteratorDurationDenom returns iterator used for getting all locks by account and denom with specific duration.
func (k Keeper) AccountLockIteratorDurationDenom(ctx sdk.Context, isUnlocking bool, addr sdk.AccAddress, denom string, duration time.Duration) storetypes.Iterator {
	unlockingPrefix := unlockingPrefix(isUnlocking)
	return k.iteratorDuration(ctx, combineKeys(unlockingPrefix, types.KeyPrefixAccountDenomLockDuration, addr, []byte(denom)), duration)
}

// getLocksFromIterator returns an array of single lock unit by period defined by the x/lockup module.
func (k Keeper) getLocksFromIterator(ctx sdk.Context, iterator db.Iterator) []types.PeriodLock {
	locks := []types.PeriodLock{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		lockID := sdk.BigEndianToUint64(iterator.Value())
		lock, err := k.GetLockByID(ctx, lockID)
		if err != nil {
			panic(err)
		}
		locks = append(locks, *lock)
	}
	return locks
}

// unlockFromIterator gets locks from the iterator, then unlocks all matured locks. Returns locks unlocked and sum of coins unlocked.
func (k Keeper) unlockFromIterator(ctx sdk.Context, iterator db.Iterator) ([]types.PeriodLock, sdk.Coins) {
	// Note: this function is only used for an account
	// and this has no conflicts with synthetic lockups

	coins := sdk.Coins{}
	locks := k.getLocksFromIterator(ctx, iterator)
	for _, lock := range locks {
		err := k.UnlockMaturedLock(ctx, lock.ID)
		if err != nil {
			panic(err)
		}
		// sum up all coins unlocked
		coins = coins.Add(lock.Coins...)
	}
	return locks, coins
}

// beginUnlockFromIterator starts unlocking coins from NotUnlocking queue.
func (k Keeper) beginUnlockFromIterator(ctx sdk.Context, iterator db.Iterator) ([]types.PeriodLock, error) {
	// Note: this function is only used for an account
	// and this has no conflicts with synthetic lockups

	locks := k.getLocksFromIterator(ctx, iterator)
	for _, lock := range locks {
		_, err := k.BeginUnlock(ctx, lock.ID, nil)
		if err != nil {
			return locks, err
		}
	}
	return locks, nil
}

// getCoinsFromIterator gets coins from locks using the iterator.
func (k Keeper) getCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromLocks(k.getLocksFromIterator(ctx, iterator))
}
