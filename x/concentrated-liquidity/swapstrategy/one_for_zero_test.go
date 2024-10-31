package swapstrategy_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

func (suite *StrategyTestSuite) setupNewOneForZeroSwapStrategy(sqrtPriceLimit osmomath.Dec, spread osmomath.Dec) swapstrategy.SwapStrategy {
	suite.SetupTest()
	return swapstrategy.New(false, osmomath.BigDecFromDec(sqrtPriceLimit), suite.App.GetKey(types.ModuleName), spread)
}

func (suite *StrategyTestSuite) TestGetSqrtTargetPrice_OneForZero() {
	tests := map[string]struct {
		sqrtPriceLimit    osmomath.Dec
		nextTickSqrtPrice osmomath.Dec
		expectedResult    osmomath.Dec
	}{
		"nextTickSqrtPrice == sqrtPriceLimit -> returns either": {
			sqrtPriceLimit:    osmomath.OneDec(),
			nextTickSqrtPrice: osmomath.OneDec(),
			expectedResult:    osmomath.OneDec(),
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
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(tc.sqrtPriceLimit, zero)
			actualSqrtTargetPrice := strategy.GetSqrtTargetPrice(osmomath.BigDecFromDec(tc.nextTickSqrtPrice))
			suite.Require().Equal(osmomath.BigDecFromDec(tc.expectedResult), actualSqrtTargetPrice)
		})
	}
}

// Note: estimates below are computed using x/concentrated-liquidity/python/clmath.py
func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_OneForZero() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceLower
		sqrtPriceNext    = defaultSqrtPriceUpper

		// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
		actualAmountZeroTargetNotReachedBigDec = osmomath.MustNewBigDecFromStr("13369.979999999989602986240259440383244931")

		sqrt = func(x int64) osmomath.Dec {
			sqrt, _ := osmomath.MonotonicSqrt(osmomath.NewDec(x))
			return sqrt
		}
	)

	tests := map[string]struct {
		// TODO revisit each test case and review values
		sqrtPriceCurrent     osmomath.BigDec
		sqrtPriceTarget      osmomath.Dec
		liquidity            osmomath.Dec
		amountOneInRemaining osmomath.Dec
		spreadFactor         osmomath.Dec

		expectedSqrtPriceNext           osmomath.BigDec
		expectedAmountInConsumed        osmomath.Dec
		expectedAmountOut               osmomath.Dec
		expectedSpreadRewardChargeTotal osmomath.Dec
	}{
		"1: no spread factor - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(sqrtPriceCurrent),
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountOneInRemaining: defaultAmountOne.Add(hundredDec),
			spreadFactor:         osmomath.ZeroDec(),

			expectedSqrtPriceNext: osmomath.BigDecFromDec(sqrtPriceNext),
			// Reached target, so 100 is not consumed.
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:               defaultAmountZeroBigDec.Dec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"2: no spread factor - do not reach target": {
			sqrtPriceCurrent:     osmomath.BigDecFromDec(sqrtPriceCurrent),
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(hundredDec),
			spreadFactor:         osmomath.ZeroDec(),

			// sqrtPriceCurrent + round_osmo_prec_down(token_in / liquidity)
			// sqrtPriceCurrent + token_in / liquidity
			expectedSqrtPriceNext:           osmomath.MustNewBigDecFromStr("70.710678085714122880779431539932994712"),
			expectedAmountInConsumed:        defaultAmountOne.Sub(hundredDec).Ceil(),
			expectedAmountOut:               actualAmountZeroTargetNotReachedBigDec.Dec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"3: 3% spread factor - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(sqrtPriceCurrent),
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,

			amountOneInRemaining:     defaultAmountOne.Add(hundredDec).Quo(oneMinusDefaultSpreadFactor),
			spreadFactor:             defaultSpreadReward,
			expectedSqrtPriceNext:    osmomath.BigDecFromDec(sqrtPriceNext),
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:               defaultAmountZeroBigDec.Dec(), // subtracting smallest dec to account for truncations in favor of the pool.
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountOne.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread factor - do not reach target": {
			sqrtPriceCurrent:     osmomath.BigDecFromDec(sqrtPriceCurrent),
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(hundredDec).QuoRoundUp(oneMinusDefaultSpreadFactor),
			spreadFactor:         defaultSpreadReward,

			// sqrtPriceCurrent + round_osmo_prec_down(round_osmo_prec_down(round_sdk_prec_up(token_in / (1 - spreadFactor )) * (1 - spreadFactor)) / liquidity)
			expectedSqrtPriceNext:    osmomath.MustNewBigDecFromStr("70.710678085714122880779431540005464097"),
			expectedAmountInConsumed: defaultAmountOne.Sub(hundredDec).Ceil(),
			expectedAmountOut:        actualAmountZeroTargetNotReachedBigDec.Dec(),
			// Difference between given amount remaining in and amount in actually consumed which qpproximately equals to spread factor.
			expectedSpreadRewardChargeTotal: defaultAmountOne.Sub(hundredDec).Quo(oneMinusDefaultSpreadFactor).Sub(defaultAmountOne.Sub(hundredDec).Ceil()),
		},
		"5: custom amounts at high price levels - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(sqrt(100_000_000)),
			sqrtPriceTarget:  sqrt(100_000_100),
			liquidity:        math.GetLiquidityFromAmounts(osmomath.OneBigDec(), osmomath.BigDecFromDec(sqrt(100_000_000)), osmomath.BigDecFromDec(sqrt(100_000_100)), defaultAmountZero.TruncateInt(), defaultAmountOne.TruncateInt()),

			// this value is exactly enough to reach the target
			amountOneInRemaining: osmomath.NewDec(1336900668450),
			spreadFactor:         osmomath.ZeroDec(),

			expectedSqrtPriceNext: osmomath.BigDecFromDec(sqrt(100_000_100)),

			expectedAmountInConsumed: osmomath.NewDec(1336900668450),
			// subtracting smallest dec as a rounding error in favor of the pool.
			expectedAmountOut:               defaultAmountZero.TruncateDec().Sub(osmomath.SmallestDec()),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"6: valid zero difference between sqrt price current and sqrt price next, amount zero in is charged": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.BigDecFromDec(osmomath.MustNewDecFromStr("70.710663976517714496")),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("70.710663976517714496"),
			liquidity:        osmomath.MustNewDecFromStr("412478955692135.521499519343199632"),

			amountOneInRemaining: osmomath.NewDec(5416667230),
			spreadFactor:         osmomath.ZeroDec(),

			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("70.710663976517714496"),

			expectedAmountInConsumed:        osmomath.ZeroDec(),
			expectedAmountOut:               osmomath.ZeroDec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		// This edge case does not occur anymore. The fix observed in PR: https://github.com/osmosis-labs/osmosis/pull/6352
		// See linked issue for details of the change.
		"7: (fixed) invalid zero difference between sqrt price current and sqrt price next due to precision loss, full amount remaining in is charged and amount out calculated from sqrt price": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.BigDecFromDec(osmomath.MustNewDecFromStr("0.000001000049998750")),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("0.000001000049998751"),
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountOneInRemaining: osmomath.NewDec(99),
			spreadFactor:         osmomath.ZeroDec(),

			// computed with x/concentrated-liquidity/python/clmath.py
			// sqrtPriceCurrent + token_in / liquidity
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("0.0000010000499987509899752698"),

			expectedAmountInConsumed: osmomath.NewDec(99),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			// calculated with x/concentrated-liquidity/python/clmath.py
			// # Calculate amount in until sqrtPriceTarget
			// amountIn = calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceTarget, True)
			// Decimal('100.002498062401598791937822606808718081')
			// # Greater than amountOneInRemaining => calculate sqrtPriceNext
			//
			// amountOneInRemaining = Decimal('99')
			// sqrtPriceNext = get_next_sqrt_price_from_amount1_in_round_down(liquidity, sqrtPriceCurrent, amountOneInRemaining)
			// Decimal("0.000001000049998750989975269800000000")
			//
			// diff = (sqrtPriceNext - sqrtPriceCurrent)
			// diff = round_decimal(diff, 36, ROUND_FLOOR)
			// mul = round_decimal(liquidity * diff, 36, ROUND_FLOOR)
			// div1= round_decimal(mul / sqrtPriceNext, 36, ROUND_FLOOR)
			// round_decimal(div1 / sqrtPriceCurrent, 36, ROUND_FLOOR)
			expectedAmountOut:               osmomath.MustNewBigDecFromStr("98990100989815.389417309941839547862158319016747061").Dec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"8: invalid zero difference between sqrt price current and sqrt price next due to precision loss. Returns 0 for amounts out. Note that the caller should detect this and fail.": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.BigDecFromDec(types.MaxSqrtPrice).Sub(osmomath.SmallestBigDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice,
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountOneInRemaining: osmomath.SmallestDec(),
			spreadFactor:         osmomath.ZeroDec(),

			expectedSqrtPriceNext: types.MaxSqrtPriceBigDec.Sub(osmomath.SmallestBigDec()),

			// Note, this case would lead to an infinite loop or no progress made in swaps.
			// As a result, the caller should detect this and fail.
			expectedAmountInConsumed:        osmomath.ZeroDec(),
			expectedAmountOut:               osmomath.ZeroDec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountInConsumed, amountZeroOut, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketOutGivenIn(tc.sqrtPriceCurrent, osmomath.BigDecFromDec(tc.sqrtPriceTarget), tc.liquidity, tc.amountOneInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountInConsumed, amountInConsumed)
			suite.Require().Equal(tc.expectedAmountOut.String(), amountZeroOut.String())
			suite.Require().Equal(tc.expectedSpreadRewardChargeTotal, spreadRewardChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_OneForZero() {
	var (
		// Target is not reached means that we stop at the sqrt price earlier
		// than expected. As a result, we recalculate the amount out and amount in
		// necessary to reach the earlier target.
		// sqrtPriceNext = liquidity * sqrtPriceCurrent / (liquidity  - tokenOut * sqrtPriceCurrent)
		sqrtPriceTargetNotReached = osmomath.MustNewBigDecFromStr("70.709031125539448609385160972133434677")
		// liq * (sqrtPriceNext - sqrtPriceCurrent)
		amountOneTargetNotReached = osmomath.MustNewDecFromStr("61829304.427824073089251659")
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		sqrtPriceCurrent osmomath.BigDec
		sqrtPriceTarget  osmomath.Dec
		liquidity        osmomath.Dec

		amountZeroOutRemaining osmomath.Dec
		spreadFactor           osmomath.Dec

		expectedSqrtPriceNext           osmomath.BigDec
		expectedAmountZeroOutConsumed   osmomath.Dec
		expectedAmountOneIn             osmomath.Dec
		expectedSpreadRewardChargeTotal osmomath.Dec
	}{
		"1: no spread reward - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			// Add 100.
			amountZeroOutRemaining: defaultAmountZero.Add(hundredDec),
			spreadFactor:           osmomath.ZeroDec(),

			expectedSqrtPriceNext: osmomath.BigDecFromDec(defaultSqrtPriceUpper),
			// Reached target, so 100 is not consumed.
			// computed with x/concentrated-liquidity/python/clmath.py
			// (liquidity * (sqrtPriceTarget - sqrtPriceCurrent)) / (sqrtPriceCurrent * sqrtPriceTarget)
			// diff = (sqrtPriceTarget - sqrtPriceCurrent)
			// diff = round_decimal(diff, 36, ROUND_FLOOR)
			// mul_denom = (sqrtPriceTarget * sqrtPriceCurrent)
			// mul_denom = round_decimal(mul_denom, 36, ROUND_CEILING)
			// mul_numer = (liquidity * diff)
			// mul_numer = round_decimal(mul_numer, 36, ROUND_FLOOR)
			// round_decimal(mul_numer / mul_denom, 18, ROUND_FLOOR)
			// 13369.999999999998920003
			// Added 1 ULP per calculations above
			expectedAmountZeroOutConsumed:   defaultAmountZero.Add(oneULPDec),
			expectedAmountOneIn:             defaultAmountOne.Ceil(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"2: no spread reward - do not reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			amountZeroOutRemaining: defaultAmountZero.Sub(osmomath.NewDec(1000)),
			spreadFactor:           osmomath.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,

			expectedAmountZeroOutConsumed: defaultAmountZero.Sub(osmomath.NewDec(1000)),

			expectedAmountOneIn:             amountOneTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"3: 3% spread reward - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			amountZeroOutRemaining: defaultAmountZero.Quo(oneMinusDefaultSpreadFactor),
			spreadFactor:           defaultSpreadReward,

			expectedSqrtPriceNext: osmomath.BigDecFromDec(defaultSqrtPriceUpper),
			// Reached target, so 100 is not consumed.
			// computed with x/concentrated-liquidity/python/clmath.py
			// (liquidity * (sqrtPriceTarget - sqrtPriceCurrent)) / (sqrtPriceCurrent * sqrtPriceTarget)
			// diff = (sqrtPriceTarget - sqrtPriceCurrent)
			// diff = round_decimal(diff, 36, ROUND_FLOOR)
			// mul_denom = (sqrtPriceTarget * sqrtPriceCurrent)
			// mul_denom = round_decimal(mul_denom, 36, ROUND_CEILING)
			// mul_numer = (liquidity * diff)
			// mul_numer = round_decimal(mul_numer, 36, ROUND_FLOOR)
			// round_decimal(mul_numer / mul_denom, 18, ROUND_FLOOR)
			// 13369.999999999998920003
			// Added 1 ULP per calculations above
			expectedAmountZeroOutConsumed:   defaultAmountZero.Add(oneULPDec),
			expectedAmountOneIn:             defaultAmountOne.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountOne.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread reward - do not reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			amountZeroOutRemaining: defaultAmountZero.Sub(osmomath.NewDec(1000)),
			spreadFactor:           defaultSpreadReward,

			expectedSqrtPriceNext:           sqrtPriceTargetNotReached,
			expectedAmountZeroOutConsumed:   defaultAmountZero.Sub(osmomath.NewDec(1000)),
			expectedAmountOneIn:             amountOneTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(amountOneTargetNotReached.Ceil(), defaultSpreadReward),
		},
		"6: valid zero difference between sqrt price current and sqrt price next, amount zero in is charged": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("70.710663976517714496"),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("70.710663976517714496"),
			liquidity:        osmomath.MustNewDecFromStr("412478955692135.521499519343199632"),

			amountZeroOutRemaining: osmomath.NewDec(5416667230),
			spreadFactor:           osmomath.ZeroDec(),

			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("70.710663976517714496"),

			expectedAmountZeroOutConsumed:   osmomath.ZeroDec(),
			expectedAmountOneIn:             osmomath.ZeroDec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"7: difference between sqrt prices is under BigDec ULP. Rounding causes amount consumed be greater than amount remaining": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewBigDecFromStr("0.000001000049998750"),
			sqrtPriceTarget:  osmomath.MustNewDecFromStr("0.000001000049998751"),
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroOutRemaining: osmomath.SmallestDec(),
			spreadFactor:           osmomath.ZeroDec(),

			// computed with x/concentrated-liquidity/python/clmath.py
			// get_next_sqrt_price_from_amount0_round_up(liquidity, sqrtPriceCurrent, tokenOut)
			expectedSqrtPriceNext: osmomath.MustNewBigDecFromStr("0.000001000049998750000000000000000001"),

			// computed with x/concentrated-liquidity/python/clmath.py
			// calc_amount_zero_delta(liquidity, sqrtPriceCurrent, Decimal("0.000001000049998750000000000000000001"), False)
			// Note: amount consumed is greater than amountZeroOutRemaining.
			// This happens because we round up sqrt price next at precision end. However, the difference between
			// sqrt price current and sqrt price next is smaller than 10^-36
			// Let's compute next sqrt price without rounding:
			// product_num = liquidity * sqrtPriceCurrent
			// product_num = round_osmo_prec_up(product_num)
			// product_den =  tokenOut * sqrtPriceCurrent
			// product_den = round_osmo_prec_up(product_den)
			// product_num / (liquidity - product_den)
			// '0.00000100004999875000000000000000000000000000000001000075017501875'
			// This can lead to negative amount zero in swaps.
			// As a result, force the amountOut to be amountZeroOutRemaining.
			// See code comments in ComputeSwapWithinBucketInGivenOut(...)
			expectedAmountZeroOutConsumed: osmomath.SmallestDec(),
			// calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True)
			// math.ceil(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True))
			expectedAmountOneIn:             osmomath.OneDec(),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"8: swapping 1 ULP of osmomath.Dec leads to zero out being consumed (no progress made)": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: types.MaxSqrtPriceBigDec.Sub(osmomath.SmallestBigDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice,
			liquidity:        osmomath.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroOutRemaining: osmomath.SmallestDec(),
			spreadFactor:           osmomath.ZeroDec(),

			// product_num = liquidity * sqrtPriceCurrent
			// product_den =  tokenOut * sqrtPriceCurrent
			// product_den = round_osmo_prec_up(product_den)
			// round_osmo_prec_up(product_num / (liquidity - product_den))
			expectedSqrtPriceNext: types.MaxSqrtPriceBigDec,

			expectedAmountZeroOutConsumed: osmomath.ZeroDec(),
			// Rounded up to 1.
			expectedAmountOneIn:             osmomath.NewDec(1),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
		"9: swapping 1 ULP of osmomath.Dec with high liquidity leads to an amount consumed being greater than amount remaining": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: types.MaxSqrtPriceBigDec.Sub(osmomath.SmallestBigDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice,
			// Choose large liquidity on purpose
			liquidity: osmomath.MustNewDecFromStr("9999999999999999999999999999999999999999999999999999999999.937822606808718081"),

			amountZeroOutRemaining: osmomath.SmallestDec(),
			spreadFactor:           osmomath.ZeroDec(),

			// product_num = liquidity * sqrtPriceCurrent
			// product_den =  tokenOut * sqrtPriceCurrent
			// product_den = round_osmo_prec_up(product_den)
			// round_osmo_prec_up(product_num / (liquidity - product_den))
			expectedSqrtPriceNext: types.MaxSqrtPriceBigDec,

			// product_num = liquidity * diff
			// product_denom = sqrtPriceA * sqrtPriceB
			// produce _num / producy_denom
			// Results in 0.0000000000000001
			// Note, that this amount is greater than the amount remaining but amountRemaining gets chosen over it
			// See code comments in ComputeSwapWithinBucketInGivenOut(...)
			expectedAmountZeroOutConsumed: osmomath.SmallestDec(),

			// calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True)
			expectedAmountOneIn:             osmomath.MustNewDecFromStr("10000000000000000000000"),
			expectedSpreadRewardChargeTotal: osmomath.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountZeroOutConsumed, amountOneIn, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketInGivenOut(tc.sqrtPriceCurrent, osmomath.BigDecFromDec(tc.sqrtPriceTarget), tc.liquidity, tc.amountZeroOutRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountZeroOutConsumed.String(), amountZeroOutConsumed.String())
			suite.Require().Equal(tc.expectedAmountOneIn, amountOneIn)
			suite.Require().Equal(tc.expectedSpreadRewardChargeTotal.String(), spreadRewardChargeTotal.String())
		})
	}
}

func (suite *StrategyTestSuite) TestInitializeNextTickIterator_OneForZero() {
	tests := map[string]tickIteratorTest{
		"1 position, one for zero": {
			preSetPositions: []position{
				{
					lowerTick: -100,
					upperTick: 100,
				},
			},
			tickSpacing:    defaultTickSpacing,
			expectIsValid:  true,
			expectNextTick: 100,
		},
		"2 positions, one for zero": {
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
			expectNextTick: 200,
		},
		"lower tick lands on current tick, one for zero": {
			preSetPositions: []position{
				{
					lowerTick: 0,
					upperTick: 100,
				},
			},
			tickSpacing:    defaultTickSpacing,
			expectIsValid:  true,
			expectNextTick: 100,
		},
		"upper tick lands on current tick, one for zero": {
			preSetPositions: []position{
				{
					lowerTick: -100,
					upperTick: 0,
				},
				{
					lowerTick: 100,
					upperTick: 200,
				},
			},
			tickSpacing:    defaultTickSpacing,
			expectIsValid:  true,
			expectNextTick: 100,
		},
		"no ticks, one for zero": {
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
			expectNextTick: 1,
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
			expectNextTick: 2,
		},
		"lower tick lands on current tick, 1 tick spacing": {
			preSetPositions: []position{
				{
					lowerTick: 0,
					upperTick: 1,
				},
			},
			tickSpacing:    1,
			expectIsValid:  true,
			expectNextTick: 1,
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
			expectNextTick: 1,
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
			expectNextTick: 10,
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
			expectNextTick: 1000,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, zero)
			suite.runTickIteratorTest(strategy, tc)
		})
	}
}
