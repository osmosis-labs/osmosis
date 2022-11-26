package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

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

func (s *arithmetic) getTwapToNow(
	ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
	startTime time.Time,
) (sdk.Dec, error) {
	// decide whether we want to use arithmetic or geometric twap
	return s.keeper.GetArithmeticTwapToNow(ctx, poolId, baseAssetDenom, quoteAssetDenom, startTime)
}

func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	// decide whether we want to use arithmetic or geometric twap
	return computeArithmeticTwap(startRecord, endRecord, quoteAsset)
}
