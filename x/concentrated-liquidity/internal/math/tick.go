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

// TickToSqrtPrice calculates the price at a given tick index based on the provided
// starting price of 1 at k=0. The price is calculated using a square root function
// with a coefficient of 9, where the price increases by a factor of 10 for every
// increment of k.
//
// The function takes in two arguments:
// 	- tickIndex: the tick index to calculate the price for
// 	- kAtPriceOne: the value of k at which the starting price of 1 is set
//
// It returns a sdk.Dec representing the calculated price and an error if any errors
// occurred during the calculation.
func TickToPrice(tickIndex, kAtPriceOne sdk.Int) (price sdk.Dec, err error) {
	if tickIndex.IsZero() {
		return sdk.OneDec(), nil
	}

	var kIncrementDistance sdk.Dec
	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	if kAtPriceOne.GTE(sdk.ZeroInt()) {
		kIncrementDistance = sdk.NewDec(9).Mul(sdk.OneDec().Quo(sdk.NewInt(10).ToDec().Power(kAtPriceOne.Uint64())))
	} else {
		kIncrementDistance = sdk.NewDec(9).Mul(sdk.NewInt(10).ToDec().Power(kAtPriceOne.Abs().Uint64()))
	}

	// Use floor division to determine how many k increments we have passed
	kDelta := tickIndex.ToDec().Quo(kIncrementDistance).TruncateInt()

	// Calculate the current k value from the starting k value and the k delta
	curK := kAtPriceOne.Add(kDelta)

	var curIncrement sdk.Dec
	if curK.GTE(sdk.ZeroInt()) {
		curIncrement = sdk.NewDec(10).Power(curK.Uint64())
	} else {
		curIncrement = sdk.NewDec(1).Quo(sdk.NewDec(10).Power(curK.Abs().Uint64()))
	}

	numAdditiveTicks := tickIndex.ToDec().Sub(kDelta.ToDec().Mul(kIncrementDistance))

	if kDelta.GTE(sdk.ZeroInt()) {
		price = sdk.NewDec(10).Power(kDelta.Uint64()).Add(numAdditiveTicks.Mul(curIncrement))
	} else {
		price = sdk.OneDec().Quo(sdk.NewDec(10).Power(kDelta.Abs().Uint64())).Add(numAdditiveTicks.Mul(curIncrement))
	}
	return price, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec, kAtPriceOne sdk.Int) (tickIndex sdk.Int) {
	if price.Equal(sdk.OneDec()) {
		return sdk.ZeroInt()
	}

	var kIncrementDistance sdk.Dec
	// The formula is as follows: k_increment_distance = 9 * 10**(-k_at_price_1)
	// Due to sdk.Power restrictions, if the resulting power is negative, we take 9 * (1/10**k_at_price_1)
	if kAtPriceOne.GTE(sdk.ZeroInt()) {
		kIncrementDistance = sdk.NewDec(9).Mul(sdk.OneDec().Quo(sdk.NewInt(10).ToDec().Power(kAtPriceOne.Uint64())))
	} else {
		kIncrementDistance = sdk.NewDec(9).Mul(sdk.NewInt(10).ToDec().Power(kAtPriceOne.Abs().Uint64()))
	}

	total := sdk.OneDec()
	ticksPassed := sdk.ZeroInt()
	currentK := kAtPriceOne

	var curIncrement sdk.Dec
	if currentK.GTE(sdk.ZeroInt()) {
		curIncrement = sdk.NewDec(10).Power(currentK.Uint64())
	} else {
		curIncrement = sdk.NewDec(1).Quo(sdk.NewDec(10).Power(currentK.Abs().Uint64()))
	}

	for total.LT(price) {
		if currentK.GTE(sdk.ZeroInt()) {
			curIncrement = sdk.NewDec(10).Power(currentK.Uint64())
			maxPriceForCurrentIncrement := kIncrementDistance.Mul(curIncrement)
			if total.Add(maxPriceForCurrentIncrement).LT(price) {
				total = total.Add(maxPriceForCurrentIncrement)
				currentK = currentK.Add(sdk.OneInt())
				ticksPassed = ticksPassed.Add(kIncrementDistance.TruncateInt())
			} else {
				break
			}
		} else {
			curIncrement = sdk.NewDec(1).Quo(sdk.NewDec(10).Power(currentK.Abs().Uint64()))
			maxPriceForCurrentIncrement := kIncrementDistance.Mul(curIncrement)
			if total.Add(maxPriceForCurrentIncrement).LT(price) {
				total = total.Add(maxPriceForCurrentIncrement)
				currentK = currentK.Add(sdk.OneInt())
				ticksPassed = ticksPassed.Add(kIncrementDistance.TruncateInt())
			} else {
				break
			}
		}
	}
	ticksToBeFulfilledByCurrentK := price.Sub(total).Quo(curIncrement)

	ticksPassed = ticksPassed.Add(ticksToBeFulfilledByCurrentK.TruncateInt())

	return ticksPassed
}
