package math

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

// TicksToSqrtPrice returns the sqrtPrice for the lower and upper ticks by
// individually calling `TickToSqrtPrice` method.
// Returns error if fails to calculate price.
func TicksToSqrtPrice(lowerTick, upperTick int64) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec, error) {
	if lowerTick >= upperTick {
		return sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	priceUpperTick, sqrtPriceUpperTick, err := TickToSqrtPrice(upperTick)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}
	priceLowerTick, sqrtPriceLowerTick, err := TickToSqrtPrice(lowerTick)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, sdk.Dec{}, err
	}
	return priceLowerTick, priceUpperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, nil
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
	sqrtPrice, err := price.ApproxSqrt()
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return price, sqrtPrice, nil
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
	if tickIndex < types.MinTick {
		return sdk.Dec{}, types.TickIndexMinimumError{MinTick: types.MinTick}
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

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}
	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index assuming
// tick spacing of 1.
func PriceToTick(price sdk.Dec) (int64, error) {
	if price.Equal(sdk.OneDec()) {
		return 0, nil
	}

	if price.IsNegative() {
		return 0, fmt.Errorf("price must be greater than zero")
	}

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return 0, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	// Determine the tick that corresponds to the price
	// This does not take into account the tickSpacing
	tickIndex := CalculatePriceToTick(price)

	return tickIndex, nil
}

// PriceToTickRoundDown takes a price and returns the corresponding tick index.
// If tickSpacing is provided, the tick index will be rounded down to the nearest multiple of tickSpacing.
// CONTRACT: tickSpacing must be smaller or equal to the max of 1 << 63 - 1.
// This is not a concern because we have authorized tick spacings that are smaller than this max,
// and we don't expect to ever require it to be this large.
func PriceToTickRoundDown(price sdk.Dec, tickSpacing uint64) (int64, error) {
	tickIndex, err := PriceToTick(price)
	if err != nil {
		return 0, err
	}

	// Round the tick index down to the nearest tick spacing if the tickIndex is in between authorized tick values
	// Note that this is Euclidean modulus.
	// The difference from default Go modulus is that Go default results
	// in a negative remainder when the dividend is negative.
	// Consider example tickIndex = -17, tickSpacing = 10
	// tickIndexModulus = tickIndex % tickSpacing = -7
	// tickIndexModulus = -7 + 10 = 3
	// tickIndex = -17 - 3 = -20
	tickIndexModulus := tickIndex % int64(tickSpacing)
	if tickIndexModulus < 0 {
		tickIndexModulus += int64(tickSpacing)
	}

	if tickIndexModulus != 0 {
		tickIndex = tickIndex - tickIndexModulus
	}

	// Defense-in-depth check to ensure that the tick index is within the authorized range
	// Should never get here.
	if tickIndex > types.MaxTick || tickIndex < types.MinTick {
		return 0, types.TickIndexNotWithinBoundariesError{ActualTick: tickIndex, MinTick: types.MinTick, MaxTick: types.MaxTick}
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

// CalculatePriceToTick takes in a price and returns the corresponding tick index.
// This function does not take into consideration tick spacing.
func CalculatePriceToTick(price sdk.Dec) (tickIndex int64) {
	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	exponentAtPriceOne := types.ExponentAtPriceOne
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne * -1)).TruncateInt64()

	// Initialize the current price to 1, the current precision to exponentAtPriceOne, and the number of ticks passed to 0
	currentPrice := sdkOneDec
	ticksPassed := int64(0)

	exponentAtCurrentTick := exponentAtPriceOne

	// Set the currentAdditiveIncrementInTicks to the exponentAtPriceOne
	currentAdditiveIncrementInTicks := powTenBigDec(exponentAtPriceOne)

	// Now, we loop through the exponentAtCurrentTicks until we have passed the price
	// Once we pass the price, we can determine what which geometric exponents we have filled in their entirety,
	// as well as how many ticks that corresponds to
	// In the opposite direction (price < 1), we do the same thing (just decrement the geometric exponent instead of incrementing).
	// The only difference is we must reduce the increment distance by a factor of 10.
	if price.GT(sdkOneDec) {
		for currentPrice.LT(price) {
			currentAdditiveIncrementInTicks = powTenBigDec(exponentAtCurrentTick)
			maxPriceForCurrentAdditiveIncrementInTicks := osmomath.NewBigDec(geometricExponentIncrementDistanceInTicks).Mul(currentAdditiveIncrementInTicks)
			currentPrice = currentPrice.Add(maxPriceForCurrentAdditiveIncrementInTicks.SDKDec())
			exponentAtCurrentTick = exponentAtCurrentTick + 1
			ticksPassed = ticksPassed + geometricExponentIncrementDistanceInTicks
		}
	} else {
		// We must decrement the exponentAtCurrentTick by one when traversing negative ticks in order to constantly step up in precision when going further down in ticks
		// Otherwise, from tick 0 to tick -(geometricExponentIncrementDistanceInTicks), we would use the same exponent as the exponentAtPriceOne
		exponentAtCurrentTick := exponentAtPriceOne - 1
		for currentPrice.GT(price) {
			currentAdditiveIncrementInTicks = powTenBigDec(exponentAtCurrentTick)
			maxPriceForCurrentAdditiveIncrementInTicks := osmomath.NewBigDec(geometricExponentIncrementDistanceInTicks).Mul(currentAdditiveIncrementInTicks)
			currentPrice = currentPrice.Sub(maxPriceForCurrentAdditiveIncrementInTicks.SDKDec())
			exponentAtCurrentTick = exponentAtCurrentTick - 1
			ticksPassed = ticksPassed - geometricExponentIncrementDistanceInTicks
		}
	}

	// Determine how many ticks we have passed in the exponentAtCurrentTick (in other words, the incomplete geometricExponent above)
	ticksToBeFulfilledByExponentAtCurrentTick := osmomath.BigDecFromSDKDec(price.Sub(currentPrice)).Quo(currentAdditiveIncrementInTicks)

	// Finally, add the ticks we have passed from the completed geometricExponent values, as well as the ticks we have passed in the current geometricExponent value
	tickIndex = ticksPassed + ticksToBeFulfilledByExponentAtCurrentTick.SDKDec().RoundInt64()
	return tickIndex
}
