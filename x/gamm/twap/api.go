package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

// GetArithmeticTwap returns an arithmetic TWAP result from (startTime, endTime),
// for the `quoteAsset / baseAsset` ratio on `poolId`.
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
	endRecord, err := k.GetLatestAccumulatorRecord(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	twap := k.getArithmeticTwap(startRecord, endRecord)
	return twap, nil
}

// GetLatestAccumulatorRecord returns a TwapRecord struct that can be stored
func (k Keeper) GetLatestAccumulatorRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	// correct ordering of args for db
	if asset1Denom > asset0Denom {
		asset0Denom, asset1Denom = asset1Denom, asset0Denom
	}
	return k.getMostRecentTWAP(ctx, poolId, asset0Denom, asset1Denom)
}
