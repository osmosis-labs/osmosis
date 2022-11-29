package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

// twapStrategy is an interface for computing TWAPs.
// We have two strategies implementing the interface - arithmetic and geometric.
// We expose a common TWAP API to reduce duplication and avoid complexity.
type twapStrategy interface {
	// getTwapToNow calculates the TWAP with endRecord as currentBlocktime.
	getTwapToNow(
		ctx sdk.Context,
		poolId uint64,
		baseAssetDenom string,
		quoteAssetDenom string,
		startTime time.Time) (sdk.Dec, error)
	// computeTwap calculates the TWAP with specific startRecord and endRecord.
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) sdk.Dec
}

type arithmetic struct {
	keeper Keeper
}

var _ twapStrategy = &arithmetic{}

func (s *arithmetic) getTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (sdk.Dec, error) {
	return s.keeper.GetArithmeticTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime)
}

func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) sdk.Dec {
	return computeArithmeticTwap(startRecord, endRecord, quoteAsset)
}
