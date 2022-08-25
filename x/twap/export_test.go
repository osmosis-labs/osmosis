package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

func (k Keeper) StoreNewRecord(ctx sdk.Context, record types.TwapRecord) {
	k.storeNewRecord(ctx, record)
}

func (k Keeper) GetMostRecentRecordStoreRepresentation(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getMostRecentRecordStoreRepresentation(ctx, poolId, asset0Denom, asset1Denom)
}

func (k Keeper) GetAllMostRecentRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.TwapRecord, error) {
	return k.getAllMostRecentRecordsForPool(ctx, poolId)
}

func (k Keeper) GetRecordAtOrBeforeTime(ctx sdk.Context, poolId uint64, time time.Time, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getRecordAtOrBeforeTime(ctx, poolId, time, asset0Denom, asset1Denom)
}

func (k Keeper) GetAllHistoricalTimeIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return k.getAllHistoricalTimeIndexedTWAPs(ctx)
}

func (k Keeper) GetAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return k.getAllHistoricalPoolIndexedTWAPs(ctx)
}

func (k Keeper) TrackChangedPool(ctx sdk.Context, poolId uint64) {
	k.trackChangedPool(ctx, poolId)
}

func (k Keeper) GetChangedPools(ctx sdk.Context) []uint64 {
	return k.getChangedPools(ctx)
}

func (k Keeper) UpdateRecord(ctx sdk.Context, record types.TwapRecord) types.TwapRecord {
	return k.updateRecord(ctx, record)
}

func (k Keeper) PruneRecordsBeforeTime(ctx sdk.Context, lastTime time.Time) error {
	return k.pruneRecordsBeforeTime(ctx, lastTime)
}

func (k Keeper) PruneRecords(ctx sdk.Context) error {
	return k.pruneRecords(ctx)
}

func ComputeArithmeticTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	return computeArithmeticTwap(startRecord, endRecord, quoteAsset)
}

func RecordWithUpdatedAccumulators(record types.TwapRecord, t time.Time) types.TwapRecord {
	return recordWithUpdatedAccumulators(record, t)
}
