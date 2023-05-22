package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/swapstrategy"
)

var onePercentFee = sdk.NewDecWithPrec(1, 2)

func (suite *StrategyTestSuite) TestComputeFeeChargePerSwapStepOutGivenIn() {
	tests := map[string]struct {
		currentSqrtPrice         sdk.Dec
		hasReachedTarget         bool
		amountIn                 sdk.Dec
		amountSpecifiedRemaining sdk.Dec
		spreadFactor             sdk.Dec

		expectedFeeCharge sdk.Dec
		expectPanic       bool
	}{
		"reached target -> charge fee on amount in": {
			hasReachedTarget:         true,
			amountIn:                 sdk.NewDec(100),
			amountSpecifiedRemaining: five,
			spreadFactor:             onePercentFee,

			// amount in * spread factor / (1 - spread factor)
			expectedFeeCharge: swapstrategy.ComputeFeeChargeFromAmountIn(sdk.NewDec(100), onePercentFee),
		},
		"did not reach target -> charge fee on the difference between amount remaining and amount in": {
			hasReachedTarget:         false,
			amountIn:                 five,
			amountSpecifiedRemaining: sdk.NewDec(100),
			spreadFactor:             onePercentFee,

			expectedFeeCharge: sdk.MustNewDecFromStr("95"),
		},
		"zero spread factor": {
			hasReachedTarget:         true,
			amountIn:                 five,
			amountSpecifiedRemaining: sdk.NewDec(100),
			spreadFactor:             sdk.ZeroDec(),

			expectedFeeCharge: sdk.ZeroDec(),
		},
		"negative spread factor - panic": {
			hasReachedTarget:         false,
			amountIn:                 sdk.NewDec(100),
			amountSpecifiedRemaining: five,
			spreadFactor:             sdk.OneDec().Neg(),

			expectPanic: true,
		},
		"amount specified remaining < amount in leads to negative fee - panic": {
			hasReachedTarget:         false,
			amountIn:                 sdk.NewDec(102),
			amountSpecifiedRemaining: sdk.NewDec(101),
			spreadFactor:             onePercentFee,

			// 101 - 102 = -1 -> panic
			expectPanic: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				actualFeeCharge := swapstrategy.ComputeFeeChargePerSwapStepOutGivenIn(tc.hasReachedTarget, tc.amountIn, tc.amountSpecifiedRemaining, tc.spreadFactor)

				suite.Require().Equal(tc.expectedFeeCharge, actualFeeCharge)
			})
		})
	}
}
