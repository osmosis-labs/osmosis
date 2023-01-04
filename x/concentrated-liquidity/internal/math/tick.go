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
func TicksToPrice(lowerTick, upperTick int64, kAtPriceOne sdk.Int) (sdk.Dec, sdk.Dec, error) {
	priceUpperTick, err := TickToPrice(sdk.NewInt(upperTick), kAtPriceOne)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	priceLowerTick, err := TickToPrice(sdk.NewInt(lowerTick), kAtPriceOne)
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return priceLowerTick, priceUpperTick, nil
}

// TickToPrice returns the price given the following two arguments:
// 	- tickIndex: the tick index to calculate the price for
// 	- kAtPriceOne: the value of k at which the starting price of 1 is set
//
// If tickIndex is zero, the function returns sdk.OneDec().
func TickToPrice(tickIndex, kAtPriceOne sdk.Int) (price sdk.Dec, err error) {
	if tickIndex.IsZero() {
		return sdk.OneDec(), nil
	}

	if kAtPriceOne.LT(types.PrecisionValueAtPriceOneMin) || kAtPriceOne.GT(types.PrecisionValueAtPriceOneMax) {
		return sdk.Dec{}, fmt.Errorf("kAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax)
	}

	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	kIncrementDistance := sdkNineDec.Mul(handleNegativeExponents(kAtPriceOne.Neg()))

	// If the price is below 1, we decrement the increment distance by a factor of 10
	if tickIndex.IsNegative() {
		kIncrementDistance = kIncrementDistance.Quo(sdkTenDec)
	}

	// Check that the tick index is between min and max value for the given k
	minTick, maxTick := GetMinAndMaxTicksFromK(kAtPriceOne, kIncrementDistance)
	if tickIndex.LT(minTick) {
		return sdk.Dec{}, fmt.Errorf("tickIndex must be greater than or equal to %s", minTick)
	}
	if tickIndex.GT(maxTick) {
		return sdk.Dec{}, fmt.Errorf("tickIndex must be less than or equal to %s", maxTick)
	}

	// Use floor division to determine how many k increments we have passed
	kDelta := tickIndex.ToDec().Quo(kIncrementDistance).TruncateInt()

	// Calculate the current k value from the starting k value and the k delta
	curK := kAtPriceOne.Add(kDelta)

	// Knowing what our curK is, we can then figure out what power of 10 this k corresponds to
	curIncrement := handleNegativeExponents(curK)

	// Now, starting at the minimum tick of the current increment, we calculate how many ticks in the current k we have passed
	numAdditiveTicks := tickIndex.ToDec().Sub(kDelta.ToDec().Mul(kIncrementDistance))

	// Finally, we can calculate the price
	price = handleNegativeExponents(kDelta).Add(numAdditiveTicks.Mul(curIncrement))

	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec, kAtPriceOne sdk.Int) (sdk.Int, error) {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt(), nil
	}

	if price.LTE(sdk.ZeroDec()) {
		return sdk.ZeroInt(), fmt.Errorf("price must be greater than zero")
	}

	if kAtPriceOne.LT(types.PrecisionValueAtPriceOneMin) || kAtPriceOne.GT(types.PrecisionValueAtPriceOneMax) {
		return sdk.Int{}, fmt.Errorf("kAtPriceOne must be in the range (%s, %s)", types.PrecisionValueAtPriceOneMin, types.PrecisionValueAtPriceOneMax)
	}

	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	kIncrementDistance := sdkNineDec.Mul(handleNegativeExponents(kAtPriceOne.Neg()))

	// If the price is less than 1, we must reduce the increment distance by a factor of 10
	if price.LT(sdk.OneDec()) {
		kIncrementDistance = kIncrementDistance.Quo(sdkTenDec)
	}

	// Initialize the total price to 1, the current k to k_at_price_1, and the number of ticks passed to 0
	totalPrice := sdk.OneDec()
	ticksPassed := sdk.ZeroInt()
	currentK := kAtPriceOne

	// Set the currentIncrement to the kAtPriceOne
	curIncrement := handleNegativeExponents(kAtPriceOne)

	// Now, we loop through the k increments until we have passed the price
	// Once we pass the price, we can determine what which k values we have filled in their entirety,
	// as well as how many ticks that corresponds to
	// In the opposite direction (price < 1), we do the same thing (just decrement k instead of incrementing).
	// The only difference is we must reduce the increment distance by a factor of 10.
	if price.GT(sdk.OneDec()) {
		for totalPrice.LT(price) {
			curIncrement = handleNegativeExponents(currentK)
			maxPriceForCurrentIncrement := kIncrementDistance.Mul(curIncrement)
			totalPrice = totalPrice.Add(maxPriceForCurrentIncrement)
			currentK = currentK.Add(sdk.OneInt())
			ticksPassed = ticksPassed.Add(kIncrementDistance.TruncateInt())
		}
	} else {
		for totalPrice.GT(price) {
			curIncrement = handleNegativeExponents(currentK)
			maxPriceForCurrentIncrement := kIncrementDistance.Mul(curIncrement)
			totalPrice = totalPrice.Sub(maxPriceForCurrentIncrement)
			currentK = currentK.Sub(sdk.OneInt())
			ticksPassed = ticksPassed.Sub(kIncrementDistance.TruncateInt())
		}
	}
	// Determine how many ticks we have passed in the current k increment
	ticksToBeFulfilledByCurrentK := price.Sub(totalPrice).Quo(curIncrement)

	// Finally, add the ticks we have passed from the completed k values, as well as the ticks we have passed in the current k value
	tickIndex := ticksPassed.Add(ticksToBeFulfilledByCurrentK.TruncateInt())

	// Add a check to make sure that the tick index is within the allowed range
	minTick, maxTick := GetMinAndMaxTicksFromK(kAtPriceOne, kIncrementDistance)
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
func handleNegativeExponents(exponent sdk.Int) sdk.Dec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdkTenDec.Power(exponent.Uint64())
	}
	return sdk.OneDec().Quo(sdkTenDec.Power(exponent.Abs().Uint64()))
}

// GetMinAndMaxTicksFromK determines min and max ticks allowed for a given k value
// This allows for a min spot price of 0.000000000000000001 and a max spot price of 200000000000 for every k value
func GetMinAndMaxTicksFromK(kAtPriceOne sdk.Int, kIncrementDistance sdk.Dec) (minTick, maxTick sdk.Int) {
	minTick = sdk.NewDec(18).Mul(kIncrementDistance).Neg().RoundInt()
	maxTick = handleNegativeExponents(kAtPriceOne.Neg().Add(sdk.NewInt(2))).RoundInt()
	return minTick, maxTick
}
