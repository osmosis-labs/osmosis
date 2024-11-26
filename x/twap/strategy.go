package twap

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"

	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

// twapStrategy is an interface for computing TWAPs.
// We have two strategies implementing the interface - arithmetic and geometric.
// We expose a common TWAP API to reduce duplication and avoid complexity.
type twapStrategy interface {
	// computeTwap calculates the TWAP with specific startRecord and endRecord.
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) osmomath.Dec
}

type arithmetic struct {
	TwapKeeper Keeper
}

type geometric struct {
	TwapKeeper Keeper
}

// computeTwap computes and returns an arithmetic TWAP between
// two records given the quote asset.
func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) osmomath.Dec {
	var accumDiff osmomath.Dec
	if quoteAsset == startRecord.Asset0Denom {
		accumDiff = endRecord.P0ArithmeticTwapAccumulator.Sub(startRecord.P0ArithmeticTwapAccumulator)
	} else {
		accumDiff = endRecord.P1ArithmeticTwapAccumulator.Sub(startRecord.P1ArithmeticTwapAccumulator)
	}
	timeDelta := types.CanonicalTimeMs(endRecord.Time) - types.CanonicalTimeMs(startRecord.Time)
	return types.AccumDiffDivDuration(accumDiff, timeDelta)
}

// computeTwap computes and returns a geometric TWAP between
// two records given the quote asset.
func (s *geometric) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) osmomath.Dec {
	accumDiff := endRecord.GeometricTwapAccumulator.Sub(startRecord.GeometricTwapAccumulator)

	if accumDiff.IsZero() {
		return osmomath.ZeroDec()
	}

	timeDelta := types.CanonicalTimeMs(endRecord.Time) - types.CanonicalTimeMs(startRecord.Time)
	arithmeticMeanOfLogPrices := types.AccumDiffDivDuration(accumDiff, timeDelta)

	exponent := arithmeticMeanOfLogPrices
	// result = 2^exponent = 2^arithmeticMeanOfLogPrices
	result := osmomath.Exp2(osmomath.BigDecFromDec(exponent.Abs()))

	isExponentNegative := exponent.IsNegative()
	isQuoteAsset0 := quoteAsset == startRecord.Asset0Denom

	// Case 1: exponent is negative and quoteAsset is asset 0.
	// This means that we need to invert the result to get the true value of the geometric mean.
	invertCase1 := isExponentNegative && isQuoteAsset0
	// Case 2: exponent is positive and quoteAsset is asset 1.
	// We need to use the following property: geometric mean of recprocals is reciprocal of geometric mean.
	// https://proofwiki.org/wiki/Geometric_Mean_of_Reciprocals_is_Reciprocal_of_Geometric_Mean
	invertCase2 := !isExponentNegative && !isQuoteAsset0
	if invertCase1 || invertCase2 {
		result = osmomath.OneBigDec().Quo(result)
	}

	// N.B. we round because this is the max number of significant figures supported
	// by the underlying spot price function.
	return osmomath.SigFigRound(result.Dec(), gammtypes.SpotPriceSigFigs)
}
