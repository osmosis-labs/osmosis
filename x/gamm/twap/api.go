package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetArithmeticTwap returns an arithmetic TWAP result from (startTime, endTime),
// for the `quoteAsset / baseAsset` ratio on `poolId`.
// startTime and endTime do not have to be real block times that occured,
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
