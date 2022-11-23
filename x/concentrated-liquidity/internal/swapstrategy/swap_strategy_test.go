package swapstrategy_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/swapstrategy"
)

type StrategyTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(StrategyTestSuite))
}

// TODO: split up this test case to be separate for each strategy.
func (suite *StrategyTestSuite) TestComputeSwapState() {
	testCases := map[string]struct {
		sqrtPCurrent          sdk.Dec
		nextSqrtPrice         sdk.Dec
		liquidity             sdk.Dec
		amountRemaining       sdk.Dec
		sqrtPriceLimit        sdk.Dec
		zeroForOne            bool
		expectedSqrtPriceNext string
		expectedAmountIn      string
		expectedAmountOut     string
	}{
		"happy path: trade asset0 for asset1": {
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			nextSqrtPrice:   sdk.MustNewDecFromStr("70.666662070529219856"), // 4993.777128190373086350
			liquidity:       sdk.MustNewDecFromStr("1517818840.967515822610790519"),
			amountRemaining: sdk.NewDec(13370),
			// sqrt price limit is less than sqrt price target so it does not affect the result
			// TODO: test case where it does affect.
			sqrtPriceLimit:        sdk.MustNewDecFromStr("70.666662070529219856").Sub(sdk.OneDec()),
			zeroForOne:            true,
			expectedSqrtPriceNext: "70.666662070529219856",
			expectedAmountIn:      "13369.999999903622360944",
			expectedAmountOut:     "66808387.149866264039333362",
		},
		"happy path: trade asset1 for asset0": {
			sqrtPCurrent:    sdk.MustNewDecFromStr("70.710678118654752440"), // 5000
			nextSqrtPrice:   sdk.MustNewDecFromStr("70.738349405152439867"), // 5003.91407656543054317
			liquidity:       sdk.MustNewDecFromStr("1517818840.967515822610790519"),
			amountRemaining: sdk.NewDec(42000000),
			// sqrt price limit is less than sqrt price target so it does not affect the result
			// TODO: test case where it does affect.
			sqrtPriceLimit:        sdk.MustNewDecFromStr("70.666662070529219856").Add(sdk.OneDec()),
			zeroForOne:            false,
			expectedSqrtPriceNext: "70.738349405152439867",
			expectedAmountIn:      "42000000.000000000650233591",
			expectedAmountOut:     "8396.714104746015980302",
		},
	}

	for name, tc := range testCases {
		tc := tc

		suite.Run(name, func() {
			swapStrategy := swapstrategy.New(tc.zeroForOne, tc.sqrtPriceLimit)
			sqrtPriceNext, amountIn, amountOut := swapStrategy.ComputeSwapStep(tc.sqrtPCurrent, tc.nextSqrtPrice, tc.liquidity, tc.amountRemaining)
			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext.String())
			suite.Require().Equal(tc.expectedAmountIn, amountIn.String())
			suite.Require().Equal(tc.expectedAmountOut, amountOut.String())
		})
	}
}
