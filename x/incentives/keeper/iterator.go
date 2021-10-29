package keeper

import (
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
)

// Returns an iterator over all gauges in the {prefix} space of state, that begin distributing rewards after a specific time
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

// UpcomingGaugesIteratorAfterTime returns the iterator to get upcoming gauges that start distribution after specific time
func (k Keeper) UpcomingGaugesIteratorAfterTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, types.KeyPrefixUpcomingGauges, time)
}

// UpcomingGaugesIteratorBeforeTime returns the iterator to get upcoming gauges that already started distribution before specific time
func (k Keeper) UpcomingGaugesIteratorBeforeTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, types.KeyPrefixUpcomingGauges, time)
}

// GaugesIterator returns iterator for all gauges
func (k Keeper) GaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixGauges)
}

// UpcomingGaugesIterator returns iterator for upcoming gauges
func (k Keeper) UpcomingGaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixUpcomingGauges)
}

// ActiveGaugesIterator returns iterator for active gauges
func (k Keeper) ActiveGaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixActiveGauges)
}

// HistoricalRewardBeforeEpochIterator returns iterator before epochNumber
func (k Keeper) HistoricalRewardBeforeEpochIterator(ctx sdk.Context, prefix []byte, epochNumber int64) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	epochKey := sdk.Uint64ToBigEndian(uint64(epochNumber))
	key := combineKeys(prefix, epochKey)
	return store.ReverseIterator(prefix, storetypes.InclusiveEndBytes(key))
}

func (k Keeper) FinishedGaugesIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixFinishedGauges)
}
