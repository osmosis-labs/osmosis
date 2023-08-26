package math

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

// TicksToSqrtPrice returns the sqrtPrice for the lower and upper ticks by
// individually calling `TickToSqrtPrice` method.
// Returns error if fails to calculate price.
func TicksToSqrtPrice(lowerTick, upperTick int64) (sdk.Dec, sdk.Dec, error) {
	if lowerTick >= upperTick {
		return sdk.Dec{}, sdk.Dec{}, types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	_, sqrtPriceUpperTick, err := TickToSqrtPrice(upperTick)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	_, sqrtPriceLowerTick, err := TickToSqrtPrice(lowerTick)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return sqrtPriceLowerTick, sqrtPriceUpperTick, nil
}

// TickToSqrtPrice returns the sqrtPrice given a tickIndex
// If tickIndex is zero, the function returns sdk.OneDec().
// It is the combination of calling TickToPrice followed by Sqrt.
func TickToSqrtPrice(tickIndex int64) (sdk.Dec, sdk.Dec, error) {
	price, err := TickToPrice(tickIndex)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	// Determine the sqrtPrice from the price
	sqrtPrice, err := osmomath.MonotonicSqrt(price)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return price, sqrtPrice, nil
}

// TickToSqrtPriceBigDec returns the sqrtPrice given a tickIndex
// as a BigDec.
func TickToSqrtPriceBigDec(tickIndex int64) (osmomath.BigDec, error) {
	price, err := TickToPrice(tickIndex)
	if err != nil {
		return osmomath.BigDec{}, err
	}

	// Determine the sqrtPrice from the price
	sqrtPrice, err := osmomath.MonotonicSqrt(price)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	return osmomath.BigDecFromSDKDec(sqrtPrice), nil
}

// TickToPrice returns the price given a tickIndex
// If tickIndex is zero, the function returns sdk.OneDec().
func TickToPrice(tickIndex int64) (price sdk.Dec, err error) {
	if tickIndex == 0 {
		return sdk.OneDec(), nil
	}

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	exponentAtPriceOne := types.ExponentAtPriceOne
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(-exponentAtPriceOne)).TruncateInt64()

	// Check that the tick index is between min and max value
	if tickIndex < types.MinCurrentTick {
		return sdk.Dec{}, types.TickIndexMinimumError{MinTick: types.MinCurrentTick}
	}
	if tickIndex > types.MaxTick {
		return sdk.Dec{}, types.TickIndexMaximumError{MaxTick: types.MaxTick}
	}

	// Use floor division to determine what the geometricExponent is now (the delta)
	geometricExponentDelta := tickIndex / geometricExponentIncrementDistanceInTicks

	// Calculate the exponentAtCurrentTick from the starting exponentAtPriceOne and the geometricExponentDelta
	exponentAtCurrentTick := exponentAtPriceOne + geometricExponentDelta
	if tickIndex < 0 {
		// We must decrement the exponentAtCurrentTick when entering the negative tick range in order to constantly step up in precision when going further down in ticks
		// Otherwise, from tick 0 to tick -(geometricExponentIncrementDistanceInTicks), we would use the same exponent as the exponentAtPriceOne
		exponentAtCurrentTick = exponentAtCurrentTick - 1
	}

	// Knowing what our exponentAtCurrentTick is, we can then figure out what power of 10 this exponent corresponds to
	// We need to utilize bigDec here since increments can go beyond the 10^-18 limits set by the sdk
	currentAdditiveIncrementInTicks := powTenBigDec(exponentAtCurrentTick)

	// Now, starting at the minimum tick of the current increment, we calculate how many ticks in the current geometricExponent we have passed
	numAdditiveTicks := tickIndex - (geometricExponentDelta * geometricExponentIncrementDistanceInTicks)

	// Finally, we can calculate the price
	price = PowTenInternal(geometricExponentDelta).Add(osmomath.NewBigDec(numAdditiveTicks).Mul(currentAdditiveIncrementInTicks).SDKDec())

	// defense in depth, this logic would not be reached due to use having checked if given tick is in between
	// min tick and max tick.
	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}
	return price, nil
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
	if tickIndex > types.MaxTick || tickIndex < types.MinInitializedTick {
		return 0, types.TickIndexNotWithinBoundariesError{ActualTick: tickIndex, MinTick: types.MinInitializedTick, MaxTick: types.MaxTick}
	}

	return tickIndex, nil
}

// SqrtPriceToTickRoundDown converts the given sqrt price to its corresponding tick rounded down
// to the nearest tick spacing.
func SqrtPriceToTickRoundDownSpacing(sqrtPrice sdk.Dec, tickSpacing uint64) (int64, error) {
	tickIndex, err := CalculateSqrtPriceToTick(osmomath.BigDecFromSDKDec(sqrtPrice))
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
// This is because the sdk.Dec.Power function does not support negative exponents
func PowTenInternal(exponent int64) sdk.Dec {
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

func CalculatePriceToTickDec(price sdk.Dec) (tickIndex sdk.Dec, err error) {
	if price.IsNegative() {
		return sdk.ZeroDec(), fmt.Errorf("price must be greater than zero")
	}
	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.ZeroDec(), types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}
	if price.Equal(sdkOneDec) {
		return sdk.ZeroDec(), nil
	}

	// The approach here is to try determine which "geometric spacing" are we in.
	// There is one geometric spacing for every power of ten.
	// If price > 1, we search for the first geometric spacing w/ a max price greater than our price.
	// If price < 1, we search for the first geometric spacing w/ a min price smaller than our price.
	// TODO: We can optimize by using smarter search algorithms
	var geoSpacing *tickExpIndexData
	if price.GT(sdkOneDec) {
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
	priceInThisExponent := osmomath.BigDecFromSDKDec(price.Sub(geoSpacing.initialPrice))
	ticksFilledByCurrentSpacing := priceInThisExponent.Quo(geoSpacing.additiveIncrementPerTick)
	// we get the bucket index by:
	// * taking the bucket index of the smallest price in this tick
	// * adding to it the number of ticks filled by the current spacing
	// We leave rounding of this to the caller
	// (NOTE: You'd expect it to be number of ticks "completely" filled by the current spacing,
	// which would be truncation. However price may have errors, hence it being callers job)
	tickIndex = ticksFilledByCurrentSpacing.SDKDec()
	tickIndex = tickIndex.Add(sdk.NewDec(geoSpacing.initialTick))
	return tickIndex, nil
}

// CalculateSqrtPriceToTick takes in a square root and returns the corresponding tick index.
// This function does not take into consideration tick spacing.
func CalculateSqrtPriceToTick(sqrtPrice osmomath.BigDec) (tickIndex int64, err error) {
	// SqrtPrice may have errors, so we take the tick obtained from the price
	// and move it in a +/- 1 tick range based on the sqrt price those ticks would imply.
	price := sqrtPrice.Mul(sqrtPrice)

	// It is acceptable to truncate price as the minimum we support is
	// 10**-12 which is above the smallest value of sdk.Dec.
	priceDec := price.SDKDec()

	tick, err := CalculatePriceToTickDec(priceDec)
	if err != nil {
		return 0, err
	}

	truncatedTick := tick.TruncateInt64()

	// We have a candidate bucket index `t`. We discern here if:
	// * sqrtPrice in [TickToSqrtPrice(t - 1), TickToSqrtPrice(t))
	// * sqrtPrice in [TickToSqrtPrice(t), TickToSqrtPrice(t + 1))
	// * sqrtPrice in [TickToSqrtPrice(t+1), TickToSqrtPrice(t + 2))
	// * sqrtPrice not in either.
	// We handle boundary checks, by saying that if our candidate is the min tick,
	// set the candidate to min tick + 1.
	// If our candidate is at or above max tick - 1, set the candidate to max tick - 2.
	// This is because to check tick t + 1, we need to go to t + 2, so to not go over
	// max tick during these checks, we need to shift it down by 2.
	// We check this at max tick - 1 instead of max tick, since we expect the output to
	// have some error that can push us over the tick boundary.
	outOfBounds := false
	if truncatedTick <= types.MinInitializedTick {
		truncatedTick = types.MinInitializedTick + 1
		outOfBounds = true
	} else if truncatedTick >= types.MaxTick-1 {
		truncatedTick = types.MaxTick - 2
		outOfBounds = true
	}

	sqrtPriceTmin1, errM1 := TickToSqrtPriceBigDec(truncatedTick - 1)
	sqrtPriceT, errT := TickToSqrtPriceBigDec(truncatedTick)
	sqrtPriceTplus1, errP1 := TickToSqrtPriceBigDec(truncatedTick + 1)
	sqrtPriceTplus2, errP2 := TickToSqrtPriceBigDec(truncatedTick + 2)
	if errM1 != nil || errT != nil || errP1 != nil || errP2 != nil {
		return 0, errors.New("internal error in computing square roots within CalculateSqrtPriceToTick")
	}

	// We error if sqrtPriceT is above sqrtPriceTplus2 or below sqrtPriceTmin1.
	// For cases where calculated tick does not fall on a limit (min/max tick), the upper end is exclusive.
	// For cases where calculated tick falls on a limit, the upper end is inclusive, since the actual tick is
	// already shifted and making it exclusive would make min/max tick impossible to reach by construction.
	// We do this primary for code simplicity, as alternatives would require more branching and special cases.
	if (!outOfBounds && sqrtPrice.GTE(sqrtPriceTplus2)) || (outOfBounds && sqrtPrice.GT(sqrtPriceTplus2)) || sqrtPrice.LT(sqrtPriceTmin1) {
		return 0, types.SqrtPriceToTickError{OutOfBounds: outOfBounds}
	}

	// We expect this case to only be hit when the original provided sqrt price is exactly equal to the max sqrt price.
	if sqrtPrice.Equal(sqrtPriceTplus2) {
		return truncatedTick + 2, nil
	}

	// The remaining cases handle shifting tick index by +/- 1.
	if sqrtPrice.GTE(sqrtPriceTplus1) {
		return truncatedTick + 1, nil
	}
	if sqrtPrice.GTE(sqrtPriceT) {
		return truncatedTick, nil
	}
	return truncatedTick - 1, nil
}
