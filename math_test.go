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
