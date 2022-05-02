package keeper

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Returns an iterator over all gauges in the {prefix} space of state, that begin distributing rewards after a specific time.
func (k Keeper) iteratorAfterTime(ctx sdk.Context, prefix []byte, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(storetypes.InclusiveEndBytes(key), storetypes.PrefixEndBytes(prefix))
}

func (k Keeper) iteratorBeforeTime(ctx sdk.Context, prefix []byte, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(prefix, storetypes.InclusiveEndBytes(key))
}

func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

// UpcomingGaugesIteratorAfterTime returns the iterator to get upcoming gauges that start distribution after specific time.
func (k Keeper) UpcomingGaugesIteratorAfterTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, types.KeyPrefixUpcomingGauges, time)
}

// UpcomingGaugesIteratorBeforeTime returns the iterator to get upcoming gauges that already started distribution before specific time.
func (k Keeper) UpcomingGaugesIteratorBeforeTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, types.KeyPrefixUpcomingGauges, time)
}

// GaugesIterator returns iterator for all gauges.
func (k Keeper) GaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixGauges)
}

// UpcomingGaugesIterator returns iterator for upcoming gauges.
func (k Keeper) UpcomingGaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixUpcomingGauges)
}

// ActiveGaugesIterator returns iterator for active gauges.
func (k Keeper) ActiveGaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixActiveGauges)
}

// FinishedGaugesIterator returns iterator for finished gauges.
func (k Keeper) FinishedGaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixFinishedGauges)
}

func FilterLocksByMinDuration(locks []lockuptypes.PeriodLock, minDuration time.Duration) []lockuptypes.PeriodLock {
	filteredLocks := make([]lockuptypes.PeriodLock, 0, len(locks)/2)
	for _, lock := range locks {
		if lock.Duration >= minDuration {
			filteredLocks = append(filteredLocks, lock)
		}
	}
	return filteredLocks
}
