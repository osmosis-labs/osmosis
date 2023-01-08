package math

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

var sdkNineDec = sdk.NewDec(9)
var sdkTenDec = sdk.NewDec(10)

// TicksToPrice returns the price for the lower and upper ticks.
// Returns error if fails to calculate price.
// TODO: spec and tests
func TicksToPrice(lowerTick, upperTick int64, exponentAtPriceOne sdk.Int) (sdk.Dec, sdk.Dec, error) {
	priceUpperTick, err := TickToPrice(sdk.NewInt(upperTick), exponentAtPriceOne)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	priceLowerTick, err := TickToPrice(sdk.NewInt(lowerTick), exponentAtPriceOne)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return priceLowerTick, priceUpperTick, nil
}

// TickToPrice returns the price given the following two arguments:
// 	- tickIndex: the tick index to calculate the price for
// 	- exponentAtPriceOne: the value of the exponent (and therefore the precision) at which the starting price of 1 is set
//
// If tickIndex is zero, the function returns sdk.OneDec().
func TickToPrice(tickIndex, exponentAtPriceOne sdk.Int) (price sdk.Dec, err error) {
	if tickIndex.IsZero() {
		return sdk.OneDec(), nil
	}

	if exponentAtPriceOne.LT(types.PrecisionValueAtPriceOneMin) || exponentAtPriceOne.GT(types.PrecisionValueAtPriceOneMax) {
		return sdk.Dec{}, fmt.Errorf("exponentAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax)
	}

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(powTen(exponentAtPriceOne.Neg()))

	// If the price is below 1, we decrement the increment distance by a factor of 10
	if tickIndex.IsNegative() {
		geometricExponentIncrementDistanceInTicks = geometricExponentIncrementDistanceInTicks.Quo(sdkTenDec)
	}

	// Check that the tick index is between min and max value for the given exponentAtPriceOne
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne, geometricExponentIncrementDistanceInTicks)
	if tickIndex.LT(minTick) {
		return sdk.Dec{}, fmt.Errorf("tickIndex must be greater than or equal to %s", minTick)
	}
	if tickIndex.GT(maxTick) {
		return sdk.Dec{}, fmt.Errorf("tickIndex must be less than or equal to %s", maxTick)
	}

	// Use floor division to determine what the geometricExponent is now (the delta)
	geometricExponentDelta := tickIndex.ToDec().Quo(geometricExponentIncrementDistanceInTicks).TruncateInt()

	// Calculate the exponentAtCurrentTick from the starting exponentAtPriceOne and the geometricExponentDelta
	exponentAtCurrentTick := exponentAtPriceOne.Add(geometricExponentDelta)

	// Knowing what our exponentAtCurrentTick is, we can then figure out what power of 10 this exponent corresponds to
	currentAdditiveIncrementInTicks := powTen(exponentAtCurrentTick)

	// Now, starting at the minimum tick of the current increment, we calculate how many ticks in the current geometricExponent we have passed
	numAdditiveTicks := tickIndex.ToDec().Sub(geometricExponentDelta.ToDec()).Mul(geometricExponentIncrementDistanceInTicks)

	// Finally, we can calculate the price
	price = powTen(geometricExponentDelta).Add(numAdditiveTicks.Mul(currentAdditiveIncrementInTicks))

	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec, exponentAtPriceOne sdk.Int) (sdk.Int, error) {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt(), nil
	}

	if price.IsNegative() {
		return sdk.Int{}, fmt.Errorf("price must be greater than zero")
	}

	if exponentAtPriceOne.LT(types.PrecisionValueAtPriceOneMin) || exponentAtPriceOne.GT(types.PrecisionValueAtPriceOneMax) {
		return sdk.Int{}, fmt.Errorf("exponentAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax)
	}

	// The formula is as follows: geometricExponentIncrementDistanceInTicks = 9 * 10**(-exponentAtPriceOne)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**exponentAtPriceOne)
	geometricExponentIncrementDistanceInTicks := sdkNineDec.Mul(powTen(exponentAtPriceOne.Neg()))

	// If the price is less than 1, we must reduce the increment distance by a factor of 10
	if price.LT(sdk.OneDec()) {
		geometricExponentIncrementDistanceInTicks = geometricExponentIncrementDistanceInTicks.Quo(sdkTenDec)
	}

	// Initialize the total price to 1, the current precision to exponentAtPriceOne, and the number of ticks passed to 0
	totalPrice := sdk.OneDec()
	ticksPassed := sdk.ZeroInt()
	exponentAtCurrentTick := exponentAtPriceOne

	// Set the currentAdditiveIncrementInTicks to the exponentAtPriceOne
	currentAdditiveIncrementInTicks := powTen(exponentAtPriceOne)

	// Now, we loop through the k increments until we have passed the price
	// Once we pass the price, we can determine what which geometric exponents we have filled in their entirety,
	// as well as how many ticks that corresponds to
	// In the opposite direction (price < 1), we do the same thing (just decrement the geometric exponent instead of incrementing).
	// The only difference is we must reduce the increment distance by a factor of 10.
	if price.GT(sdk.OneDec()) {
		for totalPrice.LT(price) {
			currentAdditiveIncrementInTicks = powTen(exponentAtCurrentTick)
			maxPriceForcurrentAdditiveIncrementInTicks := geometricExponentIncrementDistanceInTicks.Mul(currentAdditiveIncrementInTicks)
			totalPrice = totalPrice.Add(maxPriceForcurrentAdditiveIncrementInTicks)
			exponentAtCurrentTick = exponentAtCurrentTick.Add(sdk.OneInt())
			ticksPassed = ticksPassed.Add(geometricExponentIncrementDistanceInTicks.TruncateInt())
		}
	} else {
		for totalPrice.GT(price) {
			currentAdditiveIncrementInTicks = powTen(exponentAtCurrentTick)
			maxPriceForcurrentAdditiveIncrementInTicks := geometricExponentIncrementDistanceInTicks.Mul(currentAdditiveIncrementInTicks)
			totalPrice = totalPrice.Sub(maxPriceForcurrentAdditiveIncrementInTicks)
			exponentAtCurrentTick = exponentAtCurrentTick.Sub(sdk.OneInt())
			ticksPassed = ticksPassed.Sub(geometricExponentIncrementDistanceInTicks.TruncateInt())
		}
	}
	// Determine how many ticks we have passed in the exponentAtCurrentTick
	ticksToBeFulfilledByexponentAtCurrentTick := price.Sub(totalPrice).Quo(currentAdditiveIncrementInTicks)

	// Finally, add the ticks we have passed from the completed geometricExponent values, as well as the ticks we have passed in the current geometricExponent value
	tickIndex := ticksPassed.Add(ticksToBeFulfilledByexponentAtCurrentTick.TruncateInt())

	// Add a check to make sure that the tick index is within the allowed range
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne, geometricExponentIncrementDistanceInTicks)
	if tickIndex.LT(minTick) {
		return sdk.Int{}, fmt.Errorf("tickIndex must be greater than or equal to %s", minTick)
	}
	if tickIndex.GT(maxTick) {
		return sdk.Int{}, fmt.Errorf("tickIndex must be less than or equal to %s", maxTick)
	}

	return tickIndex, nil
}

// handleNegativeExponents treats negative exponents as 1/(10**|exponent|) instead of 10**-exponent
// This is because the sdk.Dec.Power function does not support negative exponents
func powTen(exponent sdk.Int) sdk.Dec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdkTenDec.Power(exponent.Uint64())
	}
	return sdk.OneDec().Quo(sdkTenDec.Power(exponent.Abs().Uint64()))
}

// GetMinAndMaxTicksFromExponentAtPriceOne determines min and max ticks allowed for a given exponentAtPriceOne value
// This allows for a min spot price of 0.000000000000000001 and a max spot price of 200000000000 for every k value
func GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne sdk.Int, geometricExponentIncrementDistanceInTicks sdk.Dec) (minTick, maxTick sdk.Int) {
	minTick = sdk.NewDec(18).Mul(geometricExponentIncrementDistanceInTicks).Neg().RoundInt()
	maxTick = powTen(exponentAtPriceOne.Neg().Add(sdk.NewInt(2))).RoundInt()
	return minTick, maxTick
}
