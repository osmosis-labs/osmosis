package twap

import (
	"errors"
	"fmt"

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
	computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error)
}

type arithmetic struct {
	TwapKeeper Keeper
}

type geometric struct {
	TwapKeeper Keeper
}

// computeTwap computes and returns an arithmetic TWAP between
// two records given the quote asset.
// TODO: test that error is always nil
func (s *arithmetic) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	var accumDiff sdk.Dec
	if quoteAsset == startRecord.Asset0Denom {
		accumDiff = endRecord.P0ArithmeticTwapAccumulator.Sub(startRecord.P0ArithmeticTwapAccumulator)
	} else {
		accumDiff = endRecord.P1ArithmeticTwapAccumulator.Sub(startRecord.P1ArithmeticTwapAccumulator)
	}
	timeDelta := endRecord.Time.Sub(startRecord.Time)
	return types.AccumDiffDivDuration(accumDiff, timeDelta), nil
}

// computeTwap computes and returns a geometric TWAP between
// two records given the quote asset.
// TODO: test all edge cases.
func (s *geometric) computeTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	accumDiff := endRecord.GeometricTwapAccumulator.Sub(startRecord.GeometricTwapAccumulator)

	timeDelta := endRecord.Time.Sub(startRecord.Time)
	arithmeticMeanOfLogPrices := types.AccumDiffDivDuration(accumDiff, timeDelta)

	if arithmeticMeanOfLogPrices.IsZero() {
		// This may happen only if every returned spot price in history had a spot price error,
		// resulting in a spot price of zero returned.
		return sdk.Dec{}, errors.New("internal geometric twap error: arithmetic mean of log prices is zero")
	}

	exponent := arithmeticMeanOfLogPrices
	// result = 2^exponent = 2^arithmeticMeanOfLogPrices
	result := osmomath.Exp2(osmomath.BigDecFromSDKDec(exponent.Abs()))

	// Case 1: exponent is negative and quoteAsset is asset 0.
	// This means that we need to invert the result to get the true value of the geometric mean.
	invertCase1 := exponent.IsNegative() && quoteAsset == startRecord.Asset0Denom
	// Case 2: exponent is positive and quoteAsset is asset 1.
	// We need to use the following property: geometric mean of recprocals is reciprocal of geometric mean.
	// https://proofwiki.org/wiki/Geometric_Mean_of_Reciprocals_is_Reciprocal_of_Geometric_Mean
	invertCase2 := !exponent.IsNegative() && quoteAsset == startRecord.Asset1Denom
	if invertCase1 || invertCase2 {
		if result.IsZero() {
			return sdk.ZeroDec(), errors.New("internal geometric twap error: denominator is zero")
		}

		result = osmomath.OneDec().Quo(result)
	}

	if result.IsZero() {
		return sdk.Dec{}, errors.New("internal geometric twap error: final twap is zero")
	}

	if result.LT(osmomath.BigDecFromSDKDec(gammtypes.MinSpotPrice)) {
		return sdk.Dec{}, fmt.Errorf("final twap is less than minimum spot price of 10^-18, was (%s). Note, result is displayed as zero if under 10^-36", result.String())
	}

	// N.B. we round because this is the max number of significant figures supported
	// by the underlying spot price function.
	return osmomath.SigFigRound(result.SDKDec(), gammtypes.SpotPriceSigFigs), nil
}
