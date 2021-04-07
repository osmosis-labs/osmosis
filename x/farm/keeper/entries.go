package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

func (k Keeper) getHistoricalEntry(ctx sdk.Context, farmId uint64, period uint64) (record types.HistoricalEntry) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalEntryKey(farmId, period)

	if !store.Has(key) {
		panic(fmt.Sprintf("historical record not exist (farmId: %d, period: %d)", farmId, period))
	}

	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

func (k Keeper) setHistoricalEntry(ctx sdk.Context, farmId uint64, period uint64, entry types.HistoricalEntry) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&entry)
	store.Set(types.GetHistoricalEntryKey(farmId, period), bz)
}

func (k Keeper) IterateHistoricalEntries(ctx sdk.Context, handler func(entry types.HistoricalEntry, farmId uint64, period uint64) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.HistoricalEntryPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var historicalEntry types.HistoricalEntry
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &historicalEntry)

		key := iter.Key()
		// Extract the farm id and period from the key.
		farmIdBz := key[len(types.HistoricalEntryPrefix) : len(types.HistoricalEntryPrefix)+8]
		farmId := sdk.BigEndianToUint64(farmIdBz)

		periodBz := key[len(types.HistoricalEntryPrefix)+8 : len(types.HistoricalEntryPrefix)+16]
		period := sdk.BigEndianToUint64(periodBz)

		if handler(historicalEntry, farmId, period) {
			break
		}
	}
}
