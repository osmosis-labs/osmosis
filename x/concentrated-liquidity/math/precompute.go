package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

var (
	sdkOneDec      = sdk.OneDec()
	sdkNineDec     = sdk.NewDec(9)
	sdkTenDec      = sdk.NewDec(10)
	powersOfTen    []sdk.Dec
	negPowersOfTen []sdk.Dec

	osmomathBigOneDec = osmomath.NewBigDec(1)
	osmomathBigTenDec = osmomath.NewBigDec(10)
	bigPowersOfTen    []osmomath.BigDec
	bigNegPowersOfTen []osmomath.BigDec

	// 9 * 10^(-types.ExponentAtPriceOne), where types.ExponentAtPriceOne is non-positive and is s.t.
	// this answer fits well within an int64.
	geometricExponentIncrementDistanceInTicks = 9 * sdk.NewDec(10).PowerMut(uint64(-types.ExponentAtPriceOne)).TruncateInt64()
)

// Builds metadata for every additive tickspacing exponent, namely:
// * what is the first price this tick spacing applies to
// * what is the first tick this applies for
// * (saves on pre-compute) what is the additive increment per tick.
//
// This would be stored in a map, keyed by:
// 0 => (1.00, 10^(types.ExponentAtPriceOne), 0)
// 1 => (10, 10^(1 + types.ExponentAtPriceOne), 9 * types.ExponentAtPriceOne)
// 2 => (100, 10^(2 + types.ExponentAtPriceOne), 9 * (types.ExponentAtPriceOne + 1))
// -1 => (0.1, 10^(types.ExponentAtPriceOne - 1), 9 * (types.ExponentAtPriceOne - 1))
type tickExpIndexData struct {
	// if price < initialPrice, we are not in this exponent range.
	initialPrice sdk.Dec
	// if price >= maxPrice, we are not in this exponent range.
	maxPrice sdk.Dec
	// TODO: Change to normal Dec, if min spot price increases.
	// additive increment per tick here.
	additiveIncrementPerTick osmomath.BigDec
	// the tick that corresponds to initial price
	initialTick int64
}

var tickExpCache map[int64]*tickExpIndexData = make(map[int64]*tickExpIndexData)

func buildTickExpCache() {
	// build positive indices first
	maxPrice := sdkOneDec
	curExpIndex := int64(0)
	for maxPrice.LT(types.MaxSpotPrice) {
		tickExpCache[curExpIndex] = &tickExpIndexData{
			// price range 10^curExpIndex to 10^(curExpIndex + 1). (10, 100)
			initialPrice:             sdkTenDec.Power(uint64(curExpIndex)),
			maxPrice:                 sdkTenDec.Power(uint64(curExpIndex + 1)),
			additiveIncrementPerTick: powTenBigDec(types.ExponentAtPriceOne + curExpIndex),
			initialTick:              geometricExponentIncrementDistanceInTicks * curExpIndex,
		}
		maxPrice = tickExpCache[curExpIndex].maxPrice
		curExpIndex += 1
	}

	minPrice := sdkOneDec
	curExpIndex = -1
	for minPrice.GT(types.MinSpotPrice) {
		tickExpCache[curExpIndex] = &tickExpIndexData{
			// price range 10^curExpIndex to 10^(curExpIndex + 1). (0.001, 0.01)
			initialPrice:             powTenBigDec(curExpIndex).SDKDec(),
			maxPrice:                 powTenBigDec(curExpIndex + 1).SDKDec(),
			additiveIncrementPerTick: powTenBigDec(types.ExponentAtPriceOne + curExpIndex),
			initialTick:              geometricExponentIncrementDistanceInTicks * curExpIndex,
		}
		minPrice = tickExpCache[curExpIndex].initialPrice
		curExpIndex -= 1
	}
}

// Set precision multipliers
func init() {
	negPowersOfTen = make([]sdk.Dec, sdk.Precision+1)
	for i := 0; i <= sdk.Precision; i++ {
		negPowersOfTen[i] = sdkOneDec.Quo(sdkTenDec.Power(uint64(i)))
	}
	// 10^77 < sdk.MaxInt < 10^78
	powersOfTen = make([]sdk.Dec, 78)
	for i := 0; i <= 77; i++ {
		powersOfTen[i] = sdkTenDec.Power(uint64(i))
	}

	bigNegPowersOfTen = make([]osmomath.BigDec, osmomath.Precision+1)
	for i := 0; i <= osmomath.Precision; i++ {
		bigNegPowersOfTen[i] = osmomathBigOneDec.Quo(osmomathBigTenDec.PowerInteger(uint64(i)))
	}
	// 10^308 < osmomath.MaxInt < 10^309
	bigPowersOfTen = make([]osmomath.BigDec, 309)
	for i := 0; i <= 308; i++ {
		bigPowersOfTen[i] = osmomathBigTenDec.PowerInteger(uint64(i))
	}

	buildTickExpCache()
}
