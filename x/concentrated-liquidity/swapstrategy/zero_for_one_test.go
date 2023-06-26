package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

func (suite *StrategyTestSuite) setupNewZeroForOneSwapStrategy(sqrtPriceLimit sdk.Dec, spread sdk.Dec) swapstrategy.SwapStrategy {
	suite.SetupTest()
	return swapstrategy.New(true, sqrtPriceLimit, suite.App.GetKey(types.ModuleName), spread)
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
		sqrtPriceNext = defaultSqrtPriceLower

		// liquidity * sqrtPriceCurrent / (liquidity + amount in * sqrtPriceCurrent)
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.688828764403676330")
		// liquidity * (sqrtPriceCurrent - sqrtPriceNext)
		amountOneTargetNotReached = sdk.MustNewDecFromStr("66329498.080160866932624794")
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		sqrtPriceCurrent sdk.Dec
		sqrtPriceTarget  sdk.Dec
		liquidity        sdk.Dec

		amountZeroInRemaining sdk.Dec
		spreadFactor          sdk.Dec

		expectedSqrtPriceNext           sdk.Dec
		amountZeroInConsumed            sdk.Dec
		expectedAmountOneOut            sdk.Dec
		expectedSpreadRewardChargeTotal sdk.Dec

		expectError error
	}{
		"1: no spread reward - reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  defaultSqrtPriceLower,
			liquidity:        defaultLiquidity,

			// add 100 more
			amountZeroInRemaining: defaultAmountZero.Add(sdk.NewDec(100)),
			spreadFactor:          sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// consumed without 100 since reached target.
			amountZeroInConsumed: defaultAmountZero.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:            defaultAmountOne,
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"2: no spread reward - do not reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  defaultSqrtPriceLower,
			liquidity:        defaultLiquidity,

			amountZeroInRemaining: defaultAmountZero.Sub(sdk.NewDec(100)),
			spreadFactor:          sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			amountZeroInConsumed:  defaultAmountZero.Sub(sdk.NewDec(100)).Ceil(),

			expectedAmountOneOut:            amountOneTargetNotReached,
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% spread reward - reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  defaultSqrtPriceLower,
			liquidity:        defaultLiquidity,

			// add 100 more
			amountZeroInRemaining: defaultAmountZero.Add(sdk.NewDec(100)).Quo(one.Sub(defaultSpreadReward)),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reached target and spread reward is applied.
			amountZeroInConsumed: defaultAmountZero.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:            defaultAmountOne,
			expectedSpreadRewardChargeTotal: defaultAmountZero.Ceil().Quo(one.Sub(defaultSpreadReward)).Mul(defaultSpreadReward),
		},
		"4: 3% spread reward - do not reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  defaultSqrtPriceLower,
			liquidity:        defaultLiquidity,

			amountZeroInRemaining: defaultAmountZero.Sub(sdk.NewDec(100)).Quo(one.Sub(defaultSpreadReward)),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			amountZeroInConsumed:  defaultAmountZero.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOneOut:  amountOneTargetNotReached,
			// Difference between amount in given and actually consumed.
			expectedSpreadRewardChargeTotal: defaultAmountZero.Sub(sdk.NewDec(100)).Quo(one.Sub(defaultSpreadReward)).Sub(defaultAmountZero.Sub(sdk.NewDec(100)).Ceil()),
		},
		"5: invalid zero difference between sqrt price current and sqrt price next due to precision loss, full amount remaining in is charged and amount out calculated from sqrt price": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: sdk.MustNewDecFromStr("0.000001000049998751"),
			sqrtPriceTarget:  sdk.MustNewDecFromStr("0.000001000049998750"),
			liquidity:        sdk.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroInRemaining: sdk.NewDec(99),
			spreadFactor:          sdk.ZeroDec(),

			// computed with x/concentrated-liquidity/python/clmath.py
			// sqrtPriceNext = liquidity * sqrtPriceCurrent / (liquidity + tokenIn * sqrtPriceCurrent)
			// sqrtPriceNext = round_decimal(sqrtPriceNext, 18, ROUND_CEILING)
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("0.000001000049998751"),

			amountZeroInConsumed: sdk.NewDec(99),
			// diff = sqrtPriceCurrent - oneULPDec
			// price = round_decimal(diff * diff, 18, ROUND_FLOOR)
			// round_decimal(amountZeroInRemaining * price, 18, ROUND_FLOOR)
			expectedAmountOneOut:            sdk.MustNewDecFromStr("0.000000000099009801"),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		// Note: was unable to construct a case where the difference between sqrt price current and sqrt price next is zero due to precision loss.
		// Leaving a test case that is as close as possible but gets a non-zero difference after sqrtPriceNext is recomputed with
		// GetNextSqrtPriceFromAmount0InRoundingUp(...)
		"6: no zero difference between sqrt price current and sqrt price next due to precision loss": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: types.MaxSqrtPrice.Sub(sdk.SmallestDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice.Sub(sdk.SmallestDec()).Sub(sdk.SmallestDec()),
			liquidity:        sdk.MustNewDecFromStr("10000249806240159879124189790812462189402487127210.937822606808718081"),

			amountZeroInRemaining: sdk.NewDecWithPrec(5, 1),
			spreadFactor:          sdk.ZeroDec(),

			// computed with x/concentrated-liquidity/python/clmath.py using
			// liquidity * sqrtPriceCurrent / (liquidity + tokenIn * sqrtPriceCurrent)
			// Smallest dec is added presumably from rounding differences. Could not reconsruct this case with
			// not differece.
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("9999999999999999999.999999999995000123").Add(sdk.SmallestDec()),

			// rounded up to 1
			amountZeroInConsumed: sdk.OneDec(),
			// round_decimal(liquidity * (sqrtPriceCurrent - sqrtPriceNext), 18, ROUND_FLOOR)
			// Adding 5 ULP to display the difference from Python calculation. Was unable to construct a 1:1 case or explain the difference.
			expectedAmountOneOut:            sdk.MustNewDecFromStr("49999998999975019375636058430338459389.23876032516378774").Add(sdk.SmallestDec().MulInt64(5)),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountZeroIn, amountOneOut, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketOutGivenIn(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountZeroInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.amountZeroInConsumed, amountZeroIn)
			suite.Require().Equal(tc.expectedAmountOneOut, amountOneOut)
			suite.Require().Equal(tc.expectedSpreadRewardChargeTotal, spreadRewardChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_ZeroForOne() {
	var (
		sqrtPriceNext = defaultSqrtPriceLower

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
		sqrtPriceCurrent sdk.Dec
		sqrtPriceTarget  sdk.Dec
		liquidity        sdk.Dec

		amountOneOutRemaining sdk.Dec
		spreadFactor          sdk.Dec

		expectedSqrtPriceNext           sdk.Dec
		amountOneOutConsumed            sdk.Dec
		expectedAmountInZero            sdk.Dec
		expectedSpreadRewardChargeTotal sdk.Dec

		expectError error
	}{
		"1: no spread reward - reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,

			// Add 100.
			amountOneOutRemaining: defaultAmountOne.Add(sdk.NewDec(100)),
			spreadFactor:          zero,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOneOutConsumed:            defaultAmountOne,
			expectedAmountInZero:            defaultAmountZero.Ceil(),
			expectedSpreadRewardChargeTotal: zero,
		},
		"2: no spread reward - do not reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,

			amountOneOutRemaining: defaultAmountOne.Sub(sdk.NewDec(10000)),
			spreadFactor:          zero,

			// sqrt_cur - amt_one / liq quo round up
			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			// subtracting 1 * smallest dec to account for truncations in favor of the pool.
			amountOneOutConsumed:            amountOneOutTargetNotReached.Sub(sdk.SmallestDec()),
			expectedAmountInZero:            amountZeroTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: zero,
		},
		"3: 3% spread reward - reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,

			// Add 100.
			amountOneOutRemaining: defaultAmountOne.Quo(one.Sub(defaultSpreadReward)),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOneOutConsumed: defaultAmountOne,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountInZero:            defaultAmountZero.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountZero.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread reward - do not reach target": {
			sqrtPriceCurrent: defaultSqrtPriceUpper,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,

			amountOneOutRemaining: defaultAmountOne.Sub(sdk.NewDec(10000)),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			// subtracting 1 * smallest dec to account for truncations in favor of the pool.
			amountOneOutConsumed:            amountOneOutTargetNotReached.Sub(sdk.SmallestDec()),
			expectedAmountInZero:            amountZeroTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(amountZeroTargetNotReached.Ceil(), defaultSpreadReward),
		},
		"5: invalid zero difference between sqrt price current and sqrt price next due to precision loss, full amount remaining in is charged and amount out calculated from difference between current and target": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: sdk.MustNewDecFromStr("0.000001000049998751"),
			sqrtPriceTarget:  sdk.MustNewDecFromStr("0.000001000049998750"),
			// Chosen to be large with the goal of making sqrt price next be equal to sqrt price current.
			// This is due to the fact that sqrtPriceNext = sqrtPriceCurrent - tokenOut / liquidity (quo round up).
			liquidity: sdk.MustNewDecFromStr("10000000000000000000.937822606808718081"),

			// Chosen to be small with the goal of making sqrt price next be equal to sqrt price current.
			// This is due to the fact that sqrtPriceNext = sqrtPriceCurrent - tokenOut / liquidity (quo round up).
			amountOneOutRemaining: sdk.SmallestDec(),
			spreadFactor:          sdk.ZeroDec(),

			// Brute forced to be equal to sqrtPriceCurrent by increasing/decreasing the numbers in the formula:
			// sqrtPriceCurrent - tokenOut / liquidity (quo round up).
			expectedSqrtPriceNext: sdk.MustNewDecFromStr("0.000001000049998751"),

			amountOneOutConsumed: sdk.SmallestDec(),
			// (liquidity * (sqrtPriceCurrent - sqrtPriceTarget)) / (sqrtPriceCurrent * sqrtPriceTarget)
			// math.ceil(round_decimal(liquidity * (sqrtPriceCurrent - sqrtPriceTarget), 18, ROUND_CEILING) / round_decimal(sqrtPriceCurrent * sqrtPriceTarget, 18, ROUND_FLOOR))
			expectedAmountInZero:            sdk.MustNewDecFromStr("9999000099991"),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountOneOut, amountZeroIn, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketInGivenOut(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountOneOutRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.amountOneOutConsumed, amountOneOut)
			suite.Require().Equal(tc.expectedAmountInZero, amountZeroIn)
			suite.Require().Equal(tc.expectedSpreadRewardChargeTotal.String(), spreadRewardChargeTotal.String())
		})
	}
}

func (suite *StrategyTestSuite) TestInitializeNextTickIterator_ZeroForOne() {
	tests := map[string]tickIteratorTest{
		"1 position, zero for one": {
			preSetPositions: []position{
				{
					lowerTick: -100,
					upperTick: 100,
				},
			},
			tickSpacing:    defaultTickSpacing,
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
			tickSpacing:    defaultTickSpacing,
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
			tickSpacing:    defaultTickSpacing,
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
			tickSpacing:    defaultTickSpacing,
			expectIsValid:  true,
			expectNextTick: 0,
		},
		"no ticks, zero for one": {
			tickSpacing:   defaultTickSpacing,
			expectIsValid: false,
		},

		// Non-default tick spacing

		"1 position, 1 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: -1,
					upperTick: 1,
				},
			},
			tickSpacing:    1,
			expectIsValid:  true,
			expectNextTick: -1,
		},
		"2 positions, 1 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: -4,
					upperTick: 3,
				},
				{
					lowerTick: -2,
					upperTick: 2,
				},
			},
			tickSpacing:    1,
			expectIsValid:  true,
			expectNextTick: -2,
		},
		"lower tick lands on current tick, 1 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: -3,
					upperTick: -2,
				},
				{
					lowerTick: 0,
					upperTick: 1,
				},
			},
			tickSpacing:    1,
			expectIsValid:  true,
			expectNextTick: 0,
		},
		"upper tick lands on current tick, 1 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: -1,
					upperTick: 0,
				},
				{
					lowerTick: 1,
					upperTick: 2,
				},
			},
			tickSpacing:    1,
			expectIsValid:  true,
			expectNextTick: 0,
		},

		"sanity check: 1 position, 10 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: -10,
					upperTick: 10,
				},
			},
			tickSpacing:    10,
			expectIsValid:  true,
			expectNextTick: -10,
		},
		"sanity check: 1 position, 1000 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: -1000,
					upperTick: 1000,
				},
			},
			tickSpacing:    1000,
			expectIsValid:  true,
			expectNextTick: -1000,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, zero)
			suite.runTickIteratorTest(strategy, tc)
		})
	}
}
