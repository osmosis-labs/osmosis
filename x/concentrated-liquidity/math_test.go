package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func (s *KeeperTestSuite) TestGetLiquidityFromAmounts() {
	currentSqrtP, err := sdk.NewDecFromStr("70.710678")
	sqrtPLow, err := sdk.NewDecFromStr("67.082039")
	sqrtPHigh, err := sdk.NewDecFromStr("74.161984")
	s.Require().NoError(err)

	amount0Desired := sdk.NewInt(1)
	amount1Desired := sdk.NewInt(5000)

	s.SetupTest()

	liquidity := cl.GetLiquidityFromAmounts(currentSqrtP, sqrtPLow, sqrtPHigh, amount0Desired, amount1Desired)
	s.Require().Equal("1377.927096082029653542", liquidity.String())
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
