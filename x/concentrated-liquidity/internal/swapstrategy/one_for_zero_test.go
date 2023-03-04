package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
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
	var (
		sqrtPriceCurrent = defaultSqrtPriceLower
		sqrtPriceNext    = defaultSqrtPriceUpper

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
			amountOneInRemaining: defaultAmountOne.Add(sdk.NewDec(100)),
			swapFee:              sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Reached target, so 100 is not consumed.
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:      defaultAmountZero,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(sdk.NewDec(100)),
			swapFee:              sdk.ZeroDec(),

			// sqrt_price_current + token_in / liquidity
			expectedSqrtPriceNext:    sqrtPriceTargetNotReached,
			expectedAmountInConsumed: defaultAmountOne.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOut:        amountZeroTargetNotReached,
			expectedFeeChargeTotal:   sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Add(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:              defaultFee,

			expectedSqrtPriceNext:    sqrtPriceNext,
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:      defaultAmountZero,
			expectedFeeChargeTotal: defaultAmountOne.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:              defaultFee,

			expectedSqrtPriceNext:    sqrtPriceTargetNotReached,
			expectedAmountInConsumed: defaultAmountOne.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOut:        amountZeroTargetNotReached,
			// Difference between given amount remaining in and amount in actually consumed which qpproximately equals to fee.
			expectedFeeChargeTotal: defaultAmountOne.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)).Sub(defaultAmountOne.Sub(sdk.NewDec(100)).Ceil()),
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
			suite.Require().Equal(tc.expectedAmountOut, amountZeroOut)
			suite.Require().Equal(tc.expectedFeeChargeTotal, feeChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_OneForZero() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceLower
		sqrtPriceNext    = defaultSqrtPriceUpper
		// Target is not reached means that we stop at the sqrt price earlier
		// than expected. As a result, we recalculate the amount out and amount in
		// necessary to reach the earlier target.
		// sqrt_next = liq * sqrt_cur / (liq  - token_out * sqrt_cur) quo round up
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.709031125539448610")
		// liq * (sqrt_next - sqrt_cur)
		amountOneTargetNotReached = sdk.MustNewDecFromStr("61829304.427824073089251659")
		// N.B.: approx eq = defaultAmountZero.Sub(sdk.NewDec(1000))
		// slight variance due to recomputing amount out when target is not reached.
		// liq * (sqrt_next - sqrt_cur) / (sqrt_next * sqrt_cur)
		amountZeroTargetNotReached = sdk.MustNewDecFromStr("12369.999999999999293322")
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
			amountZeroOutRemaining: defaultAmountZero.Add(sdk.NewDec(100)),
			swapFee:                sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Reached target, so 100 is not consumed.
			expectedAmountZeroOutConsumed: defaultAmountZero,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneIn:    defaultAmountOne.Ceil(),
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			sqrtPriceCurrent:       sqrtPriceCurrent,
			sqrtPriceTarget:        sqrtPriceNext,
			liquidity:              defaultLiquidity,
			amountZeroOutRemaining: defaultAmountZero.Sub(sdk.NewDec(1000)),
			swapFee:                sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,

			expectedAmountZeroOutConsumed: amountZeroTargetNotReached,

			expectedAmountOneIn:    amountOneTargetNotReached.Ceil(),
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			sqrtPriceCurrent:       sqrtPriceCurrent,
			sqrtPriceTarget:        sqrtPriceNext,
			liquidity:              defaultLiquidity,
			amountZeroOutRemaining: defaultAmountZero.Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:                defaultFee,

			expectedSqrtPriceNext:         sqrtPriceNext,
			expectedAmountZeroOutConsumed: defaultAmountZero,
			expectedAmountOneIn:           defaultAmountOne.Ceil(),
			expectedFeeChargeTotal:        defaultAmountOne.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			sqrtPriceCurrent:       sqrtPriceCurrent,
			sqrtPriceTarget:        sqrtPriceNext,
			liquidity:              defaultLiquidity,
			amountZeroOutRemaining: defaultAmountZero.Sub(sdk.NewDec(1000)),
			swapFee:                defaultFee,

			expectedSqrtPriceNext:         sqrtPriceTargetNotReached,
			expectedAmountZeroOutConsumed: amountZeroTargetNotReached,
			expectedAmountOneIn:           amountOneTargetNotReached.Ceil(),
			expectedFeeChargeTotal:        amountOneTargetNotReached.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
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
			suite.Require().Equal(tc.expectedAmountOneIn, amountOneIn)
			suite.Require().Equal(tc.expectedFeeChargeTotal, feeChargeTotal)
		})
	}
}
