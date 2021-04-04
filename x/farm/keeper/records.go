package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

func (k Keeper) GetHistoricalRecord(ctx sdk.Context, farmId uint64, period int64) (record types.HistoricalRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetHistoricalRecord(farmId, period))
	if len(bz) == 0 {
		panic(fmt.Sprintf("historical record not exist (farmId: %d, period: %d)", farmId, period))
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

func (k Keeper) SetHistoricalRecord(ctx sdk.Context, farmId uint64, period int64, record types.HistoricalRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&record)
	store.Set(types.GetHistoricalRecord(farmId, period), bz)
}
