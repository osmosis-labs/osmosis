package twap

import (
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
// The
//
// startTime and endTime do not have to be real block times that occurred,
// this function will interpolate between startTime.
// if endTime = now, we do {X}
// startTime must be in time range {X}, recommended parameterization for mainnet is {Y}
func (k Keeper) GetArithmeticTwap(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time,
	endTime time.Time) (sdk.Dec, error) {
	if endTime.Equal(ctx.BlockTime()) {
		return k.GetArithmeticTwapToNow(ctx, poolId, quoteAssetDenom, baseAssetDenom, startTime)
	}
	// startTwapRecord, err := k.getTwapBeforeTime(ctx, poolId, startTime, quoteAssetDenom, baseAssetDenom)
	return sdk.Dec{}, nil
}

func (k Keeper) GetArithmeticTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time) (sdk.Dec, error) {
	startRecord, err := k.getStartRecord(ctx, poolId, startTime, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	endRecord, err := k.GetBeginBlockAccumulatorRecord(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	twap := k.getArithmeticTwap(startRecord, endRecord, quoteAssetDenom)
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
