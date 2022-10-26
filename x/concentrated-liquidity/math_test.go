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
			"70.468932817413918583",
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
			"70.741158564870052996",
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
