package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (k Keeper) GetMostRecentRecordStoreRepresentation(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getMostRecentRecordStoreRepresentation(ctx, poolId, asset0Denom, asset1Denom)
}

func (k Keeper) GetAllMostRecentRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	return k.getAllMostRecentRecordsForPool(ctx, poolId)
}

func (k Keeper) GetRecordAtOrBeforeTime(ctx sdk.Context, poolId uint64, time time.Time, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getRecordAtOrBeforeTime(ctx, poolId, time, asset0Denom, asset1Denom)
}

func (k Keeper) TrackChangedPool(ctx sdk.Context, poolId uint64) {
	k.trackChangedPool(ctx, poolId)
}

func (k Keeper) GetChangedPools(ctx sdk.Context) []uint64 {
	return k.getChangedPools(ctx)
}

func (k Keeper) UpdateRecord(ctx sdk.Context, record types.TwapRecord) (types.TwapRecord, error) {
	return k.updateRecord(ctx, record)
}

func InterpolateRecord(record types.TwapRecord, t time.Time) types.TwapRecord {
	return interpolateRecord(record, t)
}
