package keeper

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

// PeriodLockIteratorAfter returns the iterator used for getting a set of locks by timestamp
// where unlock time is after a specific time
func (k Keeper) PeriodLockIteratorAfter(ctx sdk.Context, addr sdk.AccAddress, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, addr)
	key := k.GetPeriodLockKey(types.KeyPrefixLockTimestamp, time)
	return prefixStore.Iterator(key, storetypes.PrefixEndBytes(types.KeyPrefixPeriodLock))
}

// PeriodLockIteratorBefore returns the iterator used for getting a set of locks by timestamp
// where unlock time is before a specific time
func (k Keeper) PeriodLockIteratorBefore(ctx sdk.Context, addr sdk.AccAddress, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, addr)
	key := k.GetPeriodLockKey(types.KeyPrefixLockTimestamp, time)
	return prefixStore.Iterator([]byte{}, key)
}

// PeriodLockIterator returns the iterator used for getting a all locks
func (k Keeper) PeriodLockIterator(ctx sdk.Context, addr sdk.AccAddress) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, addr)
	return sdk.KVStorePrefixIterator(prefixStore, types.KeyPrefixPeriodLock)
}

// IteratorAccountsLockedPastTimeDenom Get iterator for all locks of a denom token that unlocks after timestamp
func (k Keeper) IteratorAccountsLockedPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) db.Iterator {
	return nil
}

// IteratorLockPeriodsDenom Returns all the accounts that locked coins for longer than time.Duration.  Doesn't matter how long is left until unlock.  Only based on initial locktimes
func (k Keeper) IteratorLockPeriodsDenom(ctx sdk.Context, denom string, duration time.Duration) []types.PeriodLock {
	return []types.PeriodLock{}
}
