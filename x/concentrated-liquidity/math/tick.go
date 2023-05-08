package math

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	sdkNineDec = sdk.NewDec(9)
	sdkTenDec  = sdk.NewDec(10)
)

// TicksToSqrtPrice returns the sqrtPrice for the lower and upper ticks by
// individually calling `TickToSqrtPrice` method.
// Returns error if fails to calculate price.
func TicksToSqrtPrice(lowerTick, upperTick int64) (sdk.Dec, sdk.Dec, error) {
	if lowerTick >= upperTick {
		return sdk.Dec{}, sdk.Dec{}, types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	sqrtPriceUpperTick, err := TickToSqrtPrice(sdk.NewInt(upperTick))
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	sqrtPriceLowerTick, err := TickToSqrtPrice(sdk.NewInt(lowerTick))
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return sqrtPriceLowerTick, sqrtPriceUpperTick, nil
}

// TickToSqrtPrice returns the sqrtPrice given a tickIndex
// If tickIndex is zero, the function returns sdk.OneDec().
// It is the combination of calling TickToPrice followed by Sqrt.
func TickToSqrtPrice(tickIndex sdk.Int) (sdk.Dec, error) {
	price, err := TickToPrice(tickIndex)
	if err != nil {
		return sdk.Dec{}, err
	}

	// Determine the sqrtPrice from the price
	sqrtPrice, err := price.ApproxSqrt()
	if err != nil {
		return sdk.Dec{}, err
	}
	return sqrtPrice, nil
}

// TickToSqrtPrice returns the sqrtPrice given a tickIndex
// If tickIndex is zero, the function returns sdk.OneDec().
func TickToPrice(tickIndex sdk.Int) (price sdk.Dec, err error) {
	if tickIndex.IsZero() {
		return sdk.OneDec(), nil
	}

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	exponentAtPriceOne := types.ExponentAtPriceOne
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg()))

	// Check that the tick index is between min and max value
	if tickIndex.LT(sdk.NewInt(types.MinTick)) {
		return sdk.Dec{}, types.TickIndexMinimumError{MinTick: types.MinTick}
	}
	if tickIndex.GT(sdk.NewInt(types.MaxTick)) {
		return sdk.Dec{}, types.TickIndexMaximumError{MaxTick: types.MaxTick}
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
	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index assuming
// tick spacing of 1.
func PriceToTick(price sdk.Dec) (sdk.Int, error) {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt(), nil
	}

	if price.IsNegative() {
		return sdk.Int{}, fmt.Errorf("price must be greater than zero")
	}

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Int{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	// Determine the tick that corresponds to the price
	// This does not take into account the tickSpacing
	tickIndex := CalculatePriceToTick(price)

	return tickIndex, nil
}

// PriceToTickRoundDown takes a price and returns the corresponding tick index.
// If tickSpacing is provided, the tick index will be rounded down to the nearest multiple of tickSpacing.
func PriceToTickRoundDown(price sdk.Dec, tickSpacing uint64) (sdk.Int, error) {
	tickIndex, err := PriceToTick(price)
	if err != nil {
		return sdk.Int{}, err
	}

	// Round the tick index down to the nearest tick spacing if the tickIndex is in between authorized tick values
	tickSpacingInt := sdk.NewIntFromUint64(tickSpacing)
	tickIndexModulus := tickIndex.Mod(tickSpacingInt)
	if !tickIndexModulus.Equal(sdk.ZeroInt()) {
		tickIndex = tickIndex.Sub(tickIndexModulus)
	}

	// Defense-in-depth check to ensure that the tick index is within the authorized range
	// Should never get here.
	if tickIndex.GT(sdk.NewInt(types.MaxTick)) || tickIndex.LT(sdk.NewInt(types.MinTick)) {
		return sdk.Int{}, types.TickIndexNotWithinBoundariesError{ActualTick: tickIndex.Int64(), MinTick: types.MinTick, MaxTick: types.MaxTick}
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
		return osmomath.NewBigDec(10).PowerInteger(exponent.Uint64())
	}
	return osmomath.OneDec().Quo(osmomath.NewBigDec(10).PowerInteger(exponent.Abs().Uint64()))
}

// CalculatePriceToTick takes in a price and returns the corresponding tick index.
// This function does not take into consideration tick spacing.
func CalculatePriceToTick(price sdk.Dec) (tickIndex sdk.Int) {
	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	exponentAtPriceOne := types.ExponentAtPriceOne
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg()))

	// Initialize the current price to 1, the current precision to exponentAtPriceOne, and the number of ticks passed to 0
	currentPrice := sdk.OneDec()
	ticksPassed := sdk.ZeroInt()

	exponentAtCurrentTick := exponentAtPriceOne

	// Set the currentAdditiveIncrementInTicks to the exponentAtPriceOne
	currentAdditiveIncrementInTicks := powTenBigDec(exponentAtPriceOne)

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

	// Determine how many ticks we have passed in the exponentAtCurrentTick (in other words, the incomplete geometricExponent above)
	ticksToBeFulfilledByExponentAtCurrentTick := osmomath.BigDecFromSDKDec(price.Sub(currentPrice)).Quo(currentAdditiveIncrementInTicks)

	// Finally, add the ticks we have passed from the completed geometricExponent values, as well as the ticks we have passed in the current geometricExponent value
	tickIndex = ticksPassed.Add(ticksToBeFulfilledByExponentAtCurrentTick.SDKDec().RoundInt())
	return tickIndex
}
