package swapstrategy_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	defaultSqrtPriceLowerOfz = sdk.MustNewDecFromStr("70.688664163408836321") // approx 4996.89
	defaultSqrtPriceUpperOfz = sdk.MustNewDecFromStr("70.710678118654752440") // approx 5000
	defaultAmountOneOfz      = sdk.MustNewDecFromStr("66829187.967824033199646915")
	defaultAmountZeroOfz     = sdk.MustNewDecFromStr("13369.999999999998920002")
)

func (suite *StrategyTestSuite) TestGetSqrtTargetPrice_OneForZero() {
	tests := map[string]struct {
		sqrtPriceLimit    sdk.Dec
		nextTickSqrtPrice sdk.Dec
		expectedResult    sdk.Dec
	}{
		"nextTickSqrtPrice == sqrtPriceLimit -> returns either": {
			sqrtPriceLimit:    sdk.OneDec(),
			nextTickSqrtPrice: sdk.OneDec(),
			expectedResult:    sdk.OneDec(),
		},
		"nextTickSqrtPrice > sqrtPriceLimit -> sqrtPriceLimit": {
			sqrtPriceLimit:    three,
			nextTickSqrtPrice: four,
			expectedResult:    three,
		},
		"nextTickSqrtPrice < sqrtPriceLimit -> nextTickSqrtPrice": {
			sqrtPriceLimit:    five,
			nextTickSqrtPrice: two,
			expectedResult:    two,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			sut := swapstrategy.New(false, tc.sqrtPriceLimit, suite.App.GetKey(types.ModuleName), sdk.ZeroDec())

			actualSqrtTargetPrice := sut.GetSqrtTargetPrice(tc.nextTickSqrtPrice)

			suite.Require().Equal(tc.expectedResult, actualSqrtTargetPrice)

		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_OneForZero() {
	tokenOutErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.SmallestDec().MulInt64(200),
	}

	var (
		sqrtPriceCurrent = defaultSqrtPriceLowerOfz
		sqrtPriceNext    = defaultSqrtPriceUpperOfz

		// sqrt_price_current + token_in / liquidity
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.710678085714122880")
		// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
		amountZeroTargetNotReached = sdk.MustNewDecFromStr("13369.979999999989129753")
	)

	tests := map[string]struct {
		sqrtPriceCurrent     sdk.Dec
		sqrtPriceTarget      sdk.Dec
		liquidity            sdk.Dec
		amountOneInRemaining sdk.Dec
		swapFee              sdk.Dec

		expectedSqrtPriceNext    sdk.Dec
		expectedAmountInConsumed sdk.Dec
		expectedAmountOut        sdk.Dec
		expectedFeeChargeTotal   sdk.Dec

		expectError error
	}{
		"1: no fee - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountOneInRemaining: defaultAmountOneOfz.Add(sdk.NewDec(100)),
			swapFee:              sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Reached target, so 100 is not consumed.
			expectedAmountInConsumed: defaultAmountOneOfz.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:      defaultAmountZeroOfz,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOneOfz.Sub(sdk.NewDec(100)),
			swapFee:              sdk.ZeroDec(),

			// sqrt_price_current + token_in / liquidity
			expectedSqrtPriceNext:    sqrtPriceTargetNotReached,
			expectedAmountInConsumed: defaultAmountOneOfz.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOut:        amountZeroTargetNotReached,
			expectedFeeChargeTotal:   sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOneOfz.Add(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:              defaultFee,

			expectedSqrtPriceNext:    sqrtPriceNext,
			expectedAmountInConsumed: defaultAmountOneOfz.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:      defaultAmountZeroOfz,
			expectedFeeChargeTotal: defaultAmountOneOfz.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOneOfz.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:              defaultFee,

			expectedSqrtPriceNext:    sqrtPriceTargetNotReached,
			expectedAmountInConsumed: defaultAmountOneOfz.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOut:        amountZeroTargetNotReached,
			// Difference between given amount remaining in and amount in actually consumed which qpproximately equals to fee.
			expectedFeeChargeTotal: defaultAmountOneOfz.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)).Sub(defaultAmountOneOfz.Sub(sdk.NewDec(100)).Ceil()),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			strategy := swapstrategy.New(false, types.MaxSqrtRatio, suite.App.GetKey(types.ModuleName), tc.swapFee)

			sqrtPriceNext, amountInConsumed, amountZeroOut, feeChargeTotal := strategy.ComputeSwapStepOutGivenIn(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountOneInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountInConsumed, amountInConsumed)
			suite.Require().Equal(0,
				tokenOutErrTolerance.CompareBigDec(
					osmomath.BigDecFromSDKDec(tc.expectedAmountOut),
					osmomath.BigDecFromSDKDec(amountZeroOut),
				),
				fmt.Sprintf("expected (%s), actual (%s)", tc.expectedAmountOut, amountZeroOut))
			suite.Require().Equal(tc.expectedFeeChargeTotal, feeChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_OneForZero() {
	tokenOutErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.NewDecWithPrec(1, 5),
	}

	smallestErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.SmallestDec().MulInt64(2),
	}

	var (
		sqrtPriceCurrent = defaultSqrtPriceLowerOfz
		sqrtPriceNext    = defaultSqrtPriceUpperOfz

		// sqrtPriceCurrent + amount out / liquidity
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.710513415890590952")
		// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
		amountZeroTargetNotReached = sdk.MustNewDecFromStr("13269.999999999999410905")
	)

	tests := map[string]struct {
		sqrtPriceCurrent       sdk.Dec
		sqrtPriceTarget        sdk.Dec
		liquidity              sdk.Dec
		amountZeroOutRemaining sdk.Dec
		swapFee                sdk.Dec

		expectedSqrtPriceNext         sdk.Dec
		expectedAmountZeroOutConsumed sdk.Dec
		expectedAmountOneIn           sdk.Dec
		expectedFeeChargeTotal        sdk.Dec

		expectError error
	}{
		"1: no fee - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountZeroOutRemaining: defaultAmountZeroOfz.Add(sdk.NewDec(100)),
			swapFee:                sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Reached target, so 100 is not consumed.
			expectedAmountZeroOutConsumed: defaultAmountZeroOfz,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneIn:    defaultAmountOneOfz.Ceil(),
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			sqrtPriceCurrent:       sqrtPriceCurrent,
			sqrtPriceTarget:        sqrtPriceNext,
			liquidity:              defaultLiquidity,
			amountZeroOutRemaining: defaultAmountZeroOfz.Sub(sdk.NewDec(100)),
			swapFee:                sdk.ZeroDec(),

			expectedSqrtPriceNext:         sqrtPriceTargetNotReached,
			expectedAmountZeroOutConsumed: amountZeroTargetNotReached,

			expectedAmountOneIn:    amountZeroTargetNotReached,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			sqrtPriceCurrent:       sqrtPriceCurrent,
			sqrtPriceTarget:        sqrtPriceNext,
			liquidity:              defaultLiquidity,
			amountZeroOutRemaining: defaultAmountZeroOfz.Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:                defaultFee,

			expectedSqrtPriceNext:         sqrtPriceNext,
			expectedAmountZeroOutConsumed: defaultAmountZeroOfz,
			expectedAmountOneIn:           defaultAmountOneOfz.Ceil(),
			expectedFeeChargeTotal:        defaultAmountOneOfz.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			sqrtPriceCurrent:       sqrtPriceCurrent,
			sqrtPriceTarget:        sqrtPriceNext,
			liquidity:              defaultLiquidity,
			amountZeroOutRemaining: defaultAmountZeroOfz.Sub(sdk.NewDec(100)),
			swapFee:                defaultFee,

			expectedSqrtPriceNext:         sqrtPriceTargetNotReached,
			expectedAmountZeroOutConsumed: amountZeroTargetNotReached,
			expectedAmountOneIn:           amountZeroTargetNotReached,
			expectedFeeChargeTotal:        amountZeroTargetNotReached.Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			strategy := swapstrategy.New(false, types.MaxSqrtRatio, suite.App.GetKey(types.ModuleName), tc.swapFee)

			sqrtPriceNext, amountZeroOutConsumed, amountOneIn, feeChargeTotal := strategy.ComputeSwapStepInGivenOut(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountZeroOutRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountZeroOutConsumed, amountZeroOutConsumed)
			suite.Require().Equal(0,
				tokenOutErrTolerance.CompareBigDec(
					osmomath.BigDecFromSDKDec(tc.expectedAmountOneIn),
					osmomath.BigDecFromSDKDec(amountOneIn),
				),
				fmt.Sprintf("expected (%s), actual (%s)", tc.expectedAmountOneIn, amountOneIn))

			suite.Require().Equal(0,
				smallestErrTolerance.CompareBigDec(
					osmomath.BigDecFromSDKDec(tc.expectedFeeChargeTotal),
					osmomath.BigDecFromSDKDec(feeChargeTotal),
				),
				fmt.Sprintf("expected (%s), actual (%s)", tc.expectedFeeChargeTotal, feeChargeTotal))
		})
	}
}
