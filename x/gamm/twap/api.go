package twap

import (
	"errors"
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
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
	startTime time.Time,
	endTime time.Time) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}

// GetLatestAccumulatorRecord returns a TwapRecord struct that can be stored
func (k twapkeeper) GetLatestAccumulatorRecord(poolId uint64, asset0Denom string, asset1Denom string) (types.TwapRecord, error) {
	return types.TwapRecord{}, errors.New("unimplemented")
}
