package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

func (k Keeper) GetHistoricalRecord(ctx sdk.Context, farmId uint64, period uint64) (record types.HistoricalRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetHistoricalRecord(farmId, period))
	if len(bz) == 0 {
		panic(fmt.Sprintf("historical record not exist (farmId: %d, period: %d)", farmId, period))
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

func (k Keeper) SetHistoricalRecord(ctx sdk.Context, farmId uint64, period uint64, record types.HistoricalRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&record)
	store.Set(types.GetHistoricalRecord(farmId, period), bz)
}

func (k Keeper) IterateHistoricalRecords(ctx sdk.Context, handler func(record types.HistoricalRecord, farmId uint64, period uint64) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.HistoricalRecordPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var historicalRecord types.HistoricalRecord
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &historicalRecord)

		key := iter.Key()
		// Extract the farm id and period from the key.
		farmIdBz := key[len(types.HistoricalRecordPrefix) : len(types.HistoricalRecordPrefix)+8]
		farmId := sdk.BigEndianToUint64(farmIdBz)

		periodBz := key[len(types.HistoricalRecordPrefix)+8 : len(types.HistoricalRecordPrefix)+16]
		period := sdk.BigEndianToUint64(periodBz)

		if handler(historicalRecord, farmId, period) {
			break
		}
	}
}
