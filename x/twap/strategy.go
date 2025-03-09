package twap

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v29/x/twap/types"

	gammtypes "github.com/osmosis-labs/osmosis/v29/x/gamm/types"
)

// twapStrategy is an interface for computing TWAPs.
// We have two strategies implementing the interface - arithmetic and geometric.
// We expose a common TWAP API to reduce duplication and avoid complexity.
type twapStrategy interface {
	// computeTwap calculates the TWAP with specific startRecord and endRecord.
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) osmomath.Dec
	
	// updateAccumulators updates the accumulators in a record based on the strategy.
	updateAccumulators(record types.TwapRecord, newTime time.Time) types.TwapRecord
}

type arithmetic struct {
	TwapKeeper Keeper
}

func (s *arithmetic) updateAccumulators(record types.TwapRecord, newTime time.Time) types.TwapRecord {
	// return the given record: no need to calculate and update the accumulator if record time matches.
	if record.Time.Equal(newTime) {
		return record
	}
	newRecord := record
	timeDelta := types.CanonicalTimeMs(newTime) - types.CanonicalTimeMs(record.Time)
	newRecord.Time = newTime

	// record.LastSpotPrice is the last spot price from the block the record was created in,
	// thus it is treated as the effective spot price until the new time.
	// (As there was no change until at or after this time)
	p0NewAccum := types.SpotPriceMulDuration(record.P0LastSpotPrice, timeDelta)
	newRecord.P0ArithmeticTwapAccumulator = p0NewAccum.AddMut(newRecord.P0ArithmeticTwapAccumulator)

	p1NewAccum := types.SpotPriceMulDuration(record.P1LastSpotPrice, timeDelta)
	newRecord.P1ArithmeticTwapAccumulator = p1NewAccum.AddMut(newRecord.P1ArithmeticTwapAccumulator)

	return newRecord
}

type geometric struct {
	TwapKeeper Keeper
}

func (s *geometric) updateAccumulators(record types.TwapRecord, newTime time.Time) types.TwapRecord {
	// return the given record: no need to calculate and update the accumulator if record time matches.
	if record.Time.Equal(newTime) {
		return record
	}
	newRecord := record
	timeDelta := types.CanonicalTimeMs(newTime) - types.CanonicalTimeMs(record.Time)
	newRecord.Time = newTime

	// Geometric strategy still needs to update arithmetic accumulators for backward compatibility
	p0NewAccum := types.SpotPriceMulDuration(record.P0LastSpotPrice, timeDelta)
	newRecord.P0ArithmeticTwapAccumulator = p0NewAccum.AddMut(newRecord.P0ArithmeticTwapAccumulator)

	p1NewAccum := types.SpotPriceMulDuration(record.P1LastSpotPrice, timeDelta)
	newRecord.P1ArithmeticTwapAccumulator = p1NewAccum.AddMut(newRecord.P1ArithmeticTwapAccumulator)

	// If the last spot price is zero, then the logarithm is undefined.
	// As a result, we cannot update the geometric accumulator.
	// We set the last error time to be the new time, and return the record.
	if record.P0LastSpotPrice.IsZero() {
		newRecord.LastErrorTime = newTime
		return newRecord
	}

	// logP0SpotPrice = log_{2}{P_0}
	logP0SpotPrice := twapLog(record.P0LastSpotPrice)
	// p0NewGeomAccum = log_{2}{P_0} * timeDelta
	p0NewGeomAccum := types.SpotPriceMulDuration(logP0SpotPrice, timeDelta)
	newRecord.GeometricTwapAccumulator = p0NewGeomAccum.AddMut(newRecord.GeometricTwapAccumulator)

	return newRecord
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
