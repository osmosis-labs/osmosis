package keeper

import (
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// iteratorAfterTime returns an iterator over all gauges in the {prefix} space of state, that begin distributing rewards after a specific time.
func (k Keeper) iteratorAfterTime(ctx sdk.Context, prefix []byte, time time.Time) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(storetypes.InclusiveEndBytes(key), storetypes.PrefixEndBytes(prefix))
}

// iteratorBeforeTime returns an iterator over all gauges in the {prefix} space of state, that begin distributing rewards before a specific time.
func (k Keeper) iteratorBeforeTime(ctx sdk.Context, prefix []byte, time time.Time) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(prefix, storetypes.InclusiveEndBytes(key))
}

// iterator returns an iterator over all gauges in the {prefix} space of state.
func (k Keeper) iterator(ctx sdk.Context, prefix []byte) storetypes.Iterator {
	store := ctx.KVStore(k.storeKey)
	return storetypes.KVStorePrefixIterator(store, prefix)
}

// UpcomingGaugesIteratorAfterTime returns the iterator to get all upcoming gauges that start distribution after a specific time.
func (k Keeper) UpcomingGaugesIteratorAfterTime(ctx sdk.Context, time time.Time) storetypes.Iterator {
	return k.iteratorAfterTime(ctx, types.KeyPrefixUpcomingGauges, time)
}

// UpcomingGaugesIteratorBeforeTime returns the iterator to get all upcoming gauges that have already started distribution before a specific time.
func (k Keeper) UpcomingGaugesIteratorBeforeTime(ctx sdk.Context, time time.Time) storetypes.Iterator {
	return k.iteratorBeforeTime(ctx, types.KeyPrefixUpcomingGauges, time)
}

// GaugesIterator returns the iterator for all gauges.
func (k Keeper) GaugesIterator(ctx sdk.Context) storetypes.Iterator {
	return k.iterator(ctx, types.KeyPrefixGauges)
}

// UpcomingGaugesIterator returns the iterator for all upcoming gauges.
func (k Keeper) UpcomingGaugesIterator(ctx sdk.Context) storetypes.Iterator {
	return k.iterator(ctx, types.KeyPrefixUpcomingGauges)
}

// ActiveGaugesIterator returns the iterator for all active gauges.
func (k Keeper) ActiveGaugesIterator(ctx sdk.Context) storetypes.Iterator {
	return k.iterator(ctx, types.KeyPrefixActiveGauges)
}

// FinishedGaugesIterator returns the iterator for all finished gauges.
func (k Keeper) FinishedGaugesIterator(ctx sdk.Context) storetypes.Iterator {
	return k.iterator(ctx, types.KeyPrefixFinishedGauges)
}

// FilterLocksByMinDuration returns locks whose lock duration is greater than the provided minimum duration.
func FilterLocksByMinDuration(locks []lockuptypes.PeriodLock, minDuration time.Duration, scratchSlice *[]*lockuptypes.PeriodLock) []*lockuptypes.PeriodLock {
	*scratchSlice = (*scratchSlice)[:0]
	filteredLocks := *scratchSlice
	for i := range locks {
		if locks[i].Duration >= minDuration {
			filteredLocks = append(filteredLocks, &locks[i])
		}
	}
	return filteredLocks
}
