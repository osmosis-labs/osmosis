package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

func (suite *StrategyTestSuite) setupNewZeroForOneSwapStrategy(sqrtPriceLimit sdk.Dec, spread sdk.Dec) swapstrategy.SwapStrategy {
	suite.SetupTest()
	return swapstrategy.New(true, sqrtPriceLimit, suite.App.GetKey(types.ModuleName), spread, defaultTickSpacing)
}

func (suite *StrategyTestSuite) TestGetSqrtTargetPrice_ZeroForOne() {
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
			sut := suite.setupNewZeroForOneSwapStrategy(tc.sqrtPriceLimit, zero)
			actualSqrtTargetPrice := sut.GetSqrtTargetPrice(tc.nextTickSqrtPrice)
			suite.Require().Equal(tc.expectedResult, actualSqrtTargetPrice)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_ZeroForOne() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceUpper
		sqrtPriceNext    = defaultSqrtPriceLower
		sqrtPriceTarget  = sqrtPriceNext

		// liquidity * sqrtPriceCurrent / (liquidity + amount in * sqrtPriceCurrent)
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.688828764403676330")
		// liquidity * (sqrtPriceCurrent - sqrtPriceNext)
		amountOneTargetNotReached = sdk.MustNewDecFromStr("66329498.080160866932624794")
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		amountZeroInRemaining sdk.Dec
		spreadFactor          sdk.Dec

		expectedSqrtPriceNext  sdk.Dec
		amountZeroInConsumed   sdk.Dec
		expectedAmountOneOut   sdk.Dec
		expectedFeeChargeTotal sdk.Dec

		expectError error
	}{
		"1: no fee - reach target": {
			// add 100 more
			amountZeroInRemaining: defaultAmountZero.Add(sdk.NewDec(100)),
			spreadFactor:          sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// consumed without 100 since reached target.
			amountZeroInConsumed: defaultAmountZero.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:   defaultAmountOne,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"2: no fee - do not reach target": {
			amountZeroInRemaining: defaultAmountZero.Sub(sdk.NewDec(100)),
			spreadFactor:          sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			amountZeroInConsumed:  defaultAmountZero.Sub(sdk.NewDec(100)).Ceil(),

			expectedAmountOneOut:   amountOneTargetNotReached,
			expectedFeeChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% fee - reach target": {
			// add 100 more
			amountZeroInRemaining: defaultAmountZero.Add(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			spreadFactor:          defaultFee,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reached target and fee is applied.
			amountZeroInConsumed: defaultAmountZero.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:   defaultAmountOne,
			expectedFeeChargeTotal: defaultAmountZero.Ceil().Quo(sdk.OneDec().Sub(defaultFee)).Mul(defaultFee),
		},
		"4: 3% fee - do not reach target": {
			amountZeroInRemaining: defaultAmountZero.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)),
			spreadFactor:          defaultFee,

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			amountZeroInConsumed:  defaultAmountZero.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOneOut:  amountOneTargetNotReached,
			// Difference between amount in given and actually consumed.
			expectedFeeChargeTotal: defaultAmountZero.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultFee)).Sub(defaultAmountZero.Sub(sdk.NewDec(100)).Ceil()),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountZeroIn, amountOneOut, feeChargeTotal := strategy.ComputeSwapStepOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget, defaultLiquidity, tc.amountZeroInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.amountZeroInConsumed, amountZeroIn)
			suite.Require().Equal(tc.expectedAmountOneOut, amountOneOut)
			suite.Require().Equal(tc.expectedFeeChargeTotal, feeChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_ZeroForOne() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceUpper
		sqrtPriceNext    = defaultSqrtPriceLower
		sqrtPriceTarget  = sqrtPriceNext

		// sqrt_cur - amt_one / liq quo round up
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.688667457471792243")
		// liq * (sqrt_cur - sqrt_next) / (sqrt_cur * sqrt_next)
		amountZeroTargetNotReached = sdk.MustNewDecFromStr("13367.998754214115430370")

		// N.B.: approx eq = defaultAmountOneZfo.Sub(sdk.NewDec(10000))
		// slight variance due to recomputing amount out when target is not reached.
		// liq * (sqrt_cur - sqrt_next)
		amountOneOutTargetNotReached = sdk.MustNewDecFromStr("66819187.967824033372217995")
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		amountOneOutRemaining sdk.Dec
		spreadFactor          sdk.Dec

		expectedSqrtPriceNext  sdk.Dec
		amountOneOutConsumed   sdk.Dec
		expectedAmountInZero   sdk.Dec
		expectedFeeChargeTotal sdk.Dec

		expectError error
	}{
		"1: no fee - reach target": {
			// Add 100.
			amountOneOutRemaining: defaultAmountOne.Add(sdk.NewDec(100)),
			spreadFactor:          zero,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOneOutConsumed:   defaultAmountOne,
			expectedAmountInZero:   defaultAmountZero.Ceil(),
			expectedFeeChargeTotal: zero,
		},
		"2: no fee - do not reach target": {
			amountOneOutRemaining: defaultAmountOne.Sub(sdk.NewDec(10000)),
			spreadFactor:          zero,

			// sqrt_cur - amt_one / liq quo round up
			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			// subtracting 1 * smallest dec to account for truncations in favor of the pool.
			amountOneOutConsumed:   amountOneOutTargetNotReached.Sub(sdk.SmallestDec()),
			expectedAmountInZero:   amountZeroTargetNotReached.Ceil(),
			expectedFeeChargeTotal: zero,
		},
		"3: 3% fee - reach target": {
			// Add 100.
			amountOneOutRemaining: defaultAmountOne.Quo(one.Sub(defaultFee)),
			spreadFactor:          defaultFee,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOneOutConsumed: defaultAmountOne,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountInZero:   defaultAmountZero.Ceil(),
			expectedFeeChargeTotal: swapstrategy.ComputeFeeChargeFromAmountIn(defaultAmountZero.Ceil(), defaultFee),
		},
		"4: 3% fee - do not reach target": {
			amountOneOutRemaining: defaultAmountOne.Sub(sdk.NewDec(10000)),
			spreadFactor:          defaultFee,

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			// subtracting 1 * smallest dec to account for truncations in favor of the pool.
			amountOneOutConsumed:   amountOneOutTargetNotReached.Sub(sdk.SmallestDec()),
			expectedAmountInZero:   amountZeroTargetNotReached.Ceil(),
			expectedFeeChargeTotal: swapstrategy.ComputeFeeChargeFromAmountIn(amountZeroTargetNotReached.Ceil(), defaultFee),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountOneOut, amountZeroIn, feeChargeTotal := strategy.ComputeSwapStepInGivenOut(sqrtPriceCurrent, sqrtPriceTarget, defaultLiquidity, tc.amountOneOutRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.amountOneOutConsumed, amountOneOut)
			suite.Require().Equal(tc.expectedAmountInZero, amountZeroIn)
			suite.Require().Equal(tc.expectedFeeChargeTotal.String(), feeChargeTotal.String())
		})
	}
}

func (suite *StrategyTestSuite) TestInitializeNextTickIterator_ZeroForOne() {
	tests := map[string]struct {
		currentTick     int64
		preSetPositions []position

		expectIsValid  bool
		expectNextTick int64
		expectError    error
	}{
		"1 position, zero for one": {
			preSetPositions: []position{
				{
					lowerTick: -100,
					upperTick: 100,
				},
			},
			expectIsValid:  true,
			expectNextTick: -100,
		},
		"2 positions, zero for one": {
			preSetPositions: []position{
				{
					lowerTick: -400,
					upperTick: 300,
				},
				{
					lowerTick: -200,
					upperTick: 200,
				},
			},
			expectIsValid:  true,
			expectNextTick: -200,
		},
		"lower tick lands on current tick, zero for one": {
			preSetPositions: []position{
				{
					lowerTick: 0,
					upperTick: 100,
				},
			},
			expectIsValid:  true,
			expectNextTick: 0,
		},
		"upper tick lands on current tick, zero for one": {
			preSetPositions: []position{
				{
					lowerTick: -100,
					upperTick: 0,
				},
			},
			expectIsValid:  true,
			expectNextTick: 0,
		},
		"no ticks, zero for one": {
			expectIsValid: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, zero)
			pool := suite.PrepareConcentratedPool()
			suite.setupPresetPositions(pool.GetId(), tc.preSetPositions)

			// refetch pool
			pool, err := suite.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(suite.Ctx, pool.GetId())
			suite.Require().NoError(err)

			currentTick := pool.GetCurrentTick()
			suite.Require().Equal(int64(0), currentTick)

			tickIndex := strategy.InitializeTickValue(currentTick)

			iter := strategy.InitializeNextTickIterator(suite.Ctx, defaultPoolId, tickIndex)
			defer iter.Close()

			suite.Require().Equal(tc.expectIsValid, iter.Valid())
			if tc.expectIsValid {
				actualNextTick, err := types.TickIndexFromBytes(iter.Key())
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectNextTick, actualNextTick)
			}
		})
	}
}
