package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/x/twap/types"
)

// GetArithmeticTwap returns an arithmetic time weighted average price.
// The returned twap is the time weighted average price (TWAP) of:
// * the base asset, in units of the quote asset (1 unit of base = x units of quote)
// * from (startTime, endTime),
// * as determined by prices from AMM pool `poolId`.
//
// startTime and endTime do not have to be real block times that occurred,
// the state machine will interpolate the accumulator values for those times
// from the latest Twap accumulation record prior to the provided time.
//
// startTime must be within 48 hours of ctx.BlockTime(), if you need older TWAPs,
// you will have to maintain the accumulator yourself.
//
// endTime will be set in the function ArithmeticTwap() to ctx.BlockTime() which calls GetArithmeticTwap function if:
// * it is not provided externally
// * it is set to current time
//
// This function will error if:
// * startTime > endTime
// * endTime in the future
// * startTime older than 48 hours OR pool creation
// * pool with id poolId does not exist, or does not contain quoteAssetDenom, baseAssetDenom
// * there were some computational errors during computing arithmetic twap within the time range of
//   startRecord, endRecord - including the exact record times, which indicates that the result returned could be faulty

// N.B. If there is a notable use case, the state machine could maintain more historical records, e.g. at one per hour.
func (k Keeper) GetArithmeticTwap(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	endTime time.Time,
) (sdk.Dec, error) {
	return k.getTwap(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, endTime, k.GetArithmeticStrategy())
}

func (k Keeper) GetExtraArithmeticTwap(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	endTime time.Time,
) (sdk.Dec, error) {
	return k.getExtraTwap(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, endTime, k.GetArithmeticStrategy())
}

func (k Keeper) GetGeometricTwap(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	endTime time.Time,
) (sdk.Dec, error) {
	return k.getTwap(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, endTime, k.GetGeometricStrategy())
}

// GetArithmeticTwapToNow returns arithmetic twap from start time until the current block time for quote and base
// assets in a given pool.
func (k Keeper) GetArithmeticTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (sdk.Dec, error) {
	return k.getTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, k.GetArithmeticStrategy())
}

func (k Keeper) GetGeometricTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (sdk.Dec, error) {
	return k.getTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, k.GetGeometricStrategy())
}

// getTwap computes and returns twap from the start time until the end time. The type
// of twap returned depends on the strategy given and can be either arithmetic or geometric.
func (k Keeper) getTwap(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	endTime time.Time,
	strategy twapStrategy,
) (sdk.Dec, error) {
	if startTime.After(endTime) {
		return sdk.Dec{}, types.StartTimeAfterEndTimeError{StartTime: startTime, EndTime: endTime}
	}
	if endTime.Equal(ctx.BlockTime()) {
		return k.getTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, strategy)
	} else if endTime.After(ctx.BlockTime()) {
		return sdk.Dec{}, types.EndTimeInFutureError{EndTime: endTime, BlockTime: ctx.BlockTime()}
	}
	startRecord, err := k.getInterpolatedRecord(ctx, poolId, startTime, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	endRecord, err := k.getInterpolatedRecord(ctx, poolId, endTime, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return computeTwap(startRecord, endRecord, quoteAssetDenom, strategy)
}

type AccumTracker struct {
	P0Accumulator sdk.Dec
	P1Accumulator sdk.Dec
	Time          time.Time
}

func (k Keeper) getExtraTwap(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	endTime time.Time,
	strategy twapStrategy,
) (sdk.Dec, error) {
	if startTime.After(endTime) {
		return sdk.Dec{}, types.StartTimeAfterEndTimeError{StartTime: startTime, EndTime: endTime}
	}
	if endTime.After(ctx.BlockTime()) {
		return sdk.Dec{}, types.EndTimeInFutureError{EndTime: endTime, BlockTime: ctx.BlockTime()}
	}

	interval := k.GetParams(ctx).RecordHistoryKeepPeriod

	var trackers []AccumTracker

	lastPreviousRecordKeep, err := k.getInterpolatedRecord(ctx, poolId, endTime.Add(-interval), baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	// endTime still in keep period
	if !endTime.Before(ctx.BlockTime().Add(-interval)) {
		endRecord, err := k.getInterpolatedRecord(ctx, poolId, endTime, baseAssetDenom, quoteAssetDenom)
		if err != nil {
			return sdk.Dec{}, err
		}
		trackers = append(trackers, []AccumTracker{{P0Accumulator: lastPreviousRecordKeep.P0ArithmeticTwapAccumulator, P1Accumulator: lastPreviousRecordKeep.P1ArithmeticTwapAccumulator, Time: lastPreviousRecordKeep.Time}, {P0Accumulator: endRecord.P0ArithmeticTwapAccumulator, P1Accumulator: endRecord.P1ArithmeticTwapAccumulator, Time: endRecord.Time}}...)
	} else {
		// If endTime < ctx.Height - interval
		// We loop through each 2 days interval to find the accum of endTime
		tempEndRecord := lastPreviousRecordKeep
		for tempEndRecord.Time.Add(-interval).After(endTime) {
			tracker, err := k.getAccumTracker(ctx, tempEndRecord.Time.Add(-interval), tempEndRecord, interval.Milliseconds())
			if err != nil {
				return sdk.Dec{}, err
			}
			tempEndRecord.Time = tracker.Time
			tempEndRecord.P0ArithmeticTwapAccumulator = tracker.P0Accumulator
			tempEndRecord.P1ArithmeticTwapAccumulator = tracker.P1Accumulator
		}

		diffTime := types.CanonicalTimeMs(tempEndRecord.Time) - types.CanonicalTimeMs(endTime)
		if diffTime == 0 {
			trackers = append(trackers, AccumTracker{P0Accumulator: tempEndRecord.P0ArithmeticTwapAccumulator, P1Accumulator: tempEndRecord.P1ArithmeticTwapAccumulator, Time: tempEndRecord.Time})
		} else {
			// With different time < interval, calculate accumulator directly from tempEndRecord
			tracker, err := k.getAccumTracker(ctx, endTime, tempEndRecord, diffTime)
			if err != nil {
				return sdk.Dec{}, err
			}
			trackers = append(trackers, tracker)
		}
	}

	// After have tracker, we find accum of startTime
	for !trackers[0].Time.Add(-interval).Before(startTime) {
		tempRecord := types.TwapRecord{
			PoolId: lastPreviousRecordKeep.PoolId,
			Asset0Denom: lastPreviousRecordKeep.Asset0Denom,
			Asset1Denom: lastPreviousRecordKeep.Asset1Denom,
			P0ArithmeticTwapAccumulator: trackers[0].P0Accumulator,
			P1ArithmeticTwapAccumulator: trackers[0].P1Accumulator,
		}
		tracker, err := k.getAccumTracker(ctx, trackers[0].Time.Add(-interval), tempRecord, interval.Milliseconds())
		if err != nil {
			return sdk.Dec{}, err
		}
		trackers = append([]AccumTracker{tracker}, trackers...)
	}

	if !startTime.Equal(trackers[0].Time) {
		tempRecord := types.TwapRecord{
			PoolId: lastPreviousRecordKeep.PoolId,
			Asset0Denom: lastPreviousRecordKeep.Asset0Denom,
			Asset1Denom: lastPreviousRecordKeep.Asset1Denom,
			P0ArithmeticTwapAccumulator: trackers[0].P0Accumulator,
			P1ArithmeticTwapAccumulator: trackers[0].P1Accumulator,
		}
		tracker, err := k.getAccumTracker(ctx, startTime, tempRecord, trackers[0].Time.Sub(startTime).Milliseconds())
		if err != nil {
			return sdk.Dec{}, err
		}
		trackers = append([]AccumTracker{tracker}, trackers...)
	}

	timeDelta := types.CanonicalTimeMs(endTime) - types.CanonicalTimeMs(startTime)
	var accumDiff sdk.Dec
	if quoteAssetDenom == lastPreviousRecordKeep.Asset0Denom {
		accumDiff = trackers[len(trackers) - 1].P0Accumulator.Sub(trackers[0].P0Accumulator)
	} else {
		accumDiff = trackers[len(trackers) - 1].P1Accumulator.Sub(trackers[0].P1Accumulator)
	}
	return types.AccumDiffDivDuration(accumDiff, timeDelta), nil
}

func (k Keeper) getAccumTracker(
	ctx sdk.Context, 
	time time.Time,
	existingRecord types.TwapRecord,
	diffTime int64,
) (AccumTracker, error) {
	tempCtx := ctx.WithBlockTime(time)
	sp0, err := k.poolmanagerKeeper.RouteCalculateSpotPrice(tempCtx, existingRecord.PoolId, existingRecord.Asset0Denom, existingRecord.Asset1Denom)
	if err != nil {
		return AccumTracker{}, err
	}
	sp1, err := k.poolmanagerKeeper.RouteCalculateSpotPrice(tempCtx, existingRecord.PoolId, existingRecord.Asset1Denom, existingRecord.Asset0Denom)
	if err != nil {
		return AccumTracker{}, err
	}

	accum0 := existingRecord.P0ArithmeticTwapAccumulator.Sub(sp0.MulInt64(diffTime))
	accum1 := existingRecord.P1ArithmeticTwapAccumulator.Sub(sp1.MulInt64(diffTime))
	
	return AccumTracker{P0Accumulator: accum0, P1Accumulator: accum1, Time: time}, nil
}

// getTwapToNow computes and returns twap from the start time until the current block time. The type
// of twap returned depends on the strategy given and can be either arithmetic or geometric.
func (k Keeper) getTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	strategy twapStrategy,
) (sdk.Dec, error) {
	if startTime.After(ctx.BlockTime()) {
		return sdk.Dec{}, types.StartTimeAfterEndTimeError{StartTime: startTime, EndTime: ctx.BlockTime()}
	}

	startRecord, err := k.getInterpolatedRecord(ctx, poolId, startTime, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	endRecord, err := k.GetBeginBlockAccumulatorRecord(ctx, poolId, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return computeTwap(startRecord, endRecord, quoteAssetDenom, strategy)
}

// GetBeginBlockAccumulatorRecord returns a TwapRecord struct corresponding to the state of pool `poolId`
// as of the beginning of the block this is called on.
func (k Keeper) GetBeginBlockAccumulatorRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getMostRecentRecord(ctx, poolId, asset0Denom, asset1Denom)
}
