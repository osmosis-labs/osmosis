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
