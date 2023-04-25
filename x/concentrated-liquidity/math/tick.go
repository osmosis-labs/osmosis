package math

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	sdkNineDec = osmomath.NewBigDec(9)
	sdkTenDec  = osmomath.NewBigDec(10)
)

// TicksToSqrtPrice returns the sqrtPrice for the lower and upper ticks.
// Returns error if fails to calculate price.
// TODO: spec and tests
func TicksToSqrtPrice(lowerTick, upperTick int64) (sdk.Dec, sdk.Dec, error) {
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

	fmt.Println("price", price)

	// Determine the sqrtPrice from the price
	sqrtPrice, err := Sqrt(price)
	if err != nil {
		return sdk.Dec{}, err
	}
	return sqrtPrice, nil
}

// TickToSqrtPrice returns the price given the following two arguments:
//   - tickIndex: the tick index to calculate the price for
//   - exponentAtPriceOne: the value of the exponent (and therefore the precision) at which the starting price of 1 is set
//
// If tickIndex is zero, the function returns sdk.OneDec().
func TickToPrice(tickIndex sdk.Int) (sdk.Dec, error) {
	if tickIndex.IsZero() {
		return sdk.OneDec(), nil
	}
	// Check that the tick index is between min and max value
	if tickIndex.LT(sdk.NewInt(types.MinTick)) {
		return sdk.Dec{}, types.TickIndexMinimumError{MinTick: types.MinTick}
	}
	if tickIndex.GT(sdk.NewInt(types.MaxTick)) {
		return sdk.Dec{}, types.TickIndexMaximumError{MaxTick: types.MaxTick}
	}

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	exponentAtPriceOne := types.ExponentAtPriceOne
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg())).SDKDec().TruncateInt()

	// Use floor division to determine what the geometricExponent is now (the delta)
	geometricExponentDelta := tickIndex.ToDec().QuoIntMut(geometricExponentIncrementDistanceInTicks).TruncateInt()

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
	numAdditiveTicks := osmomath.NewBigDec(tickIndex.Int64()).Sub(osmomath.NewBigDec(geometricExponentDelta.Int64()).Mul(osmomath.BigDecFromSDKDec(geometricExponentIncrementDistanceInTicks.ToDec())))

	// Finally, we can calculate the price
	price := PowTenInternal(geometricExponentDelta).Add(numAdditiveTicks.Mul(currentAdditiveIncrementInTicks)).SDKDec()

	if price.GT(types.MaxSpotPrice) || price.LT(types.MinSpotPrice) {
		return sdk.Dec{}, types.PriceBoundError{ProvidedPrice: price, MinSpotPrice: types.MinSpotPrice, MaxSpotPrice: types.MaxSpotPrice}
	}

	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index.
// If tickSpacing is provided, the tick index will be rounded up to the nearest multiple of tickSpacing.
func PriceToTick(price sdk.Dec, tickSpacing uint64) (sdk.Int, error) {
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

	// Round the tick index up to the nearest tick spacing if the tickIndex is in between authorized tick values
	tickSpacingInt := sdk.NewIntFromUint64(tickSpacing)
	tickIndexRemainder := tickIndex.Mod(sdk.NewIntFromUint64(tickSpacing))
	if !tickIndexRemainder.Equal(sdk.ZeroInt()) {
		tickIndex = tickIndex.Add(tickSpacingInt.Sub(tickIndexRemainder))
	}

	return tickIndex, nil
}

// powTen treats negative exponents as 1/(10**|exponent|) instead of 10**-exponent
// This is because the sdk.Dec.Power function does not support negative exponents
func PowTenInternal(exponent sdk.Int) osmomath.BigDec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdkTenDec.PowerInteger(exponent.Uint64())
	}
	return osmomath.OneDec().Quo(sdkTenDec.PowerInteger(exponent.Abs().Uint64()))
}

func powTenBigDec(exponent sdk.Int) osmomath.BigDec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdkTenDec.PowerInteger(exponent.Uint64())
	}
	return osmomath.OneDec().Quo(osmomath.NewBigDec(10).PowerInteger(exponent.Abs().Uint64()))
}

// CalculatePriceToTick takes in a price and returns the corresponding tick index.
// This function does not take into consideration tick spacing.
func CalculatePriceToTick(price sdk.Dec) sdk.Int {

	priceBigDec := osmomath.BigDecFromSDKDec(price)

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	exponentAtPriceOne := types.ExponentAtPriceOne
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(PowTenInternal(exponentAtPriceOne.Neg()))

	// Initialize the current price to 1, the current precision to exponentAtPriceOne, and the number of ticks passed to 0
	currentPrice := osmomath.OneDec()
	ticksPassed := osmomath.ZeroInt()

	exponentAtCurrentTick := exponentAtPriceOne

	// Set the currentAdditiveIncrementInTicks to the exponentAtPriceOne
	currentAdditiveIncrementInTicks := powTenBigDec(exponentAtPriceOne)

	// Now, we loop through the exponentAtCurrentTicks until we have passed the price
	// Once we pass the price, we can determine what which geometric exponents we have filled in their entirety,
	// as well as how many ticks that corresponds to
	// In the opposite direction (price < 1), we do the same thing (just decrement the geometric exponent instead of incrementing).
	// The only difference is we must reduce the increment distance by a factor of 10.
	if price.GT(sdk.OneDec()) {
		for currentPrice.LT(priceBigDec) {
			currentAdditiveIncrementInTicks = powTenBigDec(exponentAtCurrentTick)
			maxPriceForCurrentAdditiveIncrementInTicks := geometricExponentIncrementDistanceInTicks.Mul(currentAdditiveIncrementInTicks)
			currentPrice = currentPrice.Add(maxPriceForCurrentAdditiveIncrementInTicks)
			exponentAtCurrentTick = exponentAtCurrentTick.Add(sdk.OneInt())
			ticksPassed = ticksPassed.Add(geometricExponentIncrementDistanceInTicks.TruncateInt())
		}
	} else {
		// We must decrement the exponentAtCurrentTick by one when traversing negative ticks in order to constantly step up in precision when going further down in ticks
		// Otherwise, from tick 0 to tick -(geometricExponentIncrementDistanceInTicks), we would use the same exponent as the exponentAtPriceOne
		exponentAtCurrentTick := exponentAtPriceOne.Sub(sdk.OneInt())
		for currentPrice.GT(priceBigDec) {
			currentAdditiveIncrementInTicks = powTenBigDec(exponentAtCurrentTick)
			maxPriceForCurrentAdditiveIncrementInTicks := geometricExponentIncrementDistanceInTicks.Mul(currentAdditiveIncrementInTicks)
			currentPrice = currentPrice.Sub(maxPriceForCurrentAdditiveIncrementInTicks)
			exponentAtCurrentTick = exponentAtCurrentTick.Sub(sdk.OneInt())
			ticksPassed = ticksPassed.Sub(geometricExponentIncrementDistanceInTicks.TruncateInt())
		}
	}

	// // truncate residue from math
	// + Precision of 24

	// Determine how many ticks we have passed in the exponentAtCurrentTick (in other words, the incomplete geometricExponent above)
	ticksToBeFulfilledByExponentAtCurrentTick := priceBigDec.Sub(currentPrice).Quo(currentAdditiveIncrementInTicks)

	// Finally, add the ticks we have passed from the completed geometricExponent values, as well as the ticks we have passed in the current geometricExponent value
	tickIndex := ticksPassed.Add(ticksToBeFulfilledByExponentAtCurrentTick.RoundInt())

	return tickIndex.ToDec().SDKDec().TruncateInt()
}
