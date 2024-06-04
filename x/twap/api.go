package twap

import (
	"time"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/x/twap/types"
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
) (osmomath.Dec, error) {
	return k.getTwap(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, endTime, k.GetArithmeticStrategy())
}

func (k Keeper) GetGeometricTwap(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
	endTime time.Time,
) (osmomath.Dec, error) {
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
) (osmomath.Dec, error) {
	return k.getTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, k.GetArithmeticStrategy())
}

func (k Keeper) GetGeometricTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (osmomath.Dec, error) {
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
) (osmomath.Dec, error) {
	if startTime.After(endTime) {
		return osmomath.Dec{}, types.StartTimeAfterEndTimeError{StartTime: startTime, EndTime: endTime}
	}
	if endTime.Equal(ctx.BlockTime()) {
		return k.getTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime, strategy)
	} else if endTime.After(ctx.BlockTime()) {
		return osmomath.Dec{}, types.EndTimeInFutureError{EndTime: endTime, BlockTime: ctx.BlockTime()}
	}
	startRecord, err := k.getInterpolatedRecord(ctx, poolId, startTime, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return osmomath.Dec{}, err
	}
	endRecord, err := k.getInterpolatedRecord(ctx, poolId, endTime, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return osmomath.Dec{}, err
	}

	return computeTwap(startRecord, endRecord, quoteAssetDenom, strategy)
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
) (osmomath.Dec, error) {
	if startTime.After(ctx.BlockTime()) {
		return osmomath.Dec{}, types.StartTimeAfterEndTimeError{StartTime: startTime, EndTime: ctx.BlockTime()}
	}

	startRecord, err := k.getInterpolatedRecord(ctx, poolId, startTime, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return osmomath.Dec{}, err
	}
	endRecord, err := k.GetBeginBlockAccumulatorRecord(ctx, poolId, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		return osmomath.Dec{}, err
	}

	return computeTwap(startRecord, endRecord, quoteAssetDenom, strategy)
}

// GetBeginBlockAccumulatorRecord returns a TwapRecord struct corresponding to the state of pool `poolId`
// as of the beginning of the block this is called on.
func (k Keeper) GetBeginBlockAccumulatorRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return k.getMostRecentRecord(ctx, poolId, asset0Denom, asset1Denom)
}

// UnsafeGetMultiPoolArithmeticTwapToNow returns the TWAP price of two assets through multiple pools.
// The price is calculated by taking the arithmetic TWAP of the two assets in each pool and multiplying
// them together.
//
// Only pools with two assets are considered.
// For each pool n, its base asset will be the quote asset of pool n-1. The first pool's base asset is specified
// in baseAssetDenom and the last pool's quote asset is specified in quoteAssetDenom.
//
// N.B. This function is considered "unsafe" because it calculates the TWAP across multiple pools by multiplying
// the TWAPs of individual pools, which is not technically correct. This is akin to calculating `average(a) * average(b)`
// instead of `average(a * b)`. In general, `average(a * b)` is not necessarily equal to `average(a) * average(b)`.
// This method can safely be used in instances where both of the following are true:
// 1. The TWAP of all but one route have a stable pair. For instance, using on allBTC -> WBTC -> OSMO works because we
// expect the TWAP of allBTC -> WBTC to be fairly stable.
// 2. All assets involved are major assets that we can expected proto-rev cyclic arbs will handle arb opportunities. Therefore,
// the result must be acceptatble within 1-3% error margin (roughly ~.3% * number of pools in path to close the arb).
func (k Keeper) UnsafeGetMultiPoolArithmeticTwapToNow(
	ctx sdk.Context,
	route []*poolmanagertypes.SwapAmountInRoute,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (osmomath.Dec, error) {
	if len(route) == 0 {
		return osmomath.Dec{}, types.ErrEmptyRoute
	}
	if route[len(route)-1].TokenOutDenom != quoteAssetDenom {
		return osmomath.Dec{}, types.ErrMismatchedQuoteAsset
	}

	price := osmomath.NewDecFromInt(osmomath.OneInt())
	baseAsset := baseAssetDenom

	for _, pool := range route {
		quoteAsset := pool.TokenOutDenom
		twap, err := k.GetArithmeticTwapToNow(ctx, pool.PoolId, baseAsset, quoteAsset, startTime)
		if err != nil {
			return osmomath.Dec{}, err
		}
		price = price.Mul(twap)
		// Update the base asset to the current quote asset for the next iteration
		baseAsset = quoteAsset
	}

	return price, nil
}
