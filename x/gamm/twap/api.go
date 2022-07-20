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
func (k twapkeeper) GetArithmeticTwap(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time,
	endTime time.Time) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}

func (k twapkeeper) GetArithmeticTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time) (sdk.Dec, error) {
	return k.GetArithmeticTwap(ctx, poolId, quoteAssetDenom, baseAssetDenom, startTime, ctx.BlockTime())
}

// GetLatestAccumulatorRecord returns a TwapRecord struct that can be stored
func (k twapkeeper) GetLatestAccumulatorRecord(ctx sdk.Context, poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	// correct ordering of args for db
	if asset1Denom > asset0Denom {
		asset0Denom, asset1Denom = asset1Denom, asset0Denom
	}
	return k.getMostRecentTWAP(ctx, poolId, asset0Denom, asset1Denom)
}
