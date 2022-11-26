package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

// Since we have two methods of computing TWAP; airthmetic and geometric.
// We expose a common TWAP API to reduce duplication and avoid complexity.
type twapStrategy interface {
	getTwapToNow(
		ctx sdk.Context,
		poolId uint64,
		baseAssetDenom string,
		quoteAssetDenom string,
		startTime time.Time) (sdk.Dec, error)
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error)
}

type arithmetic struct {
	keeper Keeper
}

var _ twapStrategy = &arithmetic{}

// getTwapToNow calculates the TWAP with endRecord as currentBlocktime.
func (s *arithmetic) getTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (sdk.Dec, error) {
	// decide whether we want to use arithmetic or geometric twap.
	return s.keeper.GetArithmeticTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime)
}

// getTwapToNow calculates the TWAP with specific startRecord and endRecord.
func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	// decide whether we want to use arithmetic or geometric twap
	return computeArithmeticTwap(startRecord, endRecord, quoteAsset)
}
