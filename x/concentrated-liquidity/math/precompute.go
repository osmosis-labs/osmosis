package math

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

var (
	sdkOneDec      = osmomath.OneDec()
	sdkTenDec      = osmomath.NewDec(10)
	powersOfTen    []osmomath.Dec
	negPowersOfTen []osmomath.Dec

	osmomathBigOneDec = osmomath.NewBigDec(1)
	osmomathBigTenDec = osmomath.NewBigDec(10)
	bigPowersOfTen    []osmomath.BigDec
	bigNegPowersOfTen []osmomath.BigDec

	// 9 * 10^(-types.ExponentAtPriceOne), where types.ExponentAtPriceOne is non-positive and is s.t.
	// this answer fits well within an int64.
	geometricExponentIncrementDistanceInTicks = 9 * osmomath.NewDec(10).PowerMut(uint64(-types.ExponentAtPriceOne)).TruncateInt64()
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
	initialPrice osmomath.BigDec
	// if price >= maxPrice, we are not in this exponent range.
	maxPrice osmomath.BigDec
	// TODO: Change to normal Dec, if min spot price increases.
	// additive increment per tick here.
	additiveIncrementPerTick osmomath.BigDec
	// the tick that corresponds to initial price
	initialTick int64
}

var tickExpCache map[int64]*tickExpIndexData = make(map[int64]*tickExpIndexData)

func buildTickExpCache() {
	// build positive indices first
	maxPrice := osmomathBigOneDec
	curExpIndex := int64(0)
	for maxPrice.LT(osmomath.BigDecFromDec(types.MaxSpotPrice)) {
		tickExpCache[curExpIndex] = &tickExpIndexData{
			// price range 10^curExpIndex to 10^(curExpIndex + 1). (10, 100)
			initialPrice:             osmomathBigTenDec.PowerInteger(uint64(curExpIndex)),
			maxPrice:                 osmomathBigTenDec.PowerInteger(uint64(curExpIndex + 1)),
			additiveIncrementPerTick: powTenBigDec(types.ExponentAtPriceOne + curExpIndex),
			initialTick:              geometricExponentIncrementDistanceInTicks * curExpIndex,
		}
		maxPrice = tickExpCache[curExpIndex].maxPrice
		curExpIndex += 1
	}

	minPrice := osmomathBigOneDec
	curExpIndex = -1
	for minPrice.GT(osmomath.NewBigDecWithPrec(1, 30)) {
		tickExpCache[curExpIndex] = &tickExpIndexData{
			// price range 10^curExpIndex to 10^(curExpIndex + 1). (0.001, 0.01)
			initialPrice:             powTenBigDec(curExpIndex),
			maxPrice:                 powTenBigDec(curExpIndex + 1),
			additiveIncrementPerTick: powTenBigDec(types.ExponentAtPriceOne + curExpIndex),
			initialTick:              geometricExponentIncrementDistanceInTicks * curExpIndex,
		}
		minPrice = tickExpCache[curExpIndex].initialPrice
		curExpIndex -= 1
	}
}

// Set precision multipliers
func init() {
	negPowersOfTen = make([]osmomath.Dec, osmomath.DecPrecision+1)
	for i := 0; i <= osmomath.DecPrecision; i++ {
		negPowersOfTen[i] = sdkOneDec.Quo(sdkTenDec.Power(uint64(i)))
	}
	// 10^77 < osmomath.MaxInt < 10^78
	powersOfTen = make([]osmomath.Dec, 77)
	for i := 0; i <= 76; i++ {
		powersOfTen[i] = sdkTenDec.Power(uint64(i))
	}

	bigNegPowersOfTen = make([]osmomath.BigDec, osmomath.BigDecPrecision+1)
	for i := 0; i <= osmomath.BigDecPrecision; i++ {
		bigNegPowersOfTen[i] = osmomathBigOneDec.Quo(osmomathBigTenDec.PowerInteger(uint64(i)))
	}
	// 10^308 < osmomath.MaxInt < 10^309
	bigPowersOfTen = make([]osmomath.BigDec, 309)
	for i := 0; i <= 308; i++ {
		bigPowersOfTen[i] = osmomathBigTenDec.PowerInteger(uint64(i))
	}

	buildTickExpCache()
}
