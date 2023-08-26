package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/swapstrategy"
)

var onePercentSpreadFactor = sdk.NewDecWithPrec(1, 2)

func (suite *StrategyTestSuite) TestComputespreadRewardChargePerSwapStepOutGivenIn() {
	tests := map[string]struct {
		currentSqrtPrice         sdk.Dec
		hasReachedTarget         bool
		amountIn                 sdk.Dec
		amountSpecifiedRemaining sdk.Dec
		spreadFactor             sdk.Dec

		expectedspreadRewardCharge sdk.Dec
		expectPanic                bool
	}{
		"reached target -> charge spread factor on amount in": {
			hasReachedTarget:         true,
			amountIn:                 sdk.NewDec(100),
			amountSpecifiedRemaining: five,
			spreadFactor:             onePercentSpreadFactor,

			// amount in * spread factor / (1 - spread factor)
			expectedspreadRewardCharge: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(sdk.NewDec(100), onePercentSpreadFactor),
		},
		"did not reach target -> charge spread factor on the difference between amount remaining and amount in": {
			hasReachedTarget:         false,
			amountIn:                 five,
			amountSpecifiedRemaining: sdk.NewDec(100),
			spreadFactor:             onePercentSpreadFactor,

			expectedspreadRewardCharge: sdk.MustNewDecFromStr("95"),
		},
		"zero spread factor": {
			hasReachedTarget:         true,
			amountIn:                 five,
			amountSpecifiedRemaining: sdk.NewDec(100),
			spreadFactor:             sdk.ZeroDec(),

			expectedspreadRewardCharge: sdk.ZeroDec(),
		},
		"negative spread factor - panic": {
			hasReachedTarget:         false,
			amountIn:                 sdk.NewDec(100),
			amountSpecifiedRemaining: five,
			spreadFactor:             sdk.OneDec().Neg(),

			expectPanic: true,
		},
		"amount specified remaining < amount in leads to negative spread factor - panic": {
			hasReachedTarget:         false,
			amountIn:                 sdk.NewDec(102),
			amountSpecifiedRemaining: sdk.NewDec(101),
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
