package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	kIncrementDistance := sdk.NewDec(9).Mul(handleNegativeExponents(kAtPriceOne.Neg()))

	// Use floor division to determine how many k increments we have passed
	kDelta := tickIndex.ToDec().Quo(kIncrementDistance).TruncateInt()

	// Calculate the current k value from the starting k value and the k delta
	curK := kAtPriceOne.Add(kDelta)

	curIncrement := handleNegativeExponents(curK)

	numAdditiveTicks := tickIndex.ToDec().Sub(kDelta.ToDec().Mul(kIncrementDistance))

	price = handleNegativeExponents(kDelta).Add(numAdditiveTicks.Mul(curIncrement))

	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec, kAtPriceOne sdk.Int) (tickIndex sdk.Int) {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt()
	}

	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	kIncrementDistance := sdk.NewDec(9).Mul(handleNegativeExponents(kAtPriceOne.Neg()))

	total := sdk.OneDec()
	ticksPassed := sdk.ZeroInt()
	currentK := kAtPriceOne

	curIncrement := handleNegativeExponents(currentK)

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
	ticksToBeFulfilledByCurrentK := price.Sub(total).Quo(curIncrement)

	ticksPassed = ticksPassed.Add(ticksToBeFulfilledByCurrentK.TruncateInt())

	return ticksPassed
}

// handleNegativeExponents treats negative exponents as 1/(10**|exponent|) instead of 10**-exponent
// This is because the sdk.Dec.Power function does not support negative exponents
func handleNegativeExponents(exponent sdk.Int) sdk.Dec {
	if exponent.GTE(sdk.ZeroInt()) {
		return sdk.NewDec(10).Power(exponent.Uint64())
	}
	return sdk.OneDec().Quo(sdk.NewDec(10).Power(exponent.Abs().Uint64()))
}
