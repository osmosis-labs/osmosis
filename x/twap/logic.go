package twap

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

func newTwapRecord(k types.PoolManagerInterface, ctx sdk.Context, poolId uint64, denom0, denom1 string) (types.TwapRecord, error) {
	denom0, denom1, err := types.LexicographicalOrderDenoms(denom0, denom1)
	if err != nil {
		return types.TwapRecord{}, err
	}
	previousErrorTime := time.Time{} // no previous error
	sp0, sp1, lastErrorTime := getSpotPrices(ctx, k, poolId, denom0, denom1, previousErrorTime)
	return types.TwapRecord{
		PoolId:                      poolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom1,
		Height:                      ctx.BlockHeight(),
		Time:                        ctx.BlockTime(),
		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sp1,
		P0ArithmeticTwapAccumulator: osmomath.ZeroDec(),
		P1ArithmeticTwapAccumulator: osmomath.ZeroDec(),
		GeometricTwapAccumulator:    osmomath.ZeroDec(),
		LastErrorTime:               lastErrorTime,
	}, nil
}

// getSpotPrices gets the spot prices for the pool,
// input: ctx, amm interface, pool id, asset denoms, previous error time
// returns spot prices for both pairs of assets, and the 'latest error time'.
// The latest error time is the previous time if there is no error in getting spot prices.
// if there is an error in getting spot prices, then the latest error time is ctx.Blocktime()
func getSpotPrices(
	ctx sdk.Context,
	k types.PoolManagerInterface,
	poolId uint64,
	denom0, denom1 string,
	previousErrorTime time.Time,
) (sp0 osmomath.Dec, sp1 osmomath.Dec, latestErrTime time.Time) {
	latestErrTime = previousErrorTime
	// sp0 = denom0 quote, denom1 base.
	sp0BigDec, err0 := k.RouteCalculateSpotPrice(ctx, poolId, denom0, denom1)
	// sp1 = denom0 base, denom1 quote.
	sp1BigDec, err1 := k.RouteCalculateSpotPrice(ctx, poolId, denom1, denom0)

	if err0 != nil || err1 != nil {
		latestErrTime = ctx.BlockTime()
		// In the event of an error, we just sanity replace empty values with zero values
		// so that the numbers can be still be calculated within TWAPs over error values
		// TODO: Should we be using the last spot price?
		if (sp0BigDec == osmomath.BigDec{}) {
			sp0BigDec = osmomath.ZeroBigDec()
		}
		if (sp1BigDec == osmomath.BigDec{}) {
			sp1BigDec = osmomath.ZeroBigDec()
		}
	}
	if sp0BigDec.GT(types.MaxSpotPriceBigDec) {
		sp0BigDec, latestErrTime = types.MaxSpotPriceBigDec, ctx.BlockTime()
	}
	if sp1BigDec.GT(types.MaxSpotPriceBigDec) {
		sp1BigDec, latestErrTime = types.MaxSpotPriceBigDec, ctx.BlockTime()
	}

	// Note: truncation is appropriate since we don not support greater precision by design.
	// If support for pools with outlier spot prices is needed in the future, then this should be revisited.
	return sp0BigDec.Dec(), sp1BigDec.Dec(), latestErrTime
}

// mustTrackCreatedPool is a wrapper around afterCreatePool that panics on error.
func (k Keeper) mustTrackCreatedPool(ctx sdk.Context, poolId uint64) {
	err := k.afterCreatePool(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

// afterCreatePool creates new twap records of all the unique pairs of denoms within a pool.
func (k Keeper) afterCreatePool(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.poolmanagerKeeper.RouteGetPoolDenoms(ctx, poolId)
	denomPairs := types.GetAllUniqueDenomPairs(denoms)
	for _, denomPair := range denomPairs {
		record, err := newTwapRecord(k.poolmanagerKeeper, ctx, poolId, denomPair.Denom0, denomPair.Denom1)
		// err should be impossible given GetAllUniqueDenomPairs guarantees
		if err != nil {
			return err
		}
		// we create a record here, because we need the record to exist in the event
		// that there is a swap against this pool in this same block.
		// furthermore, this protects against an edge case where a pool is created
		// during EndBlock, after twapkeeper's endblock.
		k.StoreNewRecord(ctx, record)
	}
	k.trackChangedPool(ctx, poolId)
	return err
}

func (k Keeper) EndBlock(ctx sdk.Context) {
	// get changed pools grabs all altered pool ids from the transient store.
	// 'altered pool ids' gets automatically cleared on commit by being a transient store
	changedPoolIds := k.getChangedPools(ctx)
	for _, id := range changedPoolIds {
		err := k.updateRecords(ctx, id)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf(
				"error in TWAP end block, for updating records for pool id %d."+
					" Skipping record update. Underlying err: %w", id, err).Error())
		}
	}

	state := k.GetPruningState(ctx)
	if state.IsPruning {
		err := k.pruneRecordsBeforeTimeButNewest(ctx, state)
		if err != nil {
			ctx.Logger().Error("Error pruning old twaps at the end block", err)
		}
	}
}

// updateRecords updates all records for a given pool id.
// it does so by creating new records for all asset pairs
// with updated spot prices and spot price errors, if any.
// Returns nil on success.
// Returns error if:
//   - fails to get previous records.
//   - fails to get denoms from the pool.
//   - the number of records does not match expected relative to the
//     number of denoms in the pool.
func (k Keeper) updateRecords(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.poolmanagerKeeper.RouteGetPoolDenoms(ctx, poolId)
	if err != nil {
		return err
	}

	// Will only err if pool doesn't have most recent entry set
	records, err := k.GetAllMostRecentRecordsForPoolWithDenoms(ctx, poolId, denoms)
	if err != nil {
		return err
	}

	// given # of denoms in the pool namely, that for `k` denoms in pool,
	// there should be k * (k - 1) / 2 records
	denomNum := len(denoms)
	expectedRecordsLength := denomNum * (denomNum - 1) / 2

	if expectedRecordsLength != len(records) {
		return types.InvalidRecordCountError{Expected: expectedRecordsLength, Actual: len(records)}
	}

	for _, record := range records {
		newRecord, err := k.updateRecord(ctx, record)
		if err != nil {
			return err
		}
		k.StoreNewRecord(ctx, newRecord)
	}
	return nil
}

// updateRecord returns a new record with updated accumulators and block time
// for the current block time.
func (k Keeper) updateRecord(ctx sdk.Context, record types.TwapRecord) (types.TwapRecord, error) {
	// Returns error for last updated records in the same block.
	// Exception: record is initialized when creating a pool,
	// then the TwapAccumulator variables are zero.

	// Handle record after creating pool
	// In case record height should equal to ctx height
	// But ArithmeticTwapAccumulators should be zero
	if (record.Height == ctx.BlockHeight() || record.Time.Equal(ctx.BlockTime())) &&
		!record.P1ArithmeticTwapAccumulator.IsZero() &&
		!record.P0ArithmeticTwapAccumulator.IsZero() {
		return types.TwapRecord{}, types.InvalidUpdateRecordError{}
	} else if record.Height > ctx.BlockHeight() || record.Time.After(ctx.BlockTime()) {
		// Normal case, ctx should be after record height & time
		return types.TwapRecord{}, types.InvalidUpdateRecordError{
			RecordBlockHeight: record.Height,
			RecordTime:        record.Time,
			ActualBlockHeight: ctx.BlockHeight(),
			ActualTime:        ctx.BlockTime(),
		}
	}

	newRecord := recordWithUpdatedAccumulators(record, ctx.BlockTime())
	newRecord.Height = ctx.BlockHeight()

	newSp0, newSp1, lastErrorTime := getSpotPrices(
		ctx, k.poolmanagerKeeper, record.PoolId, record.Asset0Denom, record.Asset1Denom, record.LastErrorTime)

	// set last spot price to be last price of this block. This is what will get used in interpolation.
	newRecord.P0LastSpotPrice = newSp0
	newRecord.P1LastSpotPrice = newSp1
	newRecord.LastErrorTime = lastErrorTime

	return newRecord, nil
}

// recordWithUpdatedAccumulators returns a record, with updated accumulator values and time for provided newTime,
// otherwise referred to as "interpolating the record" to the target time.
// This does not mutate the passed in record.
//
// pre-condition: newTime >= record.Time
func recordWithUpdatedAccumulators(record types.TwapRecord, newTime time.Time) types.TwapRecord {
	// return the given record: no need to calculate and update the accumulator if record time matches.
	if record.Time.Equal(newTime) {
		return record
	}
	newRecord := record
	timeDelta := types.CanonicalTimeMs(newTime) - types.CanonicalTimeMs(record.Time)
	newRecord.Time = newTime

	// record.LastSpotPrice is the last spot price from the block the record was created in,
	// thus it is treated as the effective spot price until the new time.
	// (As there was no change until at or after this time)
	p0NewAccum := types.SpotPriceMulDuration(record.P0LastSpotPrice, timeDelta)
	newRecord.P0ArithmeticTwapAccumulator = p0NewAccum.AddMut(newRecord.P0ArithmeticTwapAccumulator)

	p1NewAccum := types.SpotPriceMulDuration(record.P1LastSpotPrice, timeDelta)
	newRecord.P1ArithmeticTwapAccumulator = p1NewAccum.AddMut(newRecord.P1ArithmeticTwapAccumulator)

	// If the last spot price is zero, then the logarithm is undefined.
	// As a result, we cannot update the geometric accumulator.
	// We set the last error time to be the new time, and return the record.
	if record.P0LastSpotPrice.IsZero() {
		newRecord.LastErrorTime = newTime
		return newRecord
	}

	// NOTE: An edge case exists here. If a pool is drained of all it's liquidity, and then the pool's
	// spot price is set to exactly one and the GeometricTWAP is queried, the the result will be zero.
	// This is because the P0LastSpotPrice is one, which makes log_{2}{P_0} = 0, and thus the geometric
	// accumulator is the same as the time of pool drain.

	// This is a edge case almost certainly to never be hit in prod, but its good to be aware of.

	// logP0SpotPrice = log_{2}{P_0}
	logP0SpotPrice := twapLog(record.P0LastSpotPrice)
	// p0NewGeomAccum = log_{2}{P_0} * timeDelta
	p0NewGeomAccum := types.SpotPriceMulDuration(logP0SpotPrice, timeDelta)
	newRecord.GeometricTwapAccumulator = p0NewGeomAccum.AddMut(newRecord.GeometricTwapAccumulator)

	return newRecord
}

// getInterpolatedRecord returns a record for this pool, representing its accumulator state at time `t`.
// This is achieved by getting the record `r` that is at, or immediately preceding in state time `t`.
// To be clear: the record r s.t. `t - r.Time` is minimized AND `t >= r.Time`
// If for the record obtained, r.Time == r.LastErrorTime, this will also hold for the interpolated record.
func (k Keeper) getInterpolatedRecord(ctx sdk.Context, poolId uint64, t time.Time, assetA, assetB string) (types.TwapRecord, error) {
	record, err := k.getRecordAtOrBeforeTime(ctx, poolId, t, assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	// if it had errored on the last record, make this record inherit the error
	if record.Time.Equal(record.LastErrorTime) {
		record.LastErrorTime = t
	}
	record = recordWithUpdatedAccumulators(record, t)
	return record, nil
}

func (k Keeper) getMostRecentRecord(ctx sdk.Context, poolId uint64, assetA, assetB string) (types.TwapRecord, error) {
	record, err := k.getMostRecentRecordStoreRepresentation(ctx, poolId, assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record = recordWithUpdatedAccumulators(record, ctx.BlockTime())
	return record, nil
}

// computeTwap computes and returns a TWAP of a given
// type - arithmetic or geometric.
// Between two records given the quote asset.
// precondition: endRecord.Time >= startRecord.Time
// if (endRecord.LastErrorTime >= startRecord.Time) returns an error at end + result
// if (startRecord.LastErrorTime == startRecord.Time) returns an error at end + result
// if (endRecord.Time == startRecord.Time) returns endRecord.LastSpotPrice
// else returns
// (endRecord.Accumulator - startRecord.Accumulator) / (endRecord.Time - startRecord.Time)
func computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string, strategy twapStrategy) (osmomath.Dec, error) {
	// see if we need to return an error, due to spot price issues
	var err error = nil
	if endRecord.LastErrorTime.After(startRecord.Time) ||
		endRecord.LastErrorTime.Equal(startRecord.Time) ||
		startRecord.LastErrorTime.Equal(startRecord.Time) {
		err = errors.New("twap: error in pool spot price occurred between start and end time, twap result may be faulty")
	}
	timeDelta := endRecord.Time.Sub(startRecord.Time)
	// if time difference is 0, then return the last spot price based off of start.
	if timeDelta == time.Duration(0) {
		if quoteAsset == startRecord.Asset0Denom {
			return endRecord.P0LastSpotPrice, err
		}
		return endRecord.P1LastSpotPrice, err
	}

	return strategy.computeTwap(startRecord, endRecord, quoteAsset), err
}

// twapLog returns the logarithm of the given spot price, base 2.
// Panics if zero is given.
func twapLog(price osmomath.Dec) osmomath.Dec {
	if price.IsZero() {
		panic("twap: cannot take logarithm of zero")
	}

	return osmomath.BigDecFromDec(price).LogBase2().Dec()
}
