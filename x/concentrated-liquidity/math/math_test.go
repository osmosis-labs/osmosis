package math_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

var (
	sqrt4545 = osmomath.MustNewDecFromStr("67.416615162732695594")
	sqrt5000 = osmomath.MustNewDecFromStr("70.710678118654752440")
	sqrt5500 = osmomath.MustNewDecFromStr("74.161984870956629487")

	sqrt4545BigDec = osmomath.BigDecFromDec(sqrt4545)
	sqrt5000BigDec = osmomath.BigDecFromDec(sqrt5000)
	sqrt5500BigDec = osmomath.BigDecFromDec(sqrt5500)

	// sqrt(10 ^-36 * 567) 36 decimals
	// chosen arbitrarily for testing the extended price range
	sqrtANearMin = osmomath.MustNewBigDecFromStr("0.000000000000000023811761799581315315")
	// sqrt(10 ^-36 * 123567) 36 decimals
	// chosen arbitrarily for testing the extended price range
	sqrtBNearMin = osmomath.MustNewBigDecFromStr("0.000000000000000351520980881653776714")
	// This value is estimated using liquidity1 function in clmath.py between sqrtANearMin and sqrtBNearMin.
	smallLiquidity = osmomath.MustNewBigDecFromStr("0.000000000000316705045072312223884779")
	// Arbitrary small value, exists to test small movement over the low price range.
	smallValue = osmomath.MustNewBigDecFromStr("10.12345678912345671234567891234567")
)

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity1 = amount1 / (sqrtPriceB - sqrtPriceA)
func TestLiquidity1(t *testing.T) {
	testCases := map[string]struct {
		currentSqrtP      osmomath.BigDec
		sqrtPLow          osmomath.BigDec
		amount1Desired    osmomath.Int
		expectedLiquidity string
	}{
		"happy path": {
			currentSqrtP:      sqrt5000BigDec, // 5000
			sqrtPLow:          sqrt4545BigDec, // 4545
			amount1Desired:    osmomath.NewInt(5000000000),
			expectedLiquidity: "1517882343.751510418088349649",
			// https://www.wolframalpha.com/input?i=5000000000+%2F+%2870.710678118654752440+-+67.416615162732695594%29
		},
		"low price range": {
			currentSqrtP:   sqrtANearMin,
			sqrtPLow:       sqrtBNearMin,
			amount1Desired: osmomath.NewInt(5000000000),
			// from math import *
			// from decimal import *
			// amount1 / (sqrtPriceB - sqrtPriceA)
			expectedLiquidity: "15257428564277849269508363.222206252646611708",
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			liquidity := math.Liquidity1(tc.amount1Desired, tc.currentSqrtP, tc.sqrtPLow)
			require.Equal(t, tc.expectedLiquidity, liquidity.String())
		})
	}
}

// TestLiquidity0 tests that liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity0 = amount0 * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)
func TestLiquidity0(t *testing.T) {
	testCases := map[string]struct {
		currentSqrtP      osmomath.BigDec
		sqrtPHigh         osmomath.BigDec
		amount0Desired    osmomath.Int
		expectedLiquidity string
	}{
		"happy path": {
			currentSqrtP:      sqrt5000BigDec, // 5000
			sqrtPHigh:         sqrt5500BigDec, // 5500
			amount0Desired:    osmomath.NewInt(1000000),
			expectedLiquidity: "1519437308.014768571720923239",
			// https://www.wolframalpha.com/input?i=1000000+*+%2870.710678118654752440*+74.161984870956629487%29+%2F+%2874.161984870956629487+-+70.710678118654752440%29
		},
		"sqrtPriceA greater than sqrtPriceB": {
			currentSqrtP:      sqrt5500BigDec, // 5000
			sqrtPHigh:         sqrt5000BigDec,
			amount0Desired:    osmomath.NewInt(1000000),
			expectedLiquidity: "1519437308.014768571720923239",
			// https://www.wolframalpha.com/input?i=1000000+*+%2870.710678118654752440*+74.161984870956629487%29+%2F+%2874.161984870956629487+-+70.710678118654752440%29
		},
		"low price range": {
			currentSqrtP:   sqrtANearMin,
			sqrtPHigh:      sqrtBNearMin,
			amount0Desired: osmomath.NewInt(123999),
			// from clmath import *
			// from math import *
			// product1 = round_osmo_prec_down(sqrtPriceA * sqrtPriceB)
			// product2 = round_osmo_prec_down(amount0 * product1)
			// diff = round_osmo_prec_down(sqrtPriceB - sqrtPriceA)
			// round_sdk_prec_down(product2 / diff)
			expectedLiquidity: "0.000000000003167050",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			liquidity := math.Liquidity0(tc.amount0Desired, tc.currentSqrtP, tc.sqrtPHigh)
			require.Equal(t, tc.expectedLiquidity, liquidity.String())
		})
	}
}

// TestCalcAmount0Delta tests that calcAmount0 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount0Delta = (liquidity * (sqrtPriceB - sqrtPriceA)) / (sqrtPriceB * sqrtPriceA)
func TestCalcAmount0Delta(t *testing.T) {
	testCases := map[string]struct {
		liquidity       osmomath.Dec
		sqrtPA          osmomath.BigDec
		sqrtPB          osmomath.BigDec
		isWithTolerance bool
		roundUp         bool
		amount0Expected osmomath.BigDec
	}{
		"happy path": {
			liquidity: osmomath.MustNewDecFromStr("1517882343.751510418088349649"), // we use the smaller liquidity between liq0 and liq1
			sqrtPA:    sqrt5000BigDec,                                              // 5000
			sqrtPB:    sqrt5500BigDec,                                              // 5500
			roundUp:   false,
			// calculated with x/concentrated-liquidity/python/clmath.py  round_decimal(amount0, 36, ROUND_FLOOR)
			amount0Expected: osmomath.MustNewBigDecFromStr("998976.618347426388356629926969277767437533"), // truncated at precision end.
			isWithTolerance: false,
		},
		"happy path, sqrtPriceA greater than sqrtPrice B": { // commute prior vector
			liquidity: osmomath.MustNewDecFromStr("1517882343.751510418088349649"),
			sqrtPA:    sqrt5500BigDec,
			sqrtPB:    sqrt5000BigDec,
			roundUp:   false,
			// calculated with x/concentrated-liquidity/python/clmath.py  round_decimal(amount0, 36, ROUND_FLOOR)
			amount0Expected: osmomath.MustNewBigDecFromStr("998976.618347426388356629926969277767437533"),
			isWithTolerance: false,
		},
		"round down: large liquidity amount in wide price range": {
			// Note the values are hand-picked to cause multiplication of 2 large numbers
			// causing the magnitude of truncations to be larger
			// while staying under bit length of osmomath.Dec
			// from decimal import *
			// from math import *
			// getcontext().prec = 100
			// max_sqrt_p = Decimal("30860351331.852813530648276680")
			// min_sqrt_p = Decimal("0.000000152731791058")
			// liq = Decimal("931361973132462178951297")
			// liq * (max_sqrt_p - min_sqrt_p) / (max_sqrt_p * min_sqrt_p)
			liquidity: osmomath.MustNewDecFromStr("931361973132462178951297"),
			// price: 0.000000000000023327
			sqrtPA: osmomath.MustNewBigDecFromStr("0.000000152731791058"),
			// price: 952361284325389721913
			sqrtPB:  osmomath.MustNewBigDecFromStr("30860351331.852813530648276680"),
			roundUp: false,
			// calculated with x/concentrated-liquidity/python/clmath.py
			amount0Expected: osmomath.MustNewBigDecFromStr("6098022989717817431593106314408.88812810159039320984467945943"), // truncated at precision end.
			isWithTolerance: true,
		},
		"round up: large liquidity amount in wide price range": {
			// Note the values are hand-picked to cause multiplication of 2 large numbers
			// causing the magnitude of truncations to be larger
			// while staying under bit length of osmomath.Dec
			// from decimal import *
			// from math import *
			// getcontext().prec = 100
			// max_sqrt_p = Decimal("30860351331.852813530648276680")
			// min_sqrt_p = Decimal("0.000000152731791058")
			// liq = Decimal("931361973132462178951297")
			// liq * (max_sqrt_p - min_sqrt_p) / (max_sqrt_p * min_sqrt_p)
			liquidity: osmomath.MustNewDecFromStr("931361973132462178951297"),
			// price: 0.000000000000023327
			sqrtPA: osmomath.MustNewBigDecFromStr("0.000000152731791058"),
			// price: 952361284325389721913
			sqrtPB:          osmomath.MustNewBigDecFromStr("30860351331.852813530648276680"),
			roundUp:         true,
			amount0Expected: osmomath.MustNewBigDecFromStr("6098022989717817431593106314408.88812810159039320984467945943").Ceil(), // rounded up at precision end.
			isWithTolerance: true,
		},
		// See: https://github.com/osmosis-labs/osmosis/issues/6351 for details.
		// The values were taken from the failing swap on the development branch that reproduced the issue.
		"edge case: low sqrt prices may cause error amplification if incorrect order of operations (round up)": {
			// from decimal import *
			// from math import *
			// getcontext().prec = 100
			// max_sqrt_p = Decimal("0.00000099994999874993749609347654199")
			// min_sqrt_p = Decimal("0.000000000000001409841835100661211756")
			// liq = Decimal("5000252259822539816806336.971796256914465071095518135400579243")
			// liq * (max_sqrt_p - min_sqrt_p) / (max_sqrt_p * min_sqrt_p)
			liquidity:       osmomath.MustNewDecFromStr("5000252259822539816806336.971796256914465071"),
			sqrtPA:          osmomath.MustNewBigDecFromStr("0.00000099994999874993749609347654199"),
			sqrtPB:          osmomath.MustNewBigDecFromStr("0.000000000000001409841835100661211756"),
			roundUp:         true,
			amount0Expected: osmomath.MustNewBigDecFromStr("3546676037185128488234786333758360815266.999539026068480181194797910898392880").Ceil(),
		},
		// See: https://github.com/osmosis-labs/osmosis/issues/6351 for details.
		// The values were taken from the failing swap on the development branch that reproduced the issue.
		"edge case: low sqrt prices may cause error amplification if incorrect order of operations (round down)": {
			// from decimal import *
			// from math import *
			// getcontext().prec = 100
			// max_sqrt_p = Decimal("0.00000099994999874993749609347654199")
			// min_sqrt_p = Decimal("0.000000000000001409841835100661211756")
			// liq = Decimal("5000252259822539816806336.971796256914465071095518135400579243")
			// liq * (max_sqrt_p - min_sqrt_p) / (max_sqrt_p * min_sqrt_p)
			liquidity:       osmomath.MustNewDecFromStr("5000252259822539816806336.971796256914465071"),
			sqrtPA:          osmomath.MustNewBigDecFromStr("0.00000099994999874993749609347654199"),
			sqrtPB:          osmomath.MustNewBigDecFromStr("0.000000000000001409841835100661211756"),
			roundUp:         false,
			isWithTolerance: true,
			amount0Expected: osmomath.MustNewBigDecFromStr("3546676037185128488234786333758360815266.999539026068480181194797910898392880"),
		},
		"low price range": {
			liquidity: smallLiquidity.Dec(),
			sqrtPA:    sqrtANearMin,
			sqrtPB:    sqrtBNearMin,
			roundUp:   false,
			// from clmath decimal import *
			// from math import *
			// calc_amount_zero_delta(liq, sqrtPriceA, sqrtPriceB, False)
			amount0Expected: osmomath.MustNewBigDecFromStr("12399.403617882634341191547243098659145924"),
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			amount0 := math.CalcAmount0Delta(tc.liquidity, tc.sqrtPA, tc.sqrtPB, tc.roundUp)

			if !tc.isWithTolerance {
				require.Equal(t, tc.amount0Expected, amount0)
				return
			}

			roundingDir := osmomath.RoundUp
			if !tc.roundUp {
				roundingDir = osmomath.RoundDown
			}

			tolerance := osmomath.ErrTolerance{
				MultiplicativeTolerance: osmomath.SmallestDec(),
				RoundingDir:             roundingDir,
			}

			osmoassert.Equal(t, tolerance, tc.amount0Expected, amount0)
		})
	}
}

// TestCalcAmount1Delta tests that calcAmount1 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount1Delta = liq * (sqrtPriceB - sqrtPriceA)
func TestCalcAmount1Delta(t *testing.T) {
	testCases := map[string]struct {
		liquidity       osmomath.Dec
		sqrtPA          osmomath.BigDec
		sqrtPB          osmomath.BigDec
		exactEqual      bool
		roundUp         bool
		amount1Expected osmomath.BigDec
	}{
		"round down": {
			liquidity: osmomath.MustNewDecFromStr("1517882343.751510418088349649"), // we use the smaller liquidity between liq0 and liq1
			sqrtPA:    sqrt5000BigDec,                                              // 5000
			sqrtPB:    sqrt4545BigDec,                                              // 4545
			roundUp:   false,
			// calculated with x/concentrated-liquidity/python/clmath.py
			amount1Expected: osmomath.MustNewBigDecFromStr("4999999999.999999999999999999696837821702147054"),
		},
		"round down: large liquidity amount in wide price range": {
			// Note the values are hand-picked to cause multiplication of 2 large numbers
			// while staying under bit length of osmomath.Dec
			// from decimal import *
			// from math import *
			// getcontext().prec = 100
			// max_sqrt_p = Decimal("30860351331.852813530648276680")
			// min_sqrt_p = Decimal("0.000000152731791058")
			// liq = Decimal("931361973132462178951297")
			// liq * (max_sqrt_p - min_sqrt_p)
			liquidity: osmomath.MustNewDecFromStr("931361973132462178951297"),
			// price: 0.000000000000023327
			sqrtPA: osmomath.MustNewBigDecFromStr("0.000000152731791058"),
			// price: 952361284325389721913
			sqrtPB:  osmomath.MustNewBigDecFromStr("30860351331.852813530648276680"),
			roundUp: false,
			// calculated with x/concentrated-liquidity/python/clmath.py
			amount1Expected: osmomath.MustNewBigDecFromStr("28742157707995443393876876754535992.801567623738751734"), // truncated at precision end.
		},
		"round up: large liquidity amount in wide price range": {
			// Note the values are hand-picked to cause multiplication of 2 large numbers
			// while staying under bit length of osmomath.Dec
			// from decimal import *
			// from math import *
			// getcontext().prec = 100
			// max_sqrt_p = Decimal("30860351331.852813530648276680")
			// min_sqrt_p = Decimal("0.000000152731791058")
			// liq = Decimal("931361973132462178951297")
			// liq * (max_sqrt_p - min_sqrt_p)
			liquidity: osmomath.MustNewDecFromStr("931361973132462178951297"),
			// price: 0.000000000000023327
			sqrtPA: osmomath.MustNewBigDecFromStr("0.000000152731791058"),
			// price: 952361284325389721913
			sqrtPB:          osmomath.MustNewBigDecFromStr("30860351331.852813530648276680"),
			roundUp:         true,
			amount1Expected: osmomath.MustNewBigDecFromStr("28742157707995443393876876754535992.801567623738751734").Ceil(), // round up at precision end.
		},
		"low price range (no round up)": {
			liquidity: smallLiquidity.Dec(),
			sqrtPA:    sqrtANearMin,
			sqrtPB:    sqrtBNearMin,
			roundUp:   false,
			// from clmath decimal import *
			// from math import *
			// calc_amount_one_delta(liq, sqrtPriceA, sqrtPriceB, False)
			amount1Expected: osmomath.MustNewBigDecFromStr("0.000000000000000000000000000103787148"),
		},
		"low price range (with round up)": {
			liquidity: smallLiquidity.Dec(),
			sqrtPA:    sqrtANearMin,
			sqrtPB:    sqrtBNearMin,
			roundUp:   true,
			// from clmath decimal import *
			// calc_amount_one_delta(liq, sqrtPriceA, sqrtPriceB, False)
			// Actual result: 0.000000000000000000000000000103787149
			amount1Expected: osmomath.MustNewBigDecFromStr("0.000000000000000000000000000103787149").Ceil(),
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			amount1 := math.CalcAmount1Delta(tc.liquidity, tc.sqrtPA, tc.sqrtPB, tc.roundUp)

			require.Equal(t, tc.amount1Expected, amount1)
		})
	}
}

func TestGetLiquidityFromAmounts(t *testing.T) {
	sqrt := func(x osmomath.Dec) osmomath.BigDec {
		sqrt, err := osmomath.MonotonicSqrt(x)
		require.NoError(t, err)
		return osmomath.BigDecFromDec(sqrt)
	}

	testCases := map[string]struct {
		currentSqrtP osmomath.BigDec
		sqrtPHigh    osmomath.BigDec
		sqrtPLow     osmomath.BigDec
		// the amount of token0 that will need to be sold to move the price from P_cur to P_low
		amount0Desired osmomath.Int
		// the amount of token 1 that will need to be sold to move the price from P_cur to P_high.
		amount1Desired    osmomath.Int
		expectedLiquidity string
		// liq0 = rate of change of reserves of token 1 for a change between sqrt(P_cur) and sqrt(P_low)
		// liq1 = rate of change of reserves of token 1 for a change between sqrt(P_cur) and sqrt(P_high)
		// price of x in terms of y
		expectedLiquidity0 osmomath.Dec
		expectedLiquidity1 osmomath.Dec
	}{
		"happy path (case A)": {
			currentSqrtP:      osmomath.MustNewBigDecFromStr("67"), // 4489
			sqrtPHigh:         sqrt5500BigDec,                      // 5500
			sqrtPLow:          sqrt4545BigDec,                      // 4545
			amount0Desired:    osmomath.NewInt(1000000),
			amount1Desired:    osmomath.ZeroInt(),
			expectedLiquidity: "741212151.448720111852782017",
		},
		"happy path (case A, but with sqrtPriceA greater than sqrtPriceB)": {
			currentSqrtP:      osmomath.MustNewBigDecFromStr("67"), // 4489
			sqrtPHigh:         sqrt4545BigDec,                      // 4545
			sqrtPLow:          sqrt5500BigDec,                      // 5500
			amount0Desired:    osmomath.NewInt(1000000),
			amount1Desired:    osmomath.ZeroInt(),
			expectedLiquidity: "741212151.448720111852782017",
		},
		"happy path (case B)": {
			currentSqrtP:      sqrt5000BigDec, // 5000
			sqrtPHigh:         sqrt5500BigDec, // 5500
			sqrtPLow:          sqrt4545BigDec, // 4545
			amount0Desired:    osmomath.NewInt(1000000),
			amount1Desired:    osmomath.NewInt(5000000000),
			expectedLiquidity: "1517882343.751510418088349649",
		},
		"happy path (case C)": {
			currentSqrtP:      osmomath.MustNewBigDecFromStr("75"), // 5625
			sqrtPHigh:         sqrt5500BigDec,                      // 5500
			sqrtPLow:          sqrt4545BigDec,                      // 4545
			amount0Desired:    osmomath.ZeroInt(),
			amount1Desired:    osmomath.NewInt(5000000000),
			expectedLiquidity: "741249214.836069764856625637",
		},
		"full range, price proportional to amounts, equal liquidities (some rounding error) price of 4": {
			currentSqrtP:   sqrt(osmomath.NewDec(4)),
			sqrtPHigh:      osmomath.BigDecFromDec(types.MaxSqrtPrice),
			sqrtPLow:       osmomath.BigDecFromDec(types.MinSqrtPrice),
			amount0Desired: osmomath.NewInt(4),
			amount1Desired: osmomath.NewInt(16),

			expectedLiquidity:  osmomath.MustNewDecFromStr("8.000000000000000001").String(),
			expectedLiquidity0: osmomath.MustNewDecFromStr("8.000000000000000001"),
			expectedLiquidity1: osmomath.MustNewDecFromStr("8.000000004000000002"),
		},
		"full range, price proportional to amounts, equal liquidities (some rounding error) price of 2": {
			currentSqrtP:   sqrt(osmomath.NewDec(2)),
			sqrtPHigh:      osmomath.BigDecFromDec(types.MaxSqrtPrice),
			sqrtPLow:       osmomath.BigDecFromDec(types.MinSqrtPrice),
			amount0Desired: osmomath.NewInt(1),
			amount1Desired: osmomath.NewInt(2),

			expectedLiquidity:  osmomath.MustNewDecFromStr("1.414213562373095049").String(),
			expectedLiquidity0: osmomath.MustNewDecFromStr("1.414213562373095049"),
			expectedLiquidity1: osmomath.MustNewDecFromStr("1.414213563373095049"),
		},
		"not full range, price proportional to amounts, non equal liquidities": {
			currentSqrtP:   sqrt(osmomath.NewDec(2)),
			sqrtPHigh:      sqrt(osmomath.NewDec(3)),
			sqrtPLow:       sqrt(osmomath.NewDec(1)),
			amount0Desired: osmomath.NewInt(1),
			amount1Desired: osmomath.NewInt(2),

			expectedLiquidity:  osmomath.MustNewDecFromStr("4.828427124746190095").String(),
			expectedLiquidity0: osmomath.MustNewDecFromStr("7.706742302257039729"),
			expectedLiquidity1: osmomath.MustNewDecFromStr("4.828427124746190095"),
		},
		"current sqrt price on upper bound": {
			currentSqrtP:   sqrt5500BigDec,
			sqrtPHigh:      sqrt5500BigDec,
			sqrtPLow:       sqrt4545BigDec,
			amount0Desired: osmomath.ZeroInt(),
			amount1Desired: osmomath.NewInt(1000000),
			// Liquidity1 = amount1 / (sqrtPriceB - sqrtPriceA)
			// https://www.wolframalpha.com/input?i=1000000%2F%2874.161984870956629487-67.416615162732695594%29
			expectedLiquidity: "148249.842967213952971325",
		},
	}

	for name, tc := range testCases {
		tc := tc

		t.Run(name, func(t *testing.T) {
			// CASE A: if the currentSqrtP is less than the sqrtPLow, all the liquidity is in asset0, so GetLiquidityFromAmounts returns the liquidity of asset0
			// CASE B: if the currentSqrtP is less than the sqrtPHigh but greater than sqrtPLow, the liquidity is split between asset0 and asset1,
			// so GetLiquidityFromAmounts returns the smaller liquidity of asset0 and asset1
			// CASE C: if the currentSqrtP is greater than the sqrtPHigh, all the liquidity is in asset1, so GetLiquidityFromAmounts returns the liquidity of asset1
			liquidity := math.GetLiquidityFromAmounts(tc.currentSqrtP, tc.sqrtPLow, tc.sqrtPHigh, tc.amount0Desired, tc.amount1Desired)
			require.Equal(t, tc.expectedLiquidity, liquidity.String())
		})
	}
}

type sqrtRoundingTestCase struct {
	sqrtPriceCurrent osmomath.BigDec
	liquidity        osmomath.BigDec
	amountRemaining  osmomath.BigDec
	expected         osmomath.BigDec
}

type sqrtRoundingDecTestCase struct {
	sqrtPriceCurrent osmomath.BigDec
	liquidity        osmomath.Dec
	amountRemaining  osmomath.BigDec
	expected         osmomath.BigDec
}

type sqrtRoundingAmtDecTestCase struct {
	sqrtPriceCurrent osmomath.BigDec
	liquidity        osmomath.BigDec
	amountRemaining  osmomath.Dec
	expected         osmomath.BigDec
}

func runSqrtRoundingTestCase(
	t *testing.T,
	name string,
	fn func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec,
	cases map[string]sqrtRoundingTestCase,
) {
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sqrtPriceNext := fn(tc.sqrtPriceCurrent, tc.liquidity, tc.amountRemaining)
			require.Equal(t, tc.expected.String(), sqrtPriceNext.String())
		})
	}
}

func runSqrtRoundingDecTestCase(
	t *testing.T,
	name string,
	fn func(osmomath.BigDec, osmomath.Dec, osmomath.BigDec) osmomath.BigDec,
	cases map[string]sqrtRoundingDecTestCase,
) {
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sqrtPriceNext := fn(tc.sqrtPriceCurrent, tc.liquidity, tc.amountRemaining)
			require.Equal(t, tc.expected.String(), sqrtPriceNext.String())
		})
	}
}

func runSqrtRoundingAmtDecTestCase(
	t *testing.T,
	name string,
	fn func(osmomath.BigDec, osmomath.BigDec, osmomath.Dec) osmomath.BigDec,
	cases map[string]sqrtRoundingAmtDecTestCase,
) {
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sqrtPriceNext := fn(tc.sqrtPriceCurrent, tc.liquidity, tc.amountRemaining)
			require.Equal(t, tc.expected.String(), sqrtPriceNext.String())
		})
	}
}

// Estimates are computed with x/concentrated-liquidity/python/clmath.py
func TestGetNextSqrtPriceFromAmount0InRoundingUp(t *testing.T) {
	tests := map[string]sqrtRoundingTestCase{
		"rounded up at precision end": {
			sqrtPriceCurrent: sqrt5000BigDec,
			liquidity:        osmomath.MustNewBigDecFromStr("3035764687.503020836176699298"),
			amountRemaining:  osmomath.MustNewBigDecFromStr("8398"),
			// get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, amountRemaining)
			expected: osmomath.MustNewBigDecFromStr("70.696849053416966148695392456511981401"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("2"),
			liquidity:        osmomath.MustNewBigDecFromStr("10"),
			amountRemaining:  osmomath.MustNewBigDecFromStr("15"),
			// liq * sqrt_cur / (liq + token_in * sqrt_cur) = 0.5
			expected: osmomath.MustNewBigDecFromStr("0.5"),
		},
		"happy path": {
			liquidity:        osmomath.MustNewBigDecFromStr("1517882343.751510418088349649"), // liquidity0 calculated above
			sqrtPriceCurrent: sqrt5000BigDec,
			amountRemaining:  osmomath.NewBigDec(13370),
			// round_osmo_prec_up(liquidity / (round_osmo_prec_down(liquidity / sqrtPriceCurrent) + amountRemaining))
			expected: osmomath.MustNewBigDecFromStr("70.666663910857144331148691821263626767"),
		},
		"low price range": {
			liquidity:        smallLiquidity,
			sqrtPriceCurrent: sqrtANearMin,
			amountRemaining:  smallValue,
			// from clmath decimal import *
			// get_next_sqrt_price_from_amount0_in_round_up(liq, sqrtPriceA, amountRemaining)
			expected: osmomath.MustNewBigDecFromStr("0.000000000000000023793654323441728435"),
		},
	}
	runSqrtRoundingTestCase(t, "TestGetNextSqrtPriceFromAmount0InRoundingUp", math.GetNextSqrtPriceFromAmount0InRoundingUp, tests)
}

// Estimates are computed with x/concentrated-liquidity/python/clmath.py
func TestGetNextSqrtPriceFromAmount0OutRoundingUp(t *testing.T) {
	tests := map[string]sqrtRoundingAmtDecTestCase{
		"rounded up at precision end": {
			sqrtPriceCurrent: sqrt5000BigDec,
			liquidity:        osmomath.MustNewBigDecFromStr("3035764687.503020836176699298"),
			amountRemaining:  osmomath.MustNewDecFromStr("8398"),
			// get_next_sqrt_price_from_amount0_out_round_up(liquidity,sqrtPriceCurrent ,amountRemaining)
			expected: osmomath.MustNewBigDecFromStr("70.724512595179305565323229510645063950"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("2"),
			liquidity:        osmomath.MustNewBigDecFromStr("10"),
			amountRemaining:  osmomath.MustNewDecFromStr("1"),
			// liq * sqrt_cur / (liq + token_out * sqrt_cur) = 2.5
			expected: osmomath.MustNewBigDecFromStr("2.5"),
		},
		"low price range": {
			liquidity:        smallLiquidity,
			sqrtPriceCurrent: sqrtANearMin,
			amountRemaining:  smallValue.Dec(),
			// from clmath decimal import *
			// get_next_sqrt_price_from_amount0_out_round_up(liq, sqrtPriceA, amountRemaining)
			expected: osmomath.MustNewBigDecFromStr("0.000000000000000023829902587267894423"),
		},
	}
	runSqrtRoundingAmtDecTestCase(t, "TestGetNextSqrtPriceFromAmount0OutRoundingUp", math.GetNextSqrtPriceFromAmount0OutRoundingUp, tests)
}

// Estimates are computed with x/concentrated-liquidity/python/clmath.py
func TestGetNextSqrtPriceFromAmount1InRoundingDown(t *testing.T) {
	tests := map[string]sqrtRoundingDecTestCase{
		"rounded down at precision end": {
			sqrtPriceCurrent: sqrt5000BigDec,
			liquidity:        osmomath.MustNewDecFromStr("3035764687.503020836176699298"),
			amountRemaining:  osmomath.MustNewBigDecFromStr("8398"),

			expected: osmomath.MustNewBigDecFromStr("70.710680885008822823343339270800000167"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("2.5"),
			liquidity:        osmomath.OneDec(),
			amountRemaining:  osmomath.MustNewBigDecFromStr("10"),
			// sqrt_next = sqrt_cur + token_in / liq
			expected: osmomath.MustNewBigDecFromStr("12.5"),
		},
		"happy path": {
			liquidity:        osmomath.MustNewDecFromStr("1519437308.014768571721000000"), // liquidity1 calculated above
			sqrtPriceCurrent: sqrt5000BigDec,                                              // 5000000000
			amountRemaining:  osmomath.NewBigDec(42000000),
			// sqrt_next = sqrt_cur + token_in / liq
			// calculated with x/concentrated-liquidity/python/clmath.py  round_decimal(sqrt_next, 36, ROUND_FLOOR)
			expected: osmomath.MustNewBigDecFromStr("70.738319930382329008049494613660784220"),
		},
		"low price range": {
			liquidity:        smallLiquidity.Dec(),
			sqrtPriceCurrent: sqrtANearMin,
			amountRemaining:  smallValue,
			// from clmath decimal import *
			// get_next_sqrt_price_from_amount1_in_round_down(liq, sqrtPriceA, amountRemaining)
			expected: osmomath.MustNewBigDecFromStr("31964941472737.900293161392817774305123129525585219"),
		},
	}
	runSqrtRoundingDecTestCase(t, "TestGetNextSqrtPriceFromAmount1InRoundingDown", math.GetNextSqrtPriceFromAmount1InRoundingDown, tests)
}

func TestGetNextSqrtPriceFromAmount1OutRoundingDown(t *testing.T) {
	tests := map[string]sqrtRoundingDecTestCase{
		"rounded down at precision end": {
			sqrtPriceCurrent: sqrt5000BigDec,
			liquidity:        osmomath.MustNewDecFromStr("3035764687.503020836176699298"),
			amountRemaining:  osmomath.MustNewBigDecFromStr("8398"),
			// round_osmo_prec_down(sqrtPriceCurrent - round_osmo_prec_up(tokenOut / liquidity))
			expected: osmomath.MustNewBigDecFromStr("70.710675352300682056656660729199999832"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("12.5"),
			liquidity:        osmomath.MustNewDecFromStr("1"),
			amountRemaining:  osmomath.MustNewBigDecFromStr("10"),
			// round_osmo_prec_down(sqrtPriceCurrent - round_osmo_prec_up(tokenOut / liquidity))
			expected: osmomath.MustNewBigDecFromStr("2.5"),
		},
		"low price range": {
			liquidity:        smallLiquidity.Dec(),
			sqrtPriceCurrent: sqrtANearMin,
			amountRemaining:  smallValue,
			// from clmath decimal import *
			// get_next_sqrt_price_from_amount1_out_round_down(liq, sqrtPriceA, amountRemaining)
			// While a negative sqrt price value is invalid and should be caught by the caller,
			// we mostly focus on testing rounding behavior and math correctness at low spot prices.
			// For the purposes of our test, this result is acceptable.
			expected: osmomath.MustNewBigDecFromStr("-31964941472737.900293161392817726681599530362954590"),
		},
	}
	runSqrtRoundingDecTestCase(t, "TestGetNextSqrtPriceFromAmount1OutRoundingDown", math.GetNextSqrtPriceFromAmount1OutRoundingDown, tests)
}
