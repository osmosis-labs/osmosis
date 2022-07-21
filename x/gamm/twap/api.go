package twap

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
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
// This function will error if:
// * startTime > endTime
// * endTime in the future
// * startTime older than 48 hours OR pool creation
// * pool with id poolId does not exist, or does not contain quoteAssetDenom, baseAssetDenom
//
// N.B. If there is a notable use case, the state machine could maintain more historical records, e.g. at one per hour.
func (k Keeper) GetArithmeticTwap(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time,
	endTime time.Time) (sdk.Dec, error) {
	if endTime.Equal(ctx.BlockTime()) {
		return k.GetArithmeticTwapToNow(ctx, poolId, quoteAssetDenom, baseAssetDenom, startTime)
	} else if endTime.After(ctx.BlockTime()) {
		return sdk.Dec{}, errors.New("called GetArithmeticTwap with an end time in the future")
	} else if startTime.After(endTime) {
		return sdk.Dec{}, errors.New("called GetArithmeticTwap with a start time that is after the end time")
	}
	startRecord, err := k.getInterpolatedRecord(ctx, poolId, startTime, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	endRecord, err := k.getInterpolatedRecord(ctx, poolId, endTime, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	twap := computeArithmeticTwap(startRecord, endRecord, quoteAssetDenom)
	return twap, nil
}

func (k Keeper) GetArithmeticTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time) (sdk.Dec, error) {
	startRecord, err := k.getInterpolatedRecord(ctx, poolId, startTime, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	endRecord, err := k.GetBeginBlockAccumulatorRecord(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	twap := computeArithmeticTwap(startRecord, endRecord, quoteAssetDenom)
	return twap, nil
}

// GetCurrentAccumulatorRecord returns a TwapRecord struct corresponding to the state of pool `poolId`
// as of the beginning of the block this is called on.
// This uses the state of the beginning of the block, as if there were swaps since the block has started,
// these swaps have had no time to be arbitraged back.
// This accumulator can be stored, to compute wider ranged twaps.
func (k Keeper) GetBeginBlockAccumulatorRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	// correct ordering of args for db
	if asset1Denom > asset0Denom {
		asset0Denom, asset1Denom = asset1Denom, asset0Denom
	}
	return k.getMostRecentRecord(ctx, poolId, asset0Denom, asset1Denom)
}
