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
