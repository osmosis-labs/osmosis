package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	kIncrementDistance := sdkNineDec.Mul(handleNegativeExponents(kAtPriceOne.Neg()))

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
func PriceToTick(price sdk.Dec, kAtPriceOne sdk.Int) sdk.Int {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt()
	}

	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	kIncrementDistance := sdkNineDec.Mul(handleNegativeExponents(kAtPriceOne.Neg()))

	// Initialize the total price to 1, the current k to k_at_price_1, and the number of ticks passed to 0
	total := sdk.OneDec()
	ticksPassed := sdk.ZeroInt()
	currentK := kAtPriceOne

	// Set the currentIncrement to the kAtPriceOne
	curIncrement := handleNegativeExponents(currentK)

	// Now, we loop through the k increments until we have passed the price
	// Once we pass the price, we can determine what which k values we have filled in their entirety,
	// as well as how many ticks that corresponds to
	for total.LT(price) {
		curIncrement = handleNegativeExponents(currentK)
		maxPriceForCurrentIncrement := kIncrementDistance.Mul(curIncrement)
		if total.Add(maxPriceForCurrentIncrement).LT(price) {
			total = total.Add(maxPriceForCurrentIncrement)
			currentK = currentK.Add(sdk.OneInt())
			ticksPassed = ticksPassed.Add(kIncrementDistance.TruncateInt())
		} else {
			break
		}
	}
	// Determine how many ticks we have passed in the current k increment
	ticksToBeFulfilledByCurrentK := price.Sub(total).Quo(curIncrement)

	// Finally, add the ticks we have passed from the completed k values, as well as the ticks we have passed in the current k value
	tickIndex := ticksPassed.Add(ticksToBeFulfilledByCurrentK.TruncateInt())

	return tickIndex
}

// handleNegativeExponents treats negative exponents as 1/(10**|exponent|) instead of 10**-exponent
// This is because the sdk.Dec.Power function does not support negative exponents
func handleNegativeExponents(exponent sdk.Int) sdk.Dec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdkTenDec.Power(exponent.Uint64())
	}
	return sdk.OneDec().Quo(sdkTenDec.Power(exponent.Abs().Uint64()))
}
