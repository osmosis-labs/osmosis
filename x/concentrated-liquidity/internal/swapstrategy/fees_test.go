package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
)

var (
	onePercentFee = sdk.NewDecWithPrec(1, 2)
)

func (suite *StrategyTestSuite) TestComputeFeeChargePerSwapStepOutGivenIn() {
	var (
		defaultCurrPrice        = sdk.NewDec(5000)
		defaultCurrSqrtPrice, _ = defaultCurrPrice.ApproxSqrt() // 70.710678118654752440
	)

	tests := map[string]struct {
		currentSqrtPrice         sdk.Dec
		hasReachedTarget         bool
		amountIn                 sdk.Dec
		amountSpecifiedRemaining sdk.Dec
		swapFee                  sdk.Dec

		expectedFeeCharge sdk.Dec
		expectPanic       bool
	}{
		"reached target -> charge fee on amount in": {
			currentSqrtPrice:         defaultCurrSqrtPrice,
			hasReachedTarget:         true,
			amountIn:                 sdk.NewDec(100),
			amountSpecifiedRemaining: five,
			swapFee:                  onePercentFee,

			// amount in * swap fee / (1 - swap fee)
			expectedFeeCharge: sdk.NewDec(100).Mul(onePercentFee).Quo(sdk.OneDec().Sub(onePercentFee)),
		},
		"did not reach target -> charge fee on the difference between amount remaining and amount in": {
			currentSqrtPrice:         defaultCurrSqrtPrice,
			hasReachedTarget:         false,
			amountIn:                 five,
			amountSpecifiedRemaining: sdk.NewDec(100),
			swapFee:                  onePercentFee,

			expectedFeeCharge: sdk.MustNewDecFromStr("95"),
		},
		"zero swap fee": {
			currentSqrtPrice:         defaultCurrSqrtPrice,
			hasReachedTarget:         true,
			amountIn:                 five,
			amountSpecifiedRemaining: sdk.NewDec(100),
			swapFee:                  sdk.ZeroDec(),

			expectedFeeCharge: sdk.ZeroDec(),
		},
		"negative swap fee - panic": {
			currentSqrtPrice:         defaultCurrSqrtPrice,
			hasReachedTarget:         false,
			amountIn:                 sdk.NewDec(100),
			amountSpecifiedRemaining: five,
			swapFee:                  sdk.OneDec().Neg(),

			expectPanic: true,
		},
		"amount specified remaining < amount in leads to negative fee - panic": {
			currentSqrtPrice:         defaultCurrSqrtPrice,
			hasReachedTarget:         false,
			amountIn:                 sdk.NewDec(102),
			amountSpecifiedRemaining: sdk.NewDec(101),
			swapFee:                  onePercentFee,

			// 101 - 102 = -1 -> panic
			expectPanic: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				actualFeeCharge := swapstrategy.ComputeFeeChargePerSwapStepOutGivenIn(tc.currentSqrtPrice, tc.hasReachedTarget, tc.amountIn, tc.amountSpecifiedRemaining, tc.swapFee)

				suite.Require().Equal(tc.expectedFeeCharge, actualFeeCharge)
			})
		})
	}
}

func (suite *StrategyTestSuite) TestGetAmountRemainingLessFee() {
	tests := map[string]struct {
		amountRemaining sdk.Dec
		swapFee         sdk.Dec
		isOutGivenIn    bool

		expected sdk.Dec
	}{
		"out given in - the fee is charged": {
			amountRemaining: five,
			swapFee:         onePercentFee,
			isOutGivenIn:    true,

			expected: five.Mul(sdk.OneDec().Sub(onePercentFee)),
		},
		"in given out - the fee is not charged": {
			amountRemaining: four,
			swapFee:         onePercentFee,
			isOutGivenIn:    false,

			expected: four,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			actual := swapstrategy.GetAmountRemainingLessFee(tc.amountRemaining, tc.swapFee, tc.isOutGivenIn)

			suite.Require().Equal(tc.expected, actual)
		})
	}
}
