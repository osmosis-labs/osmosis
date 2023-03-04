package swapstrategy_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var (
	defaultSqrtPriceLowerZfo = sdk.MustNewDecFromStr("70.688664163408836321") // approx 4996.89
	defaultSqrtPriceUpperZfo = sdk.MustNewDecFromStr("70.710678118654752440") // 5000
	defaultAmountOneZfo      = sdk.MustNewDecFromStr("66829187.967824033199646915")
	defaultAmountZeroZfo     = sdk.MustNewDecFromStr("13370")
)

func (suite *StrategyTestSuite) TestGetSqrtTargetPrice_ZeroForOne() {
	var (
		two   = sdk.NewDec(2)
		three = sdk.NewDec(2)
		four  = sdk.NewDec(4)
		five  = sdk.NewDec(5)
	)

	tests := map[string]struct {
		isZeroForOne      bool
		sqrtPriceLimit    sdk.Dec
		nextTickSqrtPrice sdk.Dec
		expectedResult    sdk.Dec
	}{
		"nextTickSqrtPrice == sqrtPriceLimit -> returns either": {
			sqrtPriceLimit:    sdk.OneDec(),
			nextTickSqrtPrice: sdk.OneDec(),
			expectedResult:    sdk.OneDec(),
		},
		"nextTickSqrtPrice > sqrtPriceLimit -> nextTickSqrtPrice": {
			sqrtPriceLimit:    three,
			nextTickSqrtPrice: four,
			expectedResult:    four,
		},
		"nextTickSqrtPrice < sqrtPriceLimit -> sqrtPriceLimit": {
			sqrtPriceLimit:    five,
			nextTickSqrtPrice: two,
			expectedResult:    five,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			sut := swapstrategy.New(true, tc.sqrtPriceLimit, suite.App.GetKey(types.ModuleName), sdk.ZeroDec())

			actualSqrtTargetPrice := sut.GetSqrtTargetPrice(tc.nextTickSqrtPrice)

			suite.Require().Equal(tc.expectedResult, actualSqrtTargetPrice)

		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_ZeroForOne() {
	tokenOutErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.SmallestDec().MulInt64(200),
	}

	var (
		sqrtPriceCurrent = defaultSqrtPriceUpperZfo
		sqrtPriceNext    = defaultSqrtPriceLowerZfo

		// liquidity * sqrtPriceCurrent / (liquidity + amount in * sqrtPriceCurrent)
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.688828764403676328")
		// liquidity * (sqrtPriceCurrent - sqrtPriceNext)
		amountOneTargetNotReached = sdk.MustNewDecFromStr("66329498.080160873004154169")
	)

	tests := map[string]struct {
		sqrtPriceCurrent      sdk.Dec
		sqrtPriceTarget       sdk.Dec
		liquidity             sdk.Dec
		amountZeroInRemaining sdk.Dec
		swapFee               sdk.Dec

		expectedSqrtPriceNext  sdk.Dec
		amountInConsumed       sdk.Dec
		expectedAmountOneOut   sdk.Dec
		expectedFeeChargeTotal sdk.Dec

		expectError error
	}{
		"1: no fee - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// add 100 more
			amountZeroInRemaining: defaultAmountZeroZfo.Add(sdk.NewDec(100)),
			swapFee:               sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// consumed without 100 since reached target.
			amountInConsumed: defaultAmountZeroZfo,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:   defaultAmountOneZfo,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			sqrtPriceCurrent:      sqrtPriceCurrent,
			sqrtPriceTarget:       sqrtPriceNext,
			liquidity:             defaultLiquidity,
			amountZeroInRemaining: defaultAmountZeroZfo.Sub(sdk.NewDec(100)),
			swapFee:               sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			amountInConsumed:      defaultAmountZeroZfo.Sub(sdk.NewDec(100)).Ceil(),

			expectedAmountOneOut:   amountOneTargetNotReached,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// add 100 more
			amountZeroInRemaining: defaultAmountZeroZfo.Add(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:               defaultFee,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reached target and fee is applied.
			amountInConsumed: defaultAmountZeroZfo.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:   defaultAmountOneZfo,
			expectedFeeChargeTotal: defaultAmountZeroZfo.Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			sqrtPriceCurrent:      sqrtPriceCurrent,
			sqrtPriceTarget:       sqrtPriceNext,
			liquidity:             defaultLiquidity,
			amountZeroInRemaining: defaultAmountZeroZfo.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:               defaultFee,

			expectedSqrtPriceNext:  sqrtPriceTargetNotReached,
			amountInConsumed:       defaultAmountZeroZfo.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOneOut:   amountOneTargetNotReached,
			expectedFeeChargeTotal: defaultAmountZeroZfo.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			strategy := swapstrategy.New(true, types.MaxSqrtRatio, suite.App.GetKey(types.ModuleName), tc.swapFee)

			sqrtPriceNext, newAmountRemainingOneIn, amountZeroOut, feeChargeTotal := strategy.ComputeSwapStepOutGivenIn(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountZeroInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.amountInConsumed, newAmountRemainingOneIn)
			suite.Require().Equal(0,
				tokenOutErrTolerance.CompareBigDec(
					osmomath.BigDecFromSDKDec(tc.expectedAmountOneOut),
					osmomath.BigDecFromSDKDec(amountZeroOut),
				),
				fmt.Sprintf("expected (%s), actual (%s)", tc.expectedAmountOneOut, amountZeroOut))
			suite.Require().Equal(tc.expectedFeeChargeTotal, feeChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_ZeroForOne() {
	tokenInErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.SmallestDec().MulInt64(200),
	}
	oneErrTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: sdk.OneDec(),
	}

	var (
		sqrtPriceCurrent = defaultSqrtPriceUpperZfo
		sqrtPriceNext    = defaultSqrtPriceLowerZfo

		// sqrt_cur - amt_one / liq quo round up
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.688667457471792243")
		// ceil(liq * (sqrt_cur - sqrt_next) / (sqrt_cur * sqrt_next))
		amountZeroTargetNotReached = sdk.MustNewDecFromStr("13367.998754214115430370").Ceil()
	)

	tests := map[string]struct {
		sqrtPriceCurrent      sdk.Dec
		sqrtPriceTarget       sdk.Dec
		liquidity             sdk.Dec
		amountOneOutRemaining sdk.Dec
		swapFee               sdk.Dec

		expectedSqrtPriceNext  sdk.Dec
		amountOutConsumed      sdk.Dec
		expectedAmountInZero   sdk.Dec
		expectedFeeChargeTotal sdk.Dec

		expectError error
	}{
		"1: no fee - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountOneOutRemaining: defaultAmountOneZfo.Add(sdk.NewDec(100)),
			swapFee:               sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOutConsumed:      defaultAmountOneZfo,
			expectedAmountInZero:   defaultAmountZeroZfo.Ceil(),
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			sqrtPriceCurrent:      sqrtPriceCurrent,
			sqrtPriceTarget:       sqrtPriceNext,
			liquidity:             defaultLiquidity,
			amountOneOutRemaining: defaultAmountOneZfo.Sub(sdk.NewDec(10000)),
			swapFee:               sdk.ZeroDec(),

			// sqrt_cur - amt_one / liq quo round up
			expectedSqrtPriceNext:  sqrtPriceTargetNotReached,
			amountOutConsumed:      defaultAmountOneZfo.Sub(sdk.NewDec(10000)),
			expectedAmountInZero:   amountZeroTargetNotReached.Ceil(),
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountOneOutRemaining: defaultAmountOneZfo.Quo(sdk.OneDec().Sub(defaultFee)),
			swapFee:               defaultFee,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOutConsumed: defaultAmountOneZfo,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountInZero:   defaultAmountZeroZfo.Ceil(),
			expectedFeeChargeTotal: defaultAmountZeroZfo.Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			sqrtPriceCurrent:      sqrtPriceCurrent,
			sqrtPriceTarget:       sqrtPriceNext,
			liquidity:             defaultLiquidity,
			amountOneOutRemaining: defaultAmountOneZfo.Sub(sdk.NewDec(10000)),
			swapFee:               defaultFee,

			expectedSqrtPriceNext:  sqrtPriceTargetNotReached,
			amountOutConsumed:      defaultAmountOneZfo.Sub(sdk.NewDec(10000)),
			expectedAmountInZero:   amountZeroTargetNotReached.Ceil(),
			expectedFeeChargeTotal: amountZeroTargetNotReached.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			strategy := swapstrategy.New(true, types.MaxSqrtRatio, suite.App.GetKey(types.ModuleName), tc.swapFee)

			sqrtPriceNext, newAmountRemainingZeroOut, amountOneIn, feeChargeTotal := strategy.ComputeSwapStepInGivenOut(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountOneOutRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(0,
				oneErrTolerance.CompareBigDec(
					osmomath.BigDecFromSDKDec(tc.amountOutConsumed),
					osmomath.BigDecFromSDKDec(newAmountRemainingZeroOut),
				),
				fmt.Sprintf("expected (%s), actual (%s)", tc.amountOutConsumed, newAmountRemainingZeroOut))
			suite.Require().Equal(0,
				tokenInErrTolerance.CompareBigDec(
					osmomath.BigDecFromSDKDec(tc.expectedAmountInZero),
					osmomath.BigDecFromSDKDec(amountOneIn),
				),
				fmt.Sprintf("expected (%s), actual (%s)", tc.expectedAmountInZero, amountOneIn))
			suite.Require().Equal(tc.expectedFeeChargeTotal, feeChargeTotal)
		})
	}
}
