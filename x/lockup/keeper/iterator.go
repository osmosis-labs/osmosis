package keeper

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) iteratorAfterTime(ctx sdk.Context, prefix []byte, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := k.GetPeriodLockKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(key, storetypes.PrefixEndBytes(types.KeyPrefixPeriodLock))
}

func (k Keeper) iteratorBeforeTime(ctx sdk.Context, prefix []byte, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := k.GetPeriodLockKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(prefix, key)
}

func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

// LockIteratorAfterTime returns the iterator to get locked coins
func (k Keeper) LockIteratorAfterTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, types.KeyPrefixLockTimestamp, time)
}

// LockIteratorBeforeTime returns the iterator to get unlockable coins
func (k Keeper) LockIteratorBeforeTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, types.KeyPrefixLockTimestamp, time)
}

// LockIterator returns the iterator used for getting all locks
func (k Keeper) LockIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixLockTimestamp)
}

// LockIteratorAfterTimeDenom returns the iterator to get locked coins by denom
func (k Keeper) LockIteratorAfterTimeDenom(ctx sdk.Context, denom string, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, combineKeys(types.KeyPrefixDenomLockTimestamp, []byte(denom)), time)
}

// LockIteratorBeforeTimeDenom returns the iterator to get unlockable coins by denom
func (k Keeper) LockIteratorBeforeTimeDenom(ctx sdk.Context, denom string, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixDenomLockTimestamp, []byte(denom)), time)
}

// LockIteratorDenom returns the iterator used for getting all locks by denom
func (k Keeper) LockIteratorDenom(ctx sdk.Context, denom string) sdk.Iterator {
	return k.iterator(ctx, combineKeys(types.KeyPrefixDenomLockTimestamp, []byte(denom)))
}

// AccountLockIteratorAfterTime returns the iterator to get locked coins by account
func (k Keeper) AccountLockIteratorAfterTime(ctx sdk.Context, acc sdk.AccAddress, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, combineKeys(types.KeyPrefixAccountLockTimestamp, acc), time)
}

// AccountLockIteratorBeforeTime returns the iterator to get unlockable coins by account
func (k Keeper) AccountLockIteratorBeforeTime(ctx sdk.Context, acc sdk.AccAddress, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixAccountLockTimestamp, acc), time)
}

// AccountLockIterator returns the iterator used for getting all locks by account
func (k Keeper) AccountLockIterator(ctx sdk.Context, acc sdk.AccAddress) sdk.Iterator {
	return k.iterator(ctx, combineKeys(types.KeyPrefixAccountLockTimestamp, acc))
}

// AccountLockIteratorAfterTimeDenom returns the iterator to get locked coins by account and denom
func (k Keeper) AccountLockIteratorAfterTimeDenom(ctx sdk.Context, acc sdk.AccAddress, denom string, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, acc, []byte(denom)), time)
}

// AccountLockIteratorBeforeTimeDenom returns the iterator to get unlockable coins by account and denom
func (k Keeper) AccountLockIteratorBeforeTimeDenom(ctx sdk.Context, acc sdk.AccAddress, denom string, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, acc, []byte(denom)), time)
}

// AccountLockIteratorDenom returns the iterator used for getting all locks by account and denom
func (k Keeper) AccountLockIteratorDenom(ctx sdk.Context, acc sdk.AccAddress, denom string) sdk.Iterator {
	return k.iterator(ctx, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, acc, []byte(denom)))
}
