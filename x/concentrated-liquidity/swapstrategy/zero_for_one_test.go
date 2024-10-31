package swapstrategy_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

var hundredDec = osmomath.NewDec(100)
var oneMinusDefaultSpreadFactor = one.Sub(defaultSpreadReward)

func (suite *StrategyTestSuite) setupNewZeroForOneSwapStrategy(sqrtPriceLimit osmomath.Dec, spread osmomath.Dec) swapstrategy.SwapStrategy {
	suite.SetupTest()
	return swapstrategy.New(true, osmomath.BigDecFromDec(sqrtPriceLimit), suite.App.GetKey(types.ModuleName), spread)
}

func (suite *StrategyTestSuite) TestGetSqrtTargetPrice_ZeroForOne() {
	tests := map[string]struct {
		isZeroForOne      bool
		sqrtPriceLimit    osmomath.Dec
		nextTickSqrtPrice osmomath.Dec
		expectedResult    osmomath.Dec
	}{
		"nextTickSqrtPrice == sqrtPriceLimit -> returns either": {
			sqrtPriceLimit:    osmomath.OneDec(),
			nextTickSqrtPrice: osmomath.OneDec(),
			expectedResult:    osmomath.OneDec(),
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
			actualSqrtTargetPrice := sut.GetSqrtTargetPrice(osmomath.BigDecFromDec(tc.nextTickSqrtPrice))
			suite.Require().Equal(osmomath.BigDecFromDec(tc.expectedResult), actualSqrtTargetPrice)
		})
	}
}

// Estimates are computed using x/concentrated-liquidity/python/clmath.py
func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_ZeroForOne() {
	var (
		sqrtPriceCurrent = osmomath.BigDecFromDec(defaultSqrtPriceUpper)
		sqrtPriceNext    = osmomath.BigDecFromDec(defaultSqrtPriceLower)
		sqrtPriceTarget  = sqrtPriceNext.Dec()

		// get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, tokenIn)
		sqrtPriceTargetNotReached = osmomath.MustNewBigDecFromStr("70.688828764403676329447109466419851492")
		// round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
		amountOneTargetNotReached = osmomath.MustNewDecFromStr("66329498.080160868611070352")

		defaultAmountZeroMinus100 = defaultAmountZero.Sub(hundredDec)
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		sqrtPriceCurrent osmomath.BigDec
		sqrtPriceTarget  osmomath.Dec
		liquidity        osmomath.Dec

		amountZeroInRemaining osmomath.Dec
		spreadFactor          osmomath.Dec

		expectedSqrtPriceNext           osmomath.BigDec
		amountZeroInConsumed            osmomath.Dec
		expectedAmountOneOut            osmomath.Dec
		expectedSpreadRewardChargeTotal osmomath.Dec
	}{
		"1: no spread reward - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			// add 100 more
			amountZeroInRemaining: defaultAmountZero.Add(hundredDec),
			spreadFactor:          osmomath.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// consumed without 100 since reached target.
			amountZeroInConsumed: defaultAmountZero.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:            defaultAmountOne,
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"2: no spread reward - do not reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			amountZeroInRemaining: defaultAmountZero.Sub(hundredDec),
			spreadFactor:          osmomath.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			amountZeroInConsumed:  defaultAmountZero.Sub(hundredDec).Ceil(),

			expectedAmountOneOut:            amountOneTargetNotReached,
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"3: 3% spread reward - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			// add 100 more
			amountZeroInRemaining: defaultAmountZero.Add(hundredDec).Quo(oneMinusDefaultSpreadFactor),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reached target and spread reward is applied.
			amountZeroInConsumed: defaultAmountZero.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneOut:            defaultAmountOne,
			expectedSpreadRewardChargeTotal: defaultAmountZero.Ceil().Mul(defaultSpreadReward.Quo(oneMinusDefaultSpreadFactor)),
		},
		"4: 3% spread reward - do not reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			amountZeroInRemaining: defaultAmountZeroMinus100.Quo(oneMinusDefaultSpreadFactor),
			spreadFactor:          defaultSpreadReward,

			// tokenIn = Decimal("13269.999999999998920002290000000000000000")
			// sqrtPriceNext = get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, tokenIn)
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("70.688828764403676329447108989075854947"),
			amountZeroInConsumed:  defaultAmountZeroMinus100.Ceil(),
			// round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
			expectedAmountOneOut: osmomath.MustNewDecFromStr("66329498.080160868611071801"),
			// Difference between amount in given and actually consumed.
			expectedSpreadRewardChargeTotal: defaultAmountZeroMinus100.Quo(oneMinusDefaultSpreadFactor).Sub(defaultAmountZeroMinus100.Ceil()),
		},
		"5: sub osmomath.Dec ULP precision movement. Supported by osmomath.BigDec ULP": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("0.000001000049998751"),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("0.000001000049998750"),
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroInRemaining: osmomath.NewDec(99),
			spreadFactor:          osmomath.ZeroDec(),

			// sqrtPriceNext = get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, 99)
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("0.000001000049998750999999999999009926"),

			amountZeroInConsumed: osmomath.NewDec(99),
			// round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
			expectedAmountOneOut:            osmomath.MustNewDecFromStr("0.000000000099009873"),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		// If such swap leads to an infinite loop in swap logic, it should be detected and failed. We have such logic implemented
		// If such swap leads to amounts consumed being non-zero, the swap should be failed for security purposes. For example,
		// the risk is amountIn being zero but amountOut being non-zero, leading to exploitable behavior. We have logic to prevent
		// this and fail the swap.
		"6: sub osmomath.BigDec ULP precision movement. Nothing charged for amountIn due to precision loss. No amounts consumed": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("0.000001000049998751"),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("0.000001000049998750"),
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroInRemaining: osmomath.SmallestDec(),
			spreadFactor:          osmomath.ZeroDec(),

			// sqrtPriceNext = get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, 99)
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("0.000001000049998751"),

			amountZeroInConsumed: osmomath.ZeroDec(),
			// round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
			expectedAmountOneOut:            osmomath.ZeroDec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"7: precisely osmomath.BigDec ULP precision movement. Amounts in and out are consumed.": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("0.000001000049998751"),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("0.000001000049998750"),
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroInRemaining: osmomath.SmallestDec().MulInt64(100000000000000),
			spreadFactor:          osmomath.ZeroDec(),

			// sqrtPriceNext = get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, oneULPDec * 100000000000000)
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("0.000001000049998750999999999999999999"),

			// ceil(calc_amount_zero_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
			// Note that amount consumed ends up being greater than amount remaining. After subtracting the amount consumed
			// from amount remaining at the end of the swap, we might end up with negative amount in remaining. This is acceptable
			// As this then gets subtracted from the minimum amount in given by user and rounded up. Therefore, in the worst case,
			// we upcharge the user by 1 unit due to rounding.
			amountZeroInConsumed: osmomath.OneDec(),
			// round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
			expectedAmountOneOut:            osmomath.SmallestDec().MulInt64(100),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountZeroIn, amountOneOut, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketOutGivenIn(tc.sqrtPriceCurrent, osmomath.BigDecFromDec(tc.sqrtPriceTarget), tc.liquidity, tc.amountZeroInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountOneOut, amountOneOut)
			suite.Require().Equal(tc.amountZeroInConsumed, amountZeroIn)
			suite.Require().Equal(tc.expectedSpreadRewardChargeTotal, spreadRewardChargeTotal)
		})
	}
}

// Estimates are computed using x/concentrated-liquidity/python/clmath.py
func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_ZeroForOne() {
	var (
		sqrtPriceCurrent = osmomath.BigDecFromDec(defaultSqrtPriceUpper)
		sqrtPriceNext    = osmomath.BigDecFromDec(defaultSqrtPriceLower)
		sqrtPriceTarget  = sqrtPriceNext.Dec()

		// get_next_sqrt_price_from_amount1_out_round_down(liquidity, sqrtPriceCurrent, tokenOut)
		sqrtPriceTargetNotReached = osmomath.MustNewBigDecFromStr("70.688667457471792243056846000067005485")
		// round_sdk_prec_up(calc_amount_zero_delta(liquidity, sqrtPriceCurrent, Decimal("70.688667457471792243056846000067005485"), True))
		amountZeroTargetNotReached = osmomath.MustNewDecFromStr("13367.998754214114788303")

		// N.B.: approx eq = defaultAmountOneZfo.Sub(osmomath.NewDec(10000))
		// slight variance due to recomputing amount out when target is not reached.
		// liq * (sqrt_cur - sqrt_next)
		// round_sdk_prec_down(calc_amount_one_delta(liquidity, sqrtPriceCurrent, Decimal("70.688667457471792243056846000067005485"), False))
		amountOneOutTargetNotReached = osmomath.MustNewDecFromStr("66819187.967824033199646915")
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		sqrtPriceCurrent osmomath.BigDec
		sqrtPriceTarget  osmomath.Dec
		liquidity        osmomath.Dec

		amountOneOutRemaining osmomath.Dec
		spreadFactor          osmomath.Dec

		expectedSqrtPriceNext           osmomath.BigDec
		amountOneOutConsumed            osmomath.Dec
		expectedAmountInZero            osmomath.Dec
		expectedSpreadRewardChargeTotal osmomath.Dec
	}{
		"1: no spread reward - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			// Add 100.
			amountOneOutRemaining: defaultAmountOne.Add(hundredDec),
			spreadFactor:          zero,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOneOutConsumed:            defaultAmountOne,
			expectedAmountInZero:            defaultAmountZero.Ceil(),
			expectedSpreadRewardChargeTotal: zero,
		},
		"2: no spread reward - do not reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			amountOneOutRemaining: defaultAmountOne.Sub(osmomath.NewDec(10000)),
			spreadFactor:          zero,

			// sqrt_cur - amt_one / liq quo round up
			expectedSqrtPriceNext:           sqrtPriceTargetNotReached,
			amountOneOutConsumed:            amountOneOutTargetNotReached,
			expectedAmountInZero:            amountZeroTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: zero,
		},
		"3: 3% spread reward - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			// Add 100.
			amountOneOutRemaining: defaultAmountOne.Quo(oneMinusDefaultSpreadFactor),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceNext,
			// Consumes without 100 since reaches target.
			amountOneOutConsumed: defaultAmountOne,
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountInZero:            defaultAmountZero.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountZero.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread reward - do not reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceTarget,
			liquidity:        defaultLiquidity,

			amountOneOutRemaining: defaultAmountOne.Sub(osmomath.NewDec(10000)),
			spreadFactor:          defaultSpreadReward,

			expectedSqrtPriceNext:           sqrtPriceTargetNotReached,
			amountOneOutConsumed:            amountOneOutTargetNotReached,
			expectedAmountInZero:            amountZeroTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(amountZeroTargetNotReached.Ceil(), defaultSpreadReward),
		},
		"5: osmomath.Dec ULP swap out, within moves sqrt price by 1 osmomath.BigDec ULP, amountOut consumed is greater than remaining": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("0.000001000049998751"),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("0.000001000049998750"),
			// Chosen to be large with the goal of making sqrt price next be equal to sqrt price current.
			// This is due to the fact that sqrtPriceNext = sqrtPriceCurrent - tokenOut / liquidity (quo round up).
			liquidity: osmomath.MustNewDecFromStr("10000000000000000000.937822606808718081"),

			// Chosen to be small with the goal of making sqrt price next be equal to sqrt price current.
			// This is due to the fact that sqrtPriceNext = sqrtPriceCurrent - tokenOut / liquidity (quo round up).
			amountOneOutRemaining: osmomath.SmallestDec(),
			spreadFactor:          osmomath.ZeroDec(),

			// sqrtPriceNext = get_next_sqrt_price_from_amount1_out_round_down(liquidity, sqrtPriceCurrent, oneULPDec)
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("0.000001000049998750999999999999999999"),

			// round_sdk_prec_up(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, False))
			// Results in 0.000000000000000010. However, notice that this value is greater than amountRemaining.
			// Therefore, the amountOut consumed gets reset to amountOutRemaining.
			// See code comments in ComputeSwapWithinBucketInGivenOut(...)
			amountOneOutConsumed: osmomath.SmallestDec(),
			// round_sdk_prec_down(calc_amount_zero_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True))
			expectedAmountInZero:            osmomath.MustNewDecFromStr("0.000099992498812332").Ceil(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewZeroForOneSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountOneOut, amountZeroIn, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketInGivenOut(tc.sqrtPriceCurrent, osmomath.BigDecFromDec(tc.sqrtPriceTarget), tc.liquidity, tc.amountOneOutRemaining)

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
