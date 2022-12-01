package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

// twapStrategy is an interface for computing TWAPs.
// We have two strategies implementing the interface - arithmetic and geometric.
// We expose a common TWAP API to reduce duplication and avoid complexity.
type twapStrategy interface {
	// computeTwap calculates the TWAP with specific startRecord and endRecord.
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error)
}

type arithmetic struct {
	keeper Keeper
}

var _ twapStrategy = &arithmetic{}

func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	return computeTwap(startRecord, endRecord, quoteAsset, arithmeticTwapType)
}
