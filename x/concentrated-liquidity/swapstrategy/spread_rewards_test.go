package swapstrategy_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/swapstrategy"
)

var onePercentSpreadFactor = osmomath.NewDecWithPrec(1, 2)

func (suite *StrategyTestSuite) TestComputespreadRewardChargePerSwapStepOutGivenIn() {
	tests := map[string]struct {
		currentSqrtPrice         osmomath.Dec
		hasReachedTarget         bool
		amountIn                 osmomath.Dec
		amountSpecifiedRemaining osmomath.Dec
		spreadFactor             osmomath.Dec

		expectedspreadRewardCharge osmomath.Dec
		expectPanic                bool
	}{
		"reached target -> charge spread factor on amount in": {
			hasReachedTarget:         true,
			amountIn:                 osmomath.NewDec(100),
			amountSpecifiedRemaining: five,
			spreadFactor:             onePercentSpreadFactor,

			// amount in * spread factor / (1 - spread factor)
			expectedspreadRewardCharge: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(osmomath.NewDec(100), onePercentSpreadFactor),
		},
		"did not reach target -> charge spread factor on the difference between amount remaining and amount in": {
			hasReachedTarget:         false,
			amountIn:                 five,
			amountSpecifiedRemaining: osmomath.NewDec(100),
			spreadFactor:             onePercentSpreadFactor,

			expectedspreadRewardCharge: osmomath.MustNewDecFromStr("95"),
		},
		"zero spread factor": {
			hasReachedTarget:           true,
			amountIn:                   five,
			amountSpecifiedRemaining:   osmomath.NewDec(100),
			spreadFactor:               osmomath.ZeroDec(),
			expectedspreadRewardCharge: osmomath.ZeroDec(),
		},
		"negative spread factor - panic": {
			hasReachedTarget:         false,
			amountIn:                 osmomath.NewDec(100),
			amountSpecifiedRemaining: five,
			spreadFactor:             osmomath.OneDec().Neg(),

			expectPanic: true,
		},
		"amount specified remaining < amount in leads to negative spread factor - panic": {
			hasReachedTarget:         false,
			amountIn:                 osmomath.NewDec(102),
			amountSpecifiedRemaining: osmomath.NewDec(101),
			spreadFactor:             onePercentSpreadFactor,

			// 101 - 102 = -1 -> panic
			expectPanic: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				actualspreadRewardCharge := swapstrategy.ComputeSpreadRewardChargePerSwapStepOutGivenIn(tc.hasReachedTarget, tc.amountIn, tc.amountSpecifiedRemaining, tc.spreadFactor)

				suite.Require().Equal(tc.expectedspreadRewardCharge, actualspreadRewardCharge)
			})
		})
	}
}
