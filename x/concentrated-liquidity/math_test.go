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

func (suite *KeeperTestSuite) TestGetNextSqrtPriceFromAmount0RoundingUp() {
	testCases := []struct {
		name                  string
		liquidity             sdk.Dec
		sqrtPCurrent          sdk.Dec
		amount0Remaining      sdk.Dec
		sqrtPriceNextExpected string
	}{
		{
			"happy path",
			sdk.NewDec(1377927219),
			sdk.NewDecWithPrec(70710678, 6),
			sdk.NewDec(133700),
			"70.468932817327539027",
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
