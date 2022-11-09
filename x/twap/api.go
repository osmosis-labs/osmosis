package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
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
	if startTime.After(endTime) {
		return sdk.Dec{}, types.StartTimeAfterEndTimeError{StartTime: startTime, EndTime: endTime}
	}
	if endTime.Equal(ctx.BlockTime()) {
		return k.GetArithmeticTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime)
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
	return computeArithmeticTwap(startRecord, endRecord, quoteAssetDenom)
}

// GetArithmeticTwapToNow returns GetArithmeticTwap on the input, with endTime being fixed to ctx.BlockTime()
// This function does not mutate records.
func (k Keeper) GetArithmeticTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
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
	return computeArithmeticTwap(startRecord, endRecord, quoteAssetDenom)
}

// GetBeginBlockAccumulatorRecord returns a TwapRecord struct corresponding to the state of pool `poolId`
// as of the beginning of the block this is called on.
// This uses the state of the beginning of the block, as if there were swaps since the block has started,
// these swaps have had no time to be arbitraged back.
// This accumulator can be stored, to compute wider ranged twaps.
func (k Keeper) GetBeginBlockAccumulatorRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getMostRecentRecord(ctx, poolId, asset0Denom, asset1Denom)
}
