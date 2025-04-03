package math

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// TicksToSqrtPrice returns the sqrtPrice for the lower and upper ticks by
// individually calling `TickToSqrtPrice` method.
// Returns error if fails to calculate price.
func TicksToSqrtPrice(lowerTick, upperTick int64) (osmomath.BigDec, osmomath.BigDec, error) {
	if lowerTick >= upperTick {
		return osmomath.BigDec{}, osmomath.BigDec{}, types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	sqrtPriceUpperTick, err := TickToSqrtPrice(upperTick)
	if err != nil {
		return osmomath.BigDec{}, osmomath.BigDec{}, err
	}
	sqrtPriceLowerTick, err := TickToSqrtPrice(lowerTick)
	if err != nil {
		return osmomath.BigDec{}, osmomath.BigDec{}, err
	}
	return sqrtPriceLowerTick, sqrtPriceUpperTick, nil
}

// TickToSqrtPrice returns the sqrtPrice given a tickIndex
// If tickIndex is zero, the function returns osmomath.OneDec().
// It is the combination of calling TickToPrice followed by Sqrt.
func TickToSqrtPrice(tickIndex int64) (osmomath.BigDec, error) {
	priceBigDec, err := TickToPrice(tickIndex)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	// N.B. at launch, we only supported price range
	// of [tick(10^-12), tick(MaxSpotPrice)].
	// To maintain backwards state-compatibility, we use the original
	// math based on 18 precision decimal on the at the launch tick range.
	if tickIndex >= types.MinInitializedTick {
		// It is acceptable to truncate here as TickToPrice() function converts
		// from osmomath.Dec to osmomath.BigDec before returning specifically for this range.
		// As a result, there is no data loss.
		price := priceBigDec.Dec()

		sqrtPrice, err := osmomath.MonotonicSqrtMut(price)
		if err != nil {
			return osmomath.BigDec{}, err
		}
		return osmomath.BigDecFromDecMut(sqrtPrice), nil
	}

	// For the newly extended range of [tick(MinSpotPriceV2), MinInitializedTick), we use the new math
	// based on 36 precision decimal.
	sqrtPrice, err := osmomath.MonotonicSqrtBigDec(priceBigDec)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	return sqrtPrice, nil
}

// TickToPrice returns the price given a tickIndex
// If tickIndex is zero, the function returns osmomath.OneDec().
func TickToPrice(tickIndex int64) (osmomath.BigDec, error) {
	if tickIndex == 0 {
		return osmomath.OneBigDec(), nil
	}

	// N.B. We special case MinInitializedTickV2 and MinCurrentTickV2 since MinInitializedTickV2
	// is the first one that requires taking 10 to the exponent of (-31 + -6) = -37
	// Given BigDec's precision of 36, that cannot be supported.
	// The fact that MinInitializedTickV2 and MinCurrentTickV2 translate to the same
	// price is acceptable since MinCurrentTickV2 cannot be initialized.
	if tickIndex == types.MinInitializedTickV2 || tickIndex == types.MinCurrentTickV2 {
		return types.MinSpotPriceV2, nil
	}

	numAdditiveTicks, geometricExponentDelta, err := TickToAdditiveGeometricIndices(tickIndex)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	// price = 10^geometricExponentDelta + numAdditiveTicks * 10^exponentAtCurrentTick
	// exponent at current tick = types.ExponentAtPriceOne + geometricExponentDelta + conditional
	// where conditional = -1 if tickIndex < 0, 0 otherwise
	// so we compute the price as (10**(geometricExponentDelta - exponentAtCurrentTick) + numAdditiveTicks) * 10**exponentAtCurrentTick
	// notice that geometricExponentDelta - exponentAtCurrentTick is either 6 or 7
	// so we compute this as unscaledPrice = (10**(geometricExponentDelta - exponentAtCurrentTick) + numAdditiveTicks)

	// Calculate the exponentAtCurrentTick from the starting exponentAtPriceOne and the geometricExponentDelta
	exponentAtCurrentTick := types.ExponentAtPriceOne + geometricExponentDelta
	var unscaledPrice int64 = 1_000_000
	if tickIndex < 0 {
		// We must decrement the exponentAtCurrentTick when entering the negative tick range in order to constantly step up in precision when going further down in ticks
		// Otherwise, from tick 0 to tick -(geometricExponentIncrementDistanceInTicks), we would use the same exponent as the exponentAtPriceOne
		exponentAtCurrentTick = exponentAtCurrentTick - 1
		unscaledPrice *= 10
	}
	unscaledPrice += numAdditiveTicks
	price := powTenBigDec(exponentAtCurrentTick).MulInt64(unscaledPrice)

	// defense in depth, this logic would not be reached due to use having checked if given tick is in between
	// min tick and max tick.
	if price.GT(types.MaxSpotPriceBigDec) || price.LT(types.MinSpotPriceV2) {
		return osmomath.BigDec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPriceV2, MaxSpotPrice: types.MaxSpotPrice}
	}
	return price, nil
}

func TickToAdditiveGeometricIndices(tickIndex int64) (additiveTicks int64, geometricExponentDelta int64, err error) {
	if tickIndex == 0 {
		return 0, 0, nil
	}

	// N.B. We special case MinInitializedTickV2 and MinCurrentTickV2 since MinInitializedTickV2
	// is the first one that requires taking 10 to the exponent of (-31 + -6) = -37
	// Given BigDec's precision of 36, that cannot be supported.
	// The fact that MinInitializedTickV2 and MinCurrentTickV2 translate to the same
	// price is acceptable since MinCurrentTickV2 cannot be initialized.
	if tickIndex == types.MinInitializedTickV2 || tickIndex == types.MinCurrentTickV2 {
		return 0, -30, nil
	}

	// Check that the tick index is between min and max value
	if tickIndex < types.MinCurrentTickV2 {
		return 0, 0, types.TickIndexMinimumError{MinTick: types.MinCurrentTickV2}
	}
	if tickIndex > types.MaxTick {
		return 0, 0, types.TickIndexMaximumError{MaxTick: types.MaxTick}
	}

	// Use floor division to determine what the geometricExponent is now (the delta from the starting exponentAtPriceOne)
	geometricExponentDelta = tickIndex / geometricExponentIncrementDistanceInTicks

	// Now, starting at the minimum tick of the current increment, we calculate how many ticks in the current geometricExponent we have passed
	numAdditiveTicks := tickIndex - (geometricExponentDelta * geometricExponentIncrementDistanceInTicks)
	return numAdditiveTicks, geometricExponentDelta, nil
}

// RoundDownTickToSpacing rounds the tick index down to the nearest tick spacing if the tickIndex is in between authorized tick values
// Note that this is Euclidean modulus.
// The difference from default Go modulus is that Go default results
// in a negative remainder when the dividend is negative.
// Consider example tickIndex = -17, tickSpacing = 10
// tickIndexModulus = tickIndex % tickSpacing = -7
// tickIndexModulus = -7 + 10 = 3
// tickIndex = -17 - 3 = -20
func RoundDownTickToSpacing(tickIndex int64, tickSpacing int64) (int64, error) {
	tickIndexModulus := tickIndex % tickSpacing
	if tickIndexModulus < 0 {
		tickIndexModulus += tickSpacing
	}

	if tickIndexModulus != 0 {
		tickIndex = tickIndex - tickIndexModulus
	}

	// Defense-in-depth check to ensure that the tick index is within the authorized range
	// Should never get here.
	if tickIndex > types.MaxTick || tickIndex < types.MinInitializedTickV2 {
		return 0, types.TickIndexNotWithinBoundariesError{ActualTick: tickIndex, MinTick: types.MinInitializedTickV2, MaxTick: types.MaxTick}
	}

	return tickIndex, nil
}

// SqrtPriceToTickRoundDown converts the given sqrt price to its corresponding tick rounded down
// to the nearest tick spacing.
func SqrtPriceToTickRoundDownSpacing(sqrtPrice osmomath.BigDec, tickSpacing uint64) (int64, error) {
	tickIndex, err := CalculateSqrtPriceToTick(sqrtPrice)
	if err != nil {
		return 0, err
	}

	tickIndex, err = RoundDownTickToSpacing(tickIndex, int64(tickSpacing))
	if err != nil {
		return 0, err
	}

	return tickIndex, nil
}

// powTen treats negative exponents as 1/(10**|exponent|) instead of 10**-exponent
// This is because the osmomath.Dec.Power function does not support negative exponents
func PowTenInternal(exponent int64) osmomath.Dec {
	if exponent >= 0 {
		return powersOfTen[exponent]
	}
	return negPowersOfTen[-exponent]
}

func powTenBigDec(exponent int64) osmomath.BigDec {
	if exponent >= 0 {
		return bigPowersOfTen[exponent]
	}
	return bigNegPowersOfTen[-exponent]
}

// CalculatePriceToTick calculates tickIndex from price. Contrary to CalculatePriceToTickV1,
// it uses BigDec in internal calculations
func CalculatePriceToTick(price osmomath.BigDec) (tickIndex int64, err error) {
	if price.IsNegative() {
		return 0, fmt.Errorf("price must be greater than zero")
	}
	if price.GT(types.MaxSpotPriceBigDec) || price.LT(types.MinSpotPriceV2) {
		return 0, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPriceV2, MaxSpotPrice: types.MaxSpotPrice}
	}
	if price.Equal(osmomathBigOneDec) {
		return 0, nil
	}

	// N.B. this exists to maintain backwards compatibility with
	// the old version of the function that operated on decimal with precision of 18.
	if price.GTE(types.MinSpotPriceBigDec) {
		// It is acceptable to truncate price as the minimum we support is
		// 10**-12 which is above the smallest value of osmomath.Dec.
		price.ChopPrecisionMut(osmomath.DecPrecision)
	}

	// The approach here is to try determine which "geometric spacing" are we in.
	// There is one geometric spacing for every power of ten.
	// If price > 1, we search for the first geometric spacing w/ a max price greater than our price.
	// If price < 1, we search for the first geometric spacing w/ a min price smaller than our price.
	// TODO: We can optimize by using smarter search algorithms
	var geoSpacing *tickExpIndexData
	if price.GT(osmomathBigOneDec) {
		index := 0
		geoSpacing = tickExpCache[int64(index)]
		for geoSpacing.maxPrice.LT(price) {
			index += 1
			geoSpacing = tickExpCache[int64(index)]
		}
	} else {
		index := -1
		geoSpacing = tickExpCache[int64(index)]
		for geoSpacing.initialPrice.GT(price) {
			index -= 1
			geoSpacing = tickExpCache[int64(index)]
		}
	}

	// We know were between (geoSpacing.initialPrice, geoSpacing.endPrice)
	// The number of ticks that need to be filled by our current spacing is
	// (price - geoSpacing.initialPrice) / geoSpacing.additiveIncrementPerTick
	priceInThisExponent := price.Sub(geoSpacing.initialPrice)
	ticksFilledByCurrentSpacing := priceInThisExponent.QuoMut(geoSpacing.additiveIncrementPerTick)
	// we get the bucket index by:
	// * taking the bucket index of the smallest price in this tick
	// * adding to it the number of ticks filled by the current spacing
	tickIndex = ticksFilledByCurrentSpacing.TruncateInt64() + geoSpacing.initialTick
	return tickIndex, nil
}

// CalculateSqrtPriceToTick takes in a square root and returns the corresponding tick index.
// This function does not take into consideration tick spacing.
func CalculateSqrtPriceToTick(sqrtPrice osmomath.BigDec) (tickIndex int64, err error) {
	// SqrtPrice may have errors, so we take the tick obtained from the price
	// and move it in a +/- 1 tick range based on the sqrt price those ticks would imply.
	price := sqrtPrice.Mul(sqrtPrice)

	tick, err := CalculatePriceToTick(price)
	if err != nil {
		return 0, err
	}

	// TODO: remove this check. It is present to maintain backwards state-compatibility with
	// v19.x and earlier major releases of Osmosis.
	// Once https://github.com/osmosis-labs/osmosis/issues/5726 is fully complete,
	// this should be removed.
	//
	// Backwards state-compatibility is maintained by having the swap and LP logic error
	// here in case the price/tick falls below the origina minimum tick bounds that are
	// consistent with v19.x and earlier release lines.
	if tick < types.MinCurrentTick {
		return 0, types.TickIndexMinimumError{MinTick: types.MinCurrentTick}
	}

	// We have a candidate bucket index `t`. We discern here if:
	// * sqrtPrice in [ TickToSqrtPrice(t - 1), TickToSqrtPrice(t)     ) => bucket t - 1
	// * sqrtPrice in [ TickToSqrtPrice(t),     TickToSqrtPrice(t + 1) ) => bucket t
	// * sqrtPrice in [ TickToSqrtPrice(t + 1), TickToSqrtPrice(t + 2) ) => bucket t + 1
	// We handle boundary checks, by saying that if our candidate is the min tick,
	// set the candidate to min tick + 1.
	// If our candidate is at or above max tick - 1, set the candidate to max tick - 2.
	// This is because to check tick t + 1, we need to go to t + 2, so to not go over
	// max tick during these checks, we need to shift it down by 2.
	// We check this at max tick - 1 instead of max tick, since we expect the output to
	// have some error that can push us over the tick boundary.
	outOfBounds := false
	if tick <= types.MinInitializedTickV2 {
		tick = types.MinInitializedTickV2 + 1
		outOfBounds = true
	} else if tick >= types.MaxTick-1 {
		tick = types.MaxTick - 2
		outOfBounds = true
	}

	sqrtPriceTplus1, err := TickToSqrtPrice(tick + 1)
	if err != nil {
		return 0, types.ErrCalculateSqrtPriceToTick
	}
	// code path where sqrtPrice is either in tick t + 1, or out of bounds.
	if sqrtPrice.GTE(sqrtPriceTplus1) {
		// out of bounds check
		sqrtPriceTplus2, err := TickToSqrtPrice(tick + 2)
		if err != nil {
			return 0, types.ErrCalculateSqrtPriceToTick
		}
		// We error if sqrtPriceT is above sqrtPriceTplus2
		// For cases where calculated tick does not fall on a limit (min/max tick), the upper end is exclusive.
		// For cases where calculated tick falls on a limit, the upper end is inclusive, since the actual tick is
		// already shifted and making it exclusive would make min/max tick impossible to reach by construction.
		// We do this primary for code simplicity, as alternatives would require more branching and special cases.
		if (!outOfBounds && sqrtPrice.GTE(sqrtPriceTplus2)) || (outOfBounds && sqrtPrice.GT(sqrtPriceTplus2)) {
			return 0, types.SqrtPriceToTickError{OutOfBounds: outOfBounds}
		}

		// We expect this case to only be hit when the original provided sqrt price is exactly equal to the max sqrt price.
		if sqrtPrice.Equal(sqrtPriceTplus2) {
			return tick + 2, nil
		}
		// we are not out of bounds, therefore its tick t+1!
		return tick + 1, nil
	}

	// code path where sqrtPrice is either in tick t - 1, t, or out of bounds.
	// The out of bounds case here should never be possible, but we need to more rigorously check this
	// to delete that sqrt call.
	sqrtPriceT, err := TickToSqrtPrice(tick)
	if err != nil {
		return 0, types.ErrCalculateSqrtPriceToTick
	}
	// sqrtPriceT <= sqrtPrice < sqrtPriceTplus1, this were in bucket t
	if sqrtPrice.GTE(sqrtPriceT) {
		return tick, nil
	}

	// check we are not out of bounds from below.
	// TODO: Validate this case is impossible, and delete it
	sqrtPriceTmin1, err := TickToSqrtPrice(tick - 1)
	if err != nil {
		return 0, types.ErrCalculateSqrtPriceToTick
	}
	if sqrtPrice.LT(sqrtPriceTmin1) {
		return 0, types.SqrtPriceToTickError{OutOfBounds: outOfBounds}
	}

	return tick - 1, nil
}
