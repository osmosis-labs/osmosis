package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

type (
	TimeTooOldError        = timeTooOldError
	TwapStrategy           = twapStrategy
	ArithmeticTwapStrategy = arithmetic
	GeometricTwapStrategy  = geometric
)

func (k Keeper) GetMostRecentRecordStoreRepresentation(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getMostRecentRecordStoreRepresentation(ctx, poolId, asset0Denom, asset1Denom)
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

func (k Keeper) UpdateRecords(ctx sdk.Context, poolId uint64) error {
	return k.updateRecords(ctx, poolId)
}

func (k Keeper) PruneRecordsBeforeTimeButNewest(ctx sdk.Context, state types.PruningState) error {
	return k.pruneRecordsBeforeTimeButNewest(ctx, state)
}

func (k Keeper) GetInterpolatedRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string, t time.Time) (types.TwapRecord, error) {
	return k.getInterpolatedRecord(ctx, poolId, t, asset0Denom, asset1Denom)
}

func ComputeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string, strategy twapStrategy) (osmomath.Dec, error) {
	return computeTwap(startRecord, endRecord, quoteAsset, strategy)
}

func (s arithmetic) ComputeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) osmomath.Dec {
	return s.computeTwap(startRecord, endRecord, quoteAsset)
}

func (s geometric) ComputeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) osmomath.Dec {
	return s.computeTwap(startRecord, endRecord, quoteAsset)
}

func RecordWithUpdatedAccumulators(record types.TwapRecord, t time.Time) types.TwapRecord {
	return recordWithUpdatedAccumulators(record, t)
}

func NewTwapRecord(k types.PoolManagerInterface, ctx sdk.Context, poolId uint64, denom0, denom1 string) (types.TwapRecord, error) {
	return newTwapRecord(k, ctx, poolId, denom0, denom1)
}

func TwapLog(x osmomath.Dec) osmomath.Dec {
	return twapLog(x)
}

// twapPow exponentiates 2 to the given exponent.
// Used as a test-helper for the power function used in geometric twap.
func TwapPow(exponent osmomath.Dec) osmomath.Dec {
	exp2 := osmomath.Exp2(osmomath.BigDecFromDec(exponent.Abs()))
	if exponent.IsNegative() {
		return osmomath.OneBigDec().Quo(exp2).Dec()
	}
	return exp2.Dec()
}

func GetSpotPrices(
	ctx sdk.Context,
	k types.PoolManagerInterface,
	poolId uint64,
	denom0, denom1 string,
	previousErrorTime time.Time,
) (sp0 osmomath.Dec, sp1 osmomath.Dec, latestErrTime time.Time) {
	return getSpotPrices(ctx, k, poolId, denom0, denom1, previousErrorTime)
}

func (k *Keeper) GetAmmInterface() types.PoolManagerInterface {
	return k.poolmanagerKeeper
}

func (k *Keeper) SetAmmInterface(poolManagerInterface types.PoolManagerInterface) {
	k.poolmanagerKeeper = poolManagerInterface
}

func (k *Keeper) AfterCreatePool(ctx sdk.Context, poolId uint64) error {
	return k.afterCreatePool(ctx, poolId)
}

func (k Keeper) GetAllHistoricalPoolIndexedTWAPs(ctx sdk.Context) ([]types.TwapRecord, error) {
	return k.getAllHistoricalPoolIndexedTWAPs(ctx)
}
