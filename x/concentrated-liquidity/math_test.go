package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func (suite *KeeperTestSuite) TestGetLiquidityFromAmounts() {
	testCases := []struct {
		name              string
		currentSqrtP      sdk.Dec
		sqrtPHigh         sdk.Dec
		sqrtPLow          sdk.Dec
		amount0Desired    sdk.Int
		amount1Desired    sdk.Int
		expectedLiquidity string
	}{
		{
			"happy path",
			sdk.MustNewDecFromStr("70.710678"),
			sdk.MustNewDecFromStr("74.161984"),
			sdk.MustNewDecFromStr("67.082039"),
			sdk.NewInt(1),
			sdk.NewInt(5000),
			"1377.927096082029653542",
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			liquidity := cl.GetLiquidityFromAmounts(tc.currentSqrtP, tc.sqrtPLow, tc.sqrtPHigh, tc.amount0Desired, tc.amount1Desired)
			suite.Require().Equal(tc.expectedLiquidity, liquidity.String())
			liq0 := cl.Liquidity0(tc.amount0Desired, tc.currentSqrtP, tc.sqrtPHigh)
			liq1 := cl.Liquidity1(tc.amount1Desired, tc.currentSqrtP, tc.sqrtPLow)
			liq := sdk.MinDec(liq0, liq1)
			suite.Require().Equal(liq.String(), liquidity.String())

		})
	}
}

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity1 = amount1 / (sqrtPriceB - sqrtPriceA)
func (suite *KeeperTestSuite) TestLiquidity1() {
	testCases := []struct {
		name              string
		currentSqrtP      sdk.Dec
		sqrtPLow          sdk.Dec
		amount1Desired    sdk.Int
		expectedLiquidity string
	}{
		{
			"happy path",
			sdk.NewDecWithPrec(70710678, 6),
			sdk.NewDecWithPrec(67082039, 6),
			sdk.NewInt(5000),
			"1377.927096082029653542",
			// https://www.wolframalpha.com/input?i=5000+%2F+%2870.710678+-+67.082039%29
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			liquidity := cl.Liquidity1(tc.amount1Desired, tc.currentSqrtP, tc.sqrtPLow)
			suite.Require().Equal(tc.expectedLiquidity, liquidity.String())
		})
	}
}

// TestLiquidity0 tests that liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity0 = amount0 * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)
func (suite *KeeperTestSuite) TestLiquidity0() {
	testCases := []struct {
		name              string
		currentSqrtP      sdk.Dec
		sqrtPHigh         sdk.Dec
		amount0Desired    sdk.Int
		expectedLiquidity string
	}{
		{
			"happy path",
			sdk.MustNewDecFromStr("70.710678"),
			sdk.MustNewDecFromStr("74.161984"),
			sdk.NewInt(1),
			"1519.437618821730672389",
			// https://www.wolframalpha.com/input?i=1+*+%2870.710678+*+74.161984%29+%2F+%2874.161984+-+70.710678%29
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			liquidity := cl.Liquidity0(tc.amount0Desired, tc.currentSqrtP, tc.sqrtPHigh)
			suite.Require().Equal(tc.expectedLiquidity, liquidity.String())
		})
	}
}

// TestGetNextSqrtPriceFromAmount0RoundingUp tests that getNextSqrtPriceFromAmount0RoundingUp utilizes
// the current squareRootPrice, liquidity of denom0, and amount of denom0 that still needs
// to be swapped in order to determine the next squareRootPrice
// PATH 1
// if (amountRemaining * sqrtPriceCurrent) / amountRemaining  == sqrtPriceCurrent AND (liquidity * 2) + (amountRemaining * sqrtPriceCurrent) >= (liquidity * 2)
// sqrtPriceNext = (liquidity * 2 * sqrtPriceCurrent) / ((liquidity * 2) + (amountRemaining * sqrtPriceCurrent))
// PATH 2
// else
// sqrtPriceNext = ((liquidity * 2)) / (((liquidity * 2) / (sqrtPriceCurrent)) + (amountRemaining))
func (suite *KeeperTestSuite) TestGetNextSqrtPriceFromAmount0RoundingUp() {
	testCases := []struct {
		name                  string
		liquidity             sdk.Dec
		sqrtPCurrent          sdk.Dec
		amount0Remaining      sdk.Dec
		sqrtPriceNextExpected string
	}{
		{
			"happy path 1",
			sdk.NewDec(1377927219),
			sdk.NewDecWithPrec(70710678, 6),
			sdk.NewDec(133700),
			"70.468932817327539027",
			// https://www.wolframalpha.com/input?i=%281377927219+*+2+*+70.710678%29+%2F+%28%281377927219+*+2%29+%2B+%28133700+*+70.710678%29%29
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			sqrtPriceNext := cl.GetNextSqrtPriceFromAmount0RoundingUp(tc.sqrtPCurrent, tc.liquidity, tc.amount0Remaining)
			suite.Require().Equal(tc.sqrtPriceNextExpected, sqrtPriceNext.String())
		})
	}
}

// TestGetNextSqrtPriceFromAmount1RoundingDown tests that getNextSqrtPriceFromAmount1RoundingDown
// utilizes the current squareRootPrice, liquidity of denom1, and amount of denom1 that still needs
// to be swapped in order to determine the next squareRootPrice
// sqrtPriceNext = sqrtPriceCurrent + (amount1Remaining / liquidity1)
func (suite *KeeperTestSuite) TestGetNextSqrtPriceFromAmount1RoundingDown() {
	testCases := []struct {
		name                  string
		liquidity             sdk.Dec
		sqrtPCurrent          sdk.Dec
		amount1Remaining      sdk.Dec
		sqrtPriceNextExpected string
	}{
		{
			"happy path",
			sdk.NewDec(1377927219),
			sdk.NewDecWithPrec(70710678, 6),
			sdk.NewDec(42000000),
			"70.741158564880981569",
			// https://www.wolframalpha.com/input?i=70.710678+%2B+%2842000000+%2F+1377927219%29
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			sqrtPriceNext := cl.GetNextSqrtPriceFromAmount1RoundingDown(tc.sqrtPCurrent, tc.liquidity, tc.amount1Remaining)
			suite.Require().Equal(tc.sqrtPriceNextExpected, sqrtPriceNext.String())
		})
	}
}

// TestCalcAmount0Delta tests that calcAmount0 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount0Delta = (liquidity * (sqrtPriceB - sqrtPriceA)) / (sqrtPriceB * sqrtPriceA)
func (suite *KeeperTestSuite) TestCalcAmount0Delta() {
	testCases := []struct {
		name            string
		liquidity       sdk.Dec
		sqrtPCurrent    sdk.Dec
		sqrtPUpper      sdk.Dec
		amount0Expected string
	}{
		{
			"happy path",
			sdk.NewDec(1377927219),
			sdk.MustNewDecFromStr("70.710678"),
			sdk.MustNewDecFromStr("74.161984"),
			"906866",
			// https://www.wolframalpha.com/input?i=%281377927219+*+%2874.161984+-+70.710678+%29%29+%2F+%2870.710678+*+74.161984%29
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			amount0 := cl.CalcAmount0Delta(tc.liquidity, tc.sqrtPCurrent, tc.sqrtPUpper)
			suite.Require().Equal(tc.amount0Expected, amount0.TruncateInt().String())
		})
	}
}

// TestCalcAmount1Delta tests that calcAmount1 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount1Delta = liq * (sqrtPriceB - sqrtPriceA)
func (suite *KeeperTestSuite) TestCalcAmount1Delta() {
	testCases := []struct {
		name            string
		liquidity       sdk.Dec
		sqrtPCurrent    sdk.Dec
		sqrtPLower      sdk.Dec
		amount1Expected string
	}{
		{
			"happy path",
			sdk.NewDec(1377927219),
			sdk.NewDecWithPrec(70710678, 6),
			sdk.NewDecWithPrec(67082039, 6),
			"5000000446",
			// https://www.wolframalpha.com/input?i=1377927219+*+%2870.710678+-+67.082039%29
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			amount1 := cl.CalcAmount1Delta(tc.liquidity, tc.sqrtPCurrent, tc.sqrtPLower)
			suite.Require().Equal(tc.amount1Expected, amount1.TruncateInt().String())
		})
	}
}

func (suite *KeeperTestSuite) TestComputeSwapState() {
	testCases := []struct {
		name                  string
		sqrtPCurrent          sdk.Dec
		sqrtPTarget           sdk.Dec
		liquidity             sdk.Dec
		amountRemaining       sdk.Dec
		zeroForOne            bool
		expectedSqrtPriceNext string
		expectedAmountIn      string
		expectedAmountOut     string
	}{
		{
			"happy path: trade asset0 for asset1",
			sdk.NewDecWithPrec(70710678, 6),
			sdk.OneDec(),
			sdk.NewDec(1377927219),
			sdk.NewDec(133700),
			true,
			"70.468932817327539027",
			"66849.999999999999897227",
			"333107267.266511136411924087",
		},
		{
			"happy path: trade asset1 for asset0",
			sdk.NewDecWithPrec(70710678, 6),
			sdk.OneDec(),
			sdk.NewDec(1377927219),
			sdk.NewDec(4199999999),
			false,
			"73.758734487372429211",
			"4199999998.999999999987594209",
			"805287.266898087447354318",
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			sqrtPriceNext, amountIn, amountOut := cl.ComputeSwapStep(tc.sqrtPCurrent, tc.sqrtPTarget, tc.liquidity, tc.amountRemaining, tc.zeroForOne)
			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext.String())
			suite.Require().Equal(tc.expectedAmountIn, amountIn.String())
			suite.Require().Equal(tc.expectedAmountOut, amountOut.String())
		})
	}
}
