package math_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
)

var (
	DefaultExponentAtPriceOne      = sdk.NewInt(-4)
	DefaultMinTick, DefaultMaxTick = cl.GetMinAndMaxTicksFromExponentAtPriceOne(DefaultExponentAtPriceOne)
)

type ConcentratedMathTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestConcentratedTestSuite(t *testing.T) {
	suite.Run(t, new(ConcentratedMathTestSuite))
}

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity1 = amount1 / (sqrtPriceB - sqrtPriceA)
func (suite *ConcentratedMathTestSuite) TestLiquidity1() {
	testCases := map[string]struct {
		currentSqrtP      sdk.Dec
		sqrtPLow          sdk.Dec
		amount1Desired    sdk.Int
		expectedLiquidity string
	}{
		"happy path": {
			currentSqrtP:      sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPLow:          sdk.MustNewDecFromStr("67.416615162732695594"), // 4545
			amount1Desired:    sdk.NewInt(5000000000),
			expectedLiquidity: "1517882343.751510418088349649",
			// https://www.wolframalpha.com/input?i=5000000000+%2F+%2870.710678118654752440+-+67.416615162732695594%29
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			liquidity := math.Liquidity1(tc.amount1Desired, tc.currentSqrtP, tc.sqrtPLow)
			suite.Require().Equal(tc.expectedLiquidity, liquidity.String())
		})
	}
}

// TestLiquidity0 tests that liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity0 = amount0 * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)
func (suite *ConcentratedMathTestSuite) TestLiquidity0() {
	testCases := map[string]struct {
		currentSqrtP      sdk.Dec
		sqrtPHigh         sdk.Dec
		amount0Desired    sdk.Int
		expectedLiquidity string
	}{
		"happy path": {
			currentSqrtP:      sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPHigh:         sdk.MustNewDecFromStr("74.161984870956629487"), // 5500
			amount0Desired:    sdk.NewInt(1000000),
			expectedLiquidity: "1519437308.014768571720923239",
			// https://www.wolframalpha.com/input?i=1000000+*+%2870.710678118654752440*+74.161984870956629487%29+%2F+%2874.161984870956629487+-+70.710678118654752440%29
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			liquidity := math.Liquidity0(tc.amount0Desired, tc.currentSqrtP, tc.sqrtPHigh)
			suite.Require().Equal(tc.expectedLiquidity, liquidity.String())
		})
	}
}

// TestGetNextSqrtPriceFromAmount0RoundingUp tests that getNextSqrtPriceFromAmount0RoundingUp utilizes
// the current squareRootPrice, liquidity of denom0, and amount of denom0 that still needs
// to be swapped in order to determine the next squareRootPrice
// PATH 1
// if (amountRemaining * sqrtPriceCurrent) / amountRemaining  == sqrtPriceCurrent AND (liquidity) + (amountRemaining * sqrtPriceCurrent) >= (liquidity)
// sqrtPriceNext = (liquidity * sqrtPriceCurrent) / ((liquidity) + (amountRemaining * sqrtPriceCurrent))
// PATH 2
// else
// sqrtPriceNext = ((liquidity)) / (((liquidity) / (sqrtPriceCurrent)) + (amountRemaining))
func (suite *ConcentratedMathTestSuite) TestGetNextSqrtPriceFromAmount0RoundingUp() {
	testCases := map[string]struct {
		liquidity             sdk.Dec
		sqrtPCurrent          sdk.Dec
		amount0Remaining      sdk.Dec
		sqrtPriceNextExpected string
	}{
		"happy path": {
			liquidity:             sdk.MustNewDecFromStr("1517882343.751510418088349649"), // liquidity0 calculated above
			sqrtPCurrent:          sdk.MustNewDecFromStr("70.710678118654752440"),
			amount0Remaining:      sdk.NewDec(13370),
			sqrtPriceNextExpected: "70.666663910857144332",
			// https://www.wolframalpha.com/input?i=%28%281517882343.751510418088349649%29%29+%2F+%28%28%281517882343.751510418088349649%29+%2F+%2870.710678118654752440%29%29+%2B+%2813370%29%29
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPriceNext := math.GetNextSqrtPriceFromAmount0InRoundingUp(tc.sqrtPCurrent, tc.liquidity, tc.amount0Remaining)
			suite.Require().Equal(tc.sqrtPriceNextExpected, sqrtPriceNext.String())
		})
	}
}

// TestGetNextSqrtPriceFromAmount1RoundingDown tests that getNextSqrtPriceFromAmount1RoundingDown
// utilizes the current squareRootPrice, liquidity of denom1, and amount of denom1 that still needs
// to be swapped in order to determine the next squareRootPrice
// sqrtPriceNext = sqrtPriceCurrent + (amount1Remaining / liquidity1)
func (suite *ConcentratedMathTestSuite) TestGetNextSqrtPriceFromAmount1RoundingDown() {
	testCases := map[string]struct {
		liquidity             sdk.Dec
		sqrtPCurrent          sdk.Dec
		amount1Remaining      sdk.Dec
		sqrtPriceNextExpected string
	}{
		"happy path": {
			liquidity:             sdk.MustNewDecFromStr("1519437308.014768571721000000"), // liquidity1 calculated above
			sqrtPCurrent:          sdk.MustNewDecFromStr("70.710678118654752440"),         // 5000000000
			amount1Remaining:      sdk.NewDec(42000000),
			sqrtPriceNextExpected: "70.738319930382329008",
			// https://www.wolframalpha.com/input?i=70.710678118654752440+%2B++++%2842000000+%2F+1519437308.014768571721000000%29
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			sqrtPriceNext := math.GetNextSqrtPriceFromAmount1InRoundingDown(tc.sqrtPCurrent, tc.liquidity, tc.amount1Remaining)
			suite.Require().Equal(tc.sqrtPriceNextExpected, sqrtPriceNext.String())
		})
	}
}

// TestCalcAmount0Delta tests that calcAmount0 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount0Delta = (liquidity * (sqrtPriceB - sqrtPriceA)) / (sqrtPriceB * sqrtPriceA)
func (suite *ConcentratedMathTestSuite) TestCalcAmount0Delta() {
	testCases := map[string]struct {
		liquidity       sdk.Dec
		sqrtPCurrent    sdk.Dec
		sqrtPUpper      sdk.Dec
		amount0Expected string
	}{
		"happy path": {
			liquidity:       sdk.MustNewDecFromStr("1517882343.751510418088349649"), // we use the smaller liquidity between liq0 and liq1
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"),         // 5000
			sqrtPUpper:      sdk.MustNewDecFromStr("74.161984870956629487"),         // 5500
			amount0Expected: "998976.618347426388356620",
			// https://www.wolframalpha.com/input?i=%281517882343.751510418088349649+*+%2874.161984870956629487+-+70.710678118654752440+%29%29+%2F+%2870.710678118654752440+*+74.161984870956629487%29
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			amount0 := math.CalcAmount0Delta(tc.liquidity, tc.sqrtPCurrent, tc.sqrtPUpper, false)
			suite.Require().Equal(tc.amount0Expected, amount0.String())
		})
	}
}

// TestCalcAmount1Delta tests that calcAmount1 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount1Delta = liq * (sqrtPriceB - sqrtPriceA)
func (suite *ConcentratedMathTestSuite) TestCalcAmount1Delta() {
	testCases := map[string]struct {
		liquidity       sdk.Dec
		sqrtPCurrent    sdk.Dec
		sqrtPLower      sdk.Dec
		amount1Expected string
	}{
		"happy path": {
			liquidity:       sdk.MustNewDecFromStr("1517882343.751510418088349649"), // we use the smaller liquidity between liq0 and liq1
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"),         // 5000
			sqrtPLower:      sdk.MustNewDecFromStr("67.416615162732695594"),         // 4545
			amount1Expected: "5000000000.000000000000000000",
			// https://www.wolframalpha.com/input?i=1517882343.751510418088349649+*+%2870.710678118654752440+-+67.416615162732695594%29
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			amount1 := math.CalcAmount1Delta(tc.liquidity, tc.sqrtPCurrent, tc.sqrtPLower, false)
			suite.Require().Equal(tc.amount1Expected, amount1.String())
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestGetLiquidityFromAmounts() {
	testCases := map[string]struct {
		currentSqrtP      sdk.Dec
		sqrtPHigh         sdk.Dec
		sqrtPLow          sdk.Dec
		amount0Desired    sdk.Int
		amount1Desired    sdk.Int
		expectedLiquidity string
	}{
		"happy path": {
			currentSqrtP:      sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			sqrtPHigh:         sdk.MustNewDecFromStr("74.161984870956629487"), // 5500
			sqrtPLow:          sdk.MustNewDecFromStr("67.416615162732695594"), // 4545
			amount0Desired:    sdk.NewInt(1000000),
			amount1Desired:    sdk.NewInt(5000000000),
			expectedLiquidity: "1517882343.751510418088349649",
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			// CASE A: if the currentSqrtP is less than the sqrtPLow, all the liquidity is in asset0, so GetLiquidityFromAmounts returns the liquidity of asset0
			// CASE B: if the currentSqrtP is less than the sqrtPHigh but greater than sqrtPLow, the liquidity is split between asset0 and asset1,
			// so GetLiquidityFromAmounts returns the smaller liquidity of asset0 and asset1
			// CASE C: if the currentSqrtP is greater than the sqrtPHigh, all the liquidity is in asset1, so GetLiquidityFromAmounts returns the liquidity of asset1
			liquidity := math.GetLiquidityFromAmounts(tc.currentSqrtP, tc.sqrtPLow, tc.sqrtPHigh, tc.amount0Desired, tc.amount1Desired)
			suite.Require().Equal(tc.expectedLiquidity, liquidity.String())
			// TODO: this check works for CASE B but needs to get reworked when CASE A and CASE C are tested
			liq0 := math.Liquidity0(tc.amount0Desired, tc.currentSqrtP, tc.sqrtPHigh)
			liq1 := math.Liquidity1(tc.amount1Desired, tc.currentSqrtP, tc.sqrtPLow)
			liq := sdk.MinDec(liq0, liq1)
			suite.Require().Equal(liq.String(), liquidity.String())

		})
	}
}

func (suite *ConcentratedMathTestSuite) TestGetNextSqrtPriceFromAmount0InRoundingUp() {
	tests := map[string]struct {
		sqrtPriceCurrent     sdk.Dec
		liquidity            sdk.Dec
		amountZeroRemaininIn sdk.Dec

		expectedSqrtPriceNext sdk.Dec
	}{
		"rounded up at precision end": {
			sqrtPriceCurrent:     sdk.MustNewDecFromStr("70.710678118654752440"),
			liquidity:            sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountZeroRemaininIn: sdk.MustNewDecFromStr("8398"),

			// liq * sqrt_cur / (liq + token_in * sqrt_cur) = 70.69684905341696614869539245
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("70.696849053416966149"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent:     sdk.MustNewDecFromStr("2"),
			liquidity:            sdk.MustNewDecFromStr("10"),
			amountZeroRemaininIn: sdk.MustNewDecFromStr("15"),

			// liq * sqrt_cur / (liq + token_in * sqrt_cur) = 0.5
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("0.5"),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {

			sqrtPriceNext := math.GetNextSqrtPriceFromAmount0InRoundingUp(tc.sqrtPriceCurrent, tc.liquidity, tc.amountZeroRemaininIn)

			suite.Require().Equal(tc.expectedSqrtPriceNext.String(), sqrtPriceNext.String())
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestGetNextSqrtPriceFromAmount0OutRoundingUp() {
	tests := map[string]struct {
		sqrtPriceCurrent       sdk.Dec
		liquidity              sdk.Dec
		amountZeroRemainingOut sdk.Dec

		expectedSqrtPriceNext sdk.Dec
	}{
		"rounded up at precision end": {
			sqrtPriceCurrent:       sdk.MustNewDecFromStr("70.710678118654752440"),
			liquidity:              sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountZeroRemainingOut: sdk.MustNewDecFromStr("8398"),

			// liq * sqrt_cur / (liq - token_out * sqrt_cur) = 70.72451259517930556540769876
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("70.724512595179305566"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent:       sdk.MustNewDecFromStr("2"),
			liquidity:              sdk.MustNewDecFromStr("10"),
			amountZeroRemainingOut: sdk.MustNewDecFromStr("1"),

			// liq * sqrt_cur / (liq + token_out * sqrt_cur) = 2.5
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("2.5"),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {

			sqrtPriceNext := math.GetNextSqrtPriceFromAmount0OutRoundingUp(tc.sqrtPriceCurrent, tc.liquidity, tc.amountZeroRemainingOut)

			suite.Require().Equal(tc.expectedSqrtPriceNext.String(), sqrtPriceNext.String())
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestGetNextSqrtPriceFromAmount1InRoundingDown() {
	tests := map[string]struct {
		sqrtPriceCurrent     sdk.Dec
		liquidity            sdk.Dec
		amountOneRemainingIn sdk.Dec

		expectedSqrtPriceNext sdk.Dec
	}{
		"rounded down at precision end": {
			sqrtPriceCurrent:     sdk.MustNewDecFromStr("70.710678118654752440"),
			liquidity:            sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountOneRemainingIn: sdk.MustNewDecFromStr("8398"),

			// sqrt_next = sqrt_cur + token_in / liq = 70.71068088500882282334333927
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("70.710680885008822823"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent:     sdk.MustNewDecFromStr("2.5"),
			liquidity:            sdk.MustNewDecFromStr("1"),
			amountOneRemainingIn: sdk.MustNewDecFromStr("10"),

			// sqrt_next = sqrt_cur + token_in / liq
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("12.5"),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {

			sqrtPriceNext := math.GetNextSqrtPriceFromAmount1InRoundingDown(tc.sqrtPriceCurrent, tc.liquidity, tc.amountOneRemainingIn)

			suite.Require().Equal(tc.expectedSqrtPriceNext.String(), sqrtPriceNext.String())
		})
	}
}

func (suite *ConcentratedMathTestSuite) TestGetNextSqrtPriceFromAmount1OutRoundingDown() {
	tests := map[string]struct {
		sqrtPriceCurrent      sdk.Dec
		liquidity             sdk.Dec
		amountOneRemainingOut sdk.Dec

		expectedSqrtPriceNext sdk.Dec
	}{
		"rounded down at precision end": {
			sqrtPriceCurrent:      sdk.MustNewDecFromStr("70.710678118654752440"),
			liquidity:             sdk.MustNewDecFromStr("3035764687.503020836176699298"),
			amountOneRemainingOut: sdk.MustNewDecFromStr("8398"),

			// sqrt_next = sqrt_cur - token_out / liq = 70.71067535230068205665666073
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("70.710675352300682056"),
		},
		"no round up due zeroes at precision end": {
			sqrtPriceCurrent:      sdk.MustNewDecFromStr("12.5"),
			liquidity:             sdk.MustNewDecFromStr("1"),
			amountOneRemainingOut: sdk.MustNewDecFromStr("10"),

			// sqrt_next = sqrt_cur - token_out / liq
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("2.5"),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {

			sqrtPriceNext := math.GetNextSqrtPriceFromAmount1OutRoundingDown(tc.sqrtPriceCurrent, tc.liquidity, tc.amountOneRemainingOut)

			suite.Require().Equal(tc.expectedSqrtPriceNext.String(), sqrtPriceNext.String())
		})
	}
}
