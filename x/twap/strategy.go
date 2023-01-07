package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

// twapStrategy is an interface for computing TWAPs.
// We have two strategies implementing the interface - arithmetic and geometric.
// We expose a common TWAP API to reduce duplication and avoid complexity.
type twapStrategy interface {
	// computeTwap calculates the TWAP with specific startRecord and endRecord.
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) sdk.Dec
}

type arithmetic struct {
	TwapKeeper Keeper
}

type geometric struct {
	TwapKeeper Keeper
}

// computeTwap computes and returns an arithmetic TWAP between
// two records given the quote asset.
func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) sdk.Dec {
	var accumDiff sdk.Dec
	if quoteAsset == startRecord.Asset0Denom {
		accumDiff = endRecord.P0ArithmeticTwapAccumulator.Sub(startRecord.P0ArithmeticTwapAccumulator)
	} else {
		accumDiff = endRecord.P1ArithmeticTwapAccumulator.Sub(startRecord.P1ArithmeticTwapAccumulator)
	}
	timeDelta := endRecord.Time.Sub(startRecord.Time)
	return types.AccumDiffDivDuration(accumDiff, timeDelta)
}

// computeTwap computes and returns a geometric TWAP between
// two records given the quote asset.
func (s *geometric) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) sdk.Dec {
	accumDiff := endRecord.GeometricTwapAccumulator.Sub(startRecord.GeometricTwapAccumulator)

	if accumDiff.IsZero() {
		return sdk.ZeroDec()
	}

	timeDelta := endRecord.Time.Sub(startRecord.Time)
	arithmeticMeanOfLogPrices := types.AccumDiffDivDuration(accumDiff, timeDelta)

	result := twapPow(arithmeticMeanOfLogPrices)
	// N.B.: Geometric mean of recprocals is reciprocal of geometric mean.
	// https://proofwiki.org/wiki/Geometric_Mean_of_Reciprocals_is_Reciprocal_of_Geometric_Mean
	if quoteAsset == startRecord.Asset1Denom {
		result = sdk.OneDec().Quo(result)
	}

	// N.B. we round because this is the max number of significant figures supported
	// by the underlying spot price function.
	return osmomath.SigFigRound(result, gammtypes.SpotPriceSigFigs)
}
