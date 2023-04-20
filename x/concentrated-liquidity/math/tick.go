package math

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	sdkNineDec        = sdk.NewDec(9)
	sdkTenDec         = sdk.NewDec(10)
	sdkEighteenDec    = sdk.NewDec(18)
	sdkThirtyEightDec = sdk.NewDec(38)
)

// TicksToSqrtPrice returns the sqrtPrice for the lower and upper ticks.
// Returns error if fails to calculate price.
// TODO: spec and tests
func TicksToSqrtPrice(lowerTick, upperTick int64, exponentAtPriceOne sdk.Int) (sdk.Dec, sdk.Dec, error) {
	sqrtPriceUpperTick, err := TickToSqrtPrice(sdk.NewInt(upperTick), exponentAtPriceOne)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	sqrtPriceLowerTick, err := TickToSqrtPrice(sdk.NewInt(lowerTick), exponentAtPriceOne)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return sqrtPriceLowerTick, sqrtPriceUpperTick, nil
}

// TickToSqrtPrice returns the sqrtPrice given the following two arguments:
//   - tickIndex: the tick index to calculate the price for
//   - exponentAtPriceOne: the value of the exponent (and therefore the precision) at which the starting price of 1 is set
//
// If tickIndex is zero, the function returns sdk.OneDec().
func TickToSqrtPrice(tickIndex, exponentAtPriceOne sdk.Int) (price sdk.Dec, err error) {
	if tickIndex.IsZero() {
		return sdk.OneDec(), nil
	}

	if exponentAtPriceOne.LT(types.ExponentAtPriceOneMin) || exponentAtPriceOne.GT(types.ExponentAtPriceOneMax) {
		return sdk.Dec{}, types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: exponentAtPriceOne, PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax}
	}

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg()))

	// Check that the tick index is between min and max value for the given exponentAtPriceOne
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)
	if tickIndex.LT(sdk.NewInt(minTick)) {
		return sdk.Dec{}, types.TickIndexMinimumError{MinTick: minTick}
	}
	if tickIndex.GT(sdk.NewInt(maxTick)) {
		return sdk.Dec{}, types.TickIndexMaximumError{MaxTick: maxTick}
	}

	// Use floor division to determine what the geometricExponent is now (the delta)
	geometricExponentDelta := tickIndex.ToDec().QuoIntMut(geometricExponentIncrementDistanceInTicks.TruncateInt()).TruncateInt()

	// Calculate the exponentAtCurrentTick from the starting exponentAtPriceOne and the geometricExponentDelta
	exponentAtCurrentTick := exponentAtPriceOne.Add(geometricExponentDelta)
	if tickIndex.IsNegative() {
		// We must decrement the exponentAtCurrentTick when entering the negative tick range in order to constantly step up in precision when going further down in ticks
		// Otherwise, from tick 0 to tick -(geometricExponentIncrementDistanceInTicks), we would use the same exponent as the exponentAtPriceOne
		exponentAtCurrentTick = exponentAtCurrentTick.Sub(sdk.OneInt())
	}

	// Knowing what our exponentAtCurrentTick is, we can then figure out what power of 10 this exponent corresponds to
	// We need to utilize bigDec here since increments can go beyond the 10^-18 limits set by the sdk
	currentAdditiveIncrementInTicks := powTenBigDec(exponentAtCurrentTick)

	// Now, starting at the minimum tick of the current increment, we calculate how many ticks in the current geometricExponent we have passed
	numAdditiveTicks := tickIndex.ToDec().Sub(geometricExponentDelta.ToDec().Mul(geometricExponentIncrementDistanceInTicks))

	// Finally, we can calculate the price
	price = PowTenInternal(geometricExponentDelta).Add(osmomath.BigDecFromSDKDec(numAdditiveTicks).Mul(currentAdditiveIncrementInTicks).SDKDec())

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	// Determine the sqrtPrice from the price
	sqrtPrice, err := price.ApproxSqrt()
	if err != nil {
		return sdk.Dec{}, err
	}
	return sqrtPrice, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec, exponentAtPriceOne sdk.Int) (sdk.Int, error) {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt(), nil
	}

	if price.IsNegative() {
		return sdk.Int{}, fmt.Errorf("price must be greater than zero")
	}

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Int{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	if exponentAtPriceOne.LT(types.ExponentAtPriceOneMin) || exponentAtPriceOne.GT(types.ExponentAtPriceOneMax) {
		return sdk.Int{}, types.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: exponentAtPriceOne, PrecisionValueAtPriceOneMin: types.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: types.ExponentAtPriceOneMax}
	}

	// Determine the number of ticks that have passed in the completed geometricExponents
	// If the ticksPassed do not completely fill the geometricExponent, then we do not count them and calculate the remaining ticks afterwards.
	currentPrice, ticksPassed, currentAdditiveIncrementInTicks := CalculatePriceAndTicksPassed(price, exponentAtPriceOne)

	// Determine how many ticks we have passed in the exponentAtCurrentTick (in other words, the incomplete geometricExponent above)
	ticksToBeFulfilledByExponentAtCurrentTick := osmomath.BigDecFromSDKDec(price.Sub(currentPrice)).Quo(currentAdditiveIncrementInTicks)

	// Finally, add the ticks we have passed from the completed geometricExponent values, as well as the ticks we have passed in the current geometricExponent value
	tickIndex := ticksPassed.Add(ticksToBeFulfilledByExponentAtCurrentTick.SDKDec().TruncateInt())

	// Add a check to make sure that the tick index is within the allowed range
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)
	if tickIndex.LT(sdk.NewInt(minTick)) {
		return sdk.Int{}, types.TickIndexMinimumError{MinTick: minTick}
	}
	if tickIndex.GT(sdk.NewInt(maxTick)) {
		return sdk.Int{}, types.TickIndexMaximumError{MaxTick: maxTick}
	}

	return tickIndex, nil
}

// powTen treats negative exponents as 1/(10**|exponent|) instead of 10**-exponent
// This is because the sdk.Dec.Power function does not support negative exponents
func PowTenInternal(exponent sdk.Int) sdk.Dec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdkTenDec.Power(exponent.Uint64())
	}
	return sdk.OneDec().Quo(sdkTenDec.Power(exponent.Abs().Uint64()))
}

func powTenBigDec(exponent sdk.Int) osmomath.BigDec {
	if exponent.GTE(sdk.ZeroInt()) {
		return osmomath.NewBigDec(10).Power(osmomath.NewBigDec(exponent.Int64()))
	}
	return osmomath.OneDec().Quo(osmomath.NewBigDec(10).Power(osmomath.NewBigDec(exponent.Abs().Int64())))
}

// ComputeMinAndMaxTicksFromExponentAtPriceOneInternal determines min and max ticks allowed for a given exponentAtPriceOne value
// This allows for a min spot price of 0.000000000000000001 and a max spot price of 100000000000000000000000000000000000000 for every exponentAtPriceOne value
func ComputeMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne sdk.Int) (minTick, maxTick int64) {
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg()))
	minTick = sdkEighteenDec.Mul(geometricExponentIncrementDistanceInTicks).Neg().RoundInt64()
	maxTick = sdkThirtyEightDec.Mul(geometricExponentIncrementDistanceInTicks).TruncateInt64()
	return minTick, maxTick
}

// GetMinAndMaxTicksFromExponentAtPriceOneInternal retrieves min and max ticks allowed for a given exponentAtPriceOne value
func GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne sdk.Int) (minTick, maxTick int64) {
	switch exponentAtPriceOne {
	default:
		return ComputeMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)
	case sdk.NewInt(-12):
		return types.MinTickNegTwelve, types.MaxTickNegTwelve
	case sdk.NewInt(-11):
		return types.MinTickNegEleven, types.MaxTickNegEleven
	case sdk.NewInt(-10):
		return types.MinTickNegTen, types.MaxTickNegTen
	case sdk.NewInt(-9):
		return types.MinTickNegNine, types.MaxTickNegNine
	case sdk.NewInt(-8):
		return types.MinTickNegEight, types.MaxTickNegEight
	case sdk.NewInt(-7):
		return types.MinTickNegSeven, types.MaxTickNegSeven
	case sdk.NewInt(-6):
		return types.MinTickNegSix, types.MaxTickNegSix
	case sdk.NewInt(-5):
		return types.MinTickNegFive, types.MaxTickNegFive
	case sdk.NewInt(-4):
		return types.MinTickNegFour, types.MaxTickNegFour
	case sdk.NewInt(-3):
		return types.MinTickNegThree, types.MaxTickNegThree
	case sdk.NewInt(-2):
		return types.MinTickNegTwo, types.MaxTickNegTwo
	case sdk.NewInt(-1):
		return types.MinTickNegOne, types.MaxTickNegOne
	}
}

// calculatePriceAndTicksPassed takes in a price and an exponentAtPriceOne, and returns the currentPrice, ticksPassed, and currentAdditiveIncrementInTicks.
// The function uses the geometricExponentIncrementDistanceInTicks formula to determine the number of ticks passed and the current additive increment in ticks.
// If the price is greater than 1, the function increments the exponentAtCurrentTick until the currentPrice is greater than the input price.
// If the price is less than 1, the function decrements the exponentAtCurrentTick until the currentPrice is less than the input price.
func CalculatePriceAndTicksPassed(price sdk.Dec, exponentAtPriceOne sdk.Int) (currentPrice sdk.Dec, ticksPassed sdk.Int, currentAdditiveIncrementInTicks osmomath.BigDec) {
	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg()))

	// Initialize the current price to 1, the current precision to exponentAtPriceOne, and the number of ticks passed to 0
	currentPrice = sdk.OneDec()
	ticksPassed = sdk.ZeroInt()
	exponentAtCurrentTick := exponentAtPriceOne

	// Set the currentAdditiveIncrementInTicks to the exponentAtPriceOne
	currentAdditiveIncrementInTicks = powTenBigDec(exponentAtPriceOne)

	// Now, we loop through the exponentAtCurrentTicks until we have passed the price
	// Once we pass the price, we can determine what which geometric exponents we have filled in their entirety,
	// as well as how many ticks that corresponds to
	// In the opposite direction (price < 1), we do the same thing (just decrement the geometric exponent instead of incrementing).
	// The only difference is we must reduce the increment distance by a factor of 10.
	if price.GT(sdk.OneDec()) {
		for currentPrice.LT(price) {
			currentAdditiveIncrementInTicks = powTenBigDec(exponentAtCurrentTick)
			maxPriceForCurrentAdditiveIncrementInTicks := osmomath.BigDecFromSDKDec(geometricExponentIncrementDistanceInTicks).Mul(currentAdditiveIncrementInTicks)
			currentPrice = currentPrice.Add(maxPriceForCurrentAdditiveIncrementInTicks.SDKDec())
			exponentAtCurrentTick = exponentAtCurrentTick.Add(sdk.OneInt())
			ticksPassed = ticksPassed.Add(geometricExponentIncrementDistanceInTicks.TruncateInt())
		}
	} else {
		// We must decrement the exponentAtCurrentTick by one when traversing negative ticks in order to constantly step up in precision when going further down in ticks
		// Otherwise, from tick 0 to tick -(geometricExponentIncrementDistanceInTicks), we would use the same exponent as the exponentAtPriceOne
		exponentAtCurrentTick := exponentAtPriceOne.Sub(sdk.OneInt())
		for currentPrice.GT(price) {
			currentAdditiveIncrementInTicks = powTenBigDec(exponentAtCurrentTick)
			maxPriceForCurrentAdditiveIncrementInTicks := osmomath.BigDecFromSDKDec(geometricExponentIncrementDistanceInTicks).Mul(currentAdditiveIncrementInTicks)
			currentPrice = currentPrice.Sub(maxPriceForCurrentAdditiveIncrementInTicks.SDKDec())
			exponentAtCurrentTick = exponentAtCurrentTick.Sub(sdk.OneInt())
			ticksPassed = ticksPassed.Sub(geometricExponentIncrementDistanceInTicks.TruncateInt())
		}
	}
	return currentPrice, ticksPassed, currentAdditiveIncrementInTicks
}
