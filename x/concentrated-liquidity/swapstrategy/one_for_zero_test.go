package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

func (suite *StrategyTestSuite) setupNewOneForZeroSwapStrategy(sqrtPriceLimit sdk.Dec, spread sdk.Dec) swapstrategy.SwapStrategy {
	suite.SetupTest()
	return swapstrategy.New(false, sqrtPriceLimit, suite.App.GetKey(types.ModuleName), spread)
}

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
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(tc.sqrtPriceLimit, zero)
			actualSqrtTargetPrice := strategy.GetSqrtTargetPrice(tc.nextTickSqrtPrice)
			suite.Require().Equal(tc.expectedResult, actualSqrtTargetPrice)
		})
	}
}

// Note: estimates below are computed using x/concentrated-liquidity/python/clmath.py
func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_OneForZero() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceLower
		sqrtPriceNext    = defaultSqrtPriceUpper

		// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
		actualAmountZeroTargetNotReachedBigDec = osmomath.MustNewDecFromStr("13369.979999999989602986240259440383244931")

		sqrt = func(x int64) sdk.Dec {
			sqrt, _ := osmomath.MonotonicSqrt(sdk.NewDec(x))
			return sqrt
		}
	)

	tests := map[string]struct {
		// TODO revisit each test case and review values
		sqrtPriceCurrent     osmomath.BigDec
		sqrtPriceTarget      sdk.Dec
		liquidity            sdk.Dec
		amountOneInRemaining sdk.Dec
		spreadFactor         sdk.Dec

		expectedSqrtPriceNext           osmomath.BigDec
		expectedAmountInConsumed        sdk.Dec
		expectedAmountOut               sdk.Dec
		expectedSpreadRewardChargeTotal sdk.Dec
	}{
		"1: no spread factor - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(sqrtPriceCurrent),
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountOneInRemaining: defaultAmountOne.Add(sdk.NewDec(100)),
			spreadFactor:         sdk.ZeroDec(),

			expectedSqrtPriceNext: osmomath.BigDecFromSDKDec(sqrtPriceNext),
			// Reached target, so 100 is not consumed.
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:               defaultAmountZeroBigDec.SDKDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"2: no spread factor - do not reach target": {
			sqrtPriceCurrent:     osmomath.BigDecFromSDKDec(sqrtPriceCurrent),
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(sdk.NewDec(100)),
			spreadFactor:         sdk.ZeroDec(),

			// sqrtPriceCurrent + round_osmo_prec_down(token_in / liquidity)
			// sqrtPriceCurrent + token_in / liquidity
			expectedSqrtPriceNext:           osmomath.MustNewDecFromStr("70.710678085714122880779431539932994712"),
			expectedAmountInConsumed:        defaultAmountOne.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOut:               actualAmountZeroTargetNotReachedBigDec.SDKDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% spread factor - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(sqrtPriceCurrent),
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,

			amountOneInRemaining:     defaultAmountOne.Add(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultSpreadReward)),
			spreadFactor:             defaultSpreadReward,
			expectedSqrtPriceNext:    osmomath.BigDecFromSDKDec(sqrtPriceNext),
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:               defaultAmountZeroBigDec.SDKDec(), // subtracting smallest dec to account for truncations in favor of the pool.
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountOne.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread factor - do not reach target": {
			sqrtPriceCurrent:     osmomath.BigDecFromSDKDec(sqrtPriceCurrent),
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(sdk.NewDec(100)).QuoRoundUp(sdk.OneDec().Sub(defaultSpreadReward)),
			spreadFactor:         defaultSpreadReward,

			// sqrtPriceCurrent + round_osmo_prec_down(round_osmo_prec_down(round_sdk_prec_up(token_in / (1 - spreadFactor )) * (1 - spreadFactor)) / liquidity)
			expectedSqrtPriceNext:    osmomath.MustNewDecFromStr("70.710678085714122880779431540005464097"),
			expectedAmountInConsumed: defaultAmountOne.Sub(sdk.NewDec(100)).Ceil(),
			expectedAmountOut:        actualAmountZeroTargetNotReachedBigDec.SDKDec(),
			// Difference between given amount remaining in and amount in actually consumed which qpproximately equals to spread factor.
			expectedSpreadRewardChargeTotal: defaultAmountOne.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultSpreadReward)).Sub(defaultAmountOne.Sub(sdk.NewDec(100)).Ceil()),
		},
		"5: custom amounts at high price levels - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(sqrt(100_000_000)),
			sqrtPriceTarget:  sqrt(100_000_100),
			liquidity:        math.GetLiquidityFromAmounts(osmomath.OneDec(), sqrt(100_000_000), sqrt(100_000_100), defaultAmountZero.TruncateInt(), defaultAmountOne.TruncateInt()),

			// this value is exactly enough to reach the target
			amountOneInRemaining: sdk.NewDec(1336900668450),
			spreadFactor:         sdk.ZeroDec(),

			expectedSqrtPriceNext: osmomath.BigDecFromSDKDec(sqrt(100_000_100)),

			expectedAmountInConsumed: sdk.NewDec(1336900668450),
			// subtracting smallest dec as a rounding error in favor of the pool.
			expectedAmountOut:               defaultAmountZero.TruncateDec().Sub(sdk.SmallestDec()),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"6: valid zero difference between sqrt price current and sqrt price next, amount zero in is charged": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(sdk.MustNewDecFromStr("70.710663976517714496")),
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.710663976517714496"),
			liquidity:        sdk.MustNewDecFromStr("412478955692135.521499519343199632"),

			amountOneInRemaining: sdk.NewDec(5416667230),
			spreadFactor:         sdk.ZeroDec(),

			expectedSqrtPriceNext: osmomath.MustNewDecFromStr("70.710663976517714496"),

			expectedAmountInConsumed:        sdk.ZeroDec(),
			expectedAmountOut:               sdk.ZeroDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"7: invalid zero difference between sqrt price current and sqrt price next due to precision loss, full amount remaining in is charged and amount out calculated from sqrt price": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(sdk.MustNewDecFromStr("0.000001000049998750")),
			sqrtPriceTarget:  sdk.MustNewDecFromStr("0.000001000049998751"),
			liquidity:        sdk.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountOneInRemaining: sdk.NewDec(99),
			spreadFactor:         sdk.ZeroDec(),

			// computed with x/concentrated-liquidity/python/clmath.py
			// sqrtPriceCurrent + token_in / liquidity
			expectedSqrtPriceNext: osmomath.MustNewDecFromStr("0.0000010000499987509899752698"),

			expectedAmountInConsumed: sdk.NewDec(99),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			// calculated with x/concentrated-liquidity/python/clmath.py
			// diff = (sqrtPriceNext - sqrtPriceCurrent)
			// diff = round_decimal(diff, 36, ROUND_FLOOR) (0.000000000000000000989975269800000000)
			// mul = (sqrtPriceNext * sqrtPriceCurrent)
			// mul = round_decimal(mul, 36, ROUND_CEILING) (0.000000000001000100000000865026329827)
			//  round_decimal(liquidity * diff / mul, 36, ROUND_FLOOR)
			expectedAmountOut:               osmomath.MustNewDecFromStr("98990100989815.389417309844929293132374729779331247").SDKDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"8: invalid zero difference between sqrt price current and sqrt price next due to precision loss. Returns 0 for amounts out. Note that the caller should detect this and fail.": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(types.MaxSqrtPrice).Sub(osmomath.SmallestDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice,
			liquidity:        sdk.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountOneInRemaining: sdk.SmallestDec(),
			spreadFactor:         sdk.ZeroDec(),

			expectedSqrtPriceNext: types.MaxSqrtPriceBigDec.Sub(osmomath.SmallestDec()),

			// Note, this case would lead to an infinite loop or no progress made in swaps.
			// As a result, the caller should detect this and fail.
			expectedAmountInConsumed:        sdk.ZeroDec(),
			expectedAmountOut:               sdk.ZeroDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountInConsumed, amountZeroOut, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketOutGivenIn(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountOneInRemaining)

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
		sqrtPriceTargetNotReached = osmomath.MustNewDecFromStr("70.709031125539448609385160972133434677")
		// liq * (sqrtPriceNext - sqrtPriceCurrent)
		amountOneTargetNotReached = sdk.MustNewDecFromStr("61829304.427824073089251659")
	)

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		sqrtPriceCurrent osmomath.BigDec
		sqrtPriceTarget  sdk.Dec
		liquidity        sdk.Dec

		amountZeroOutRemaining sdk.Dec
		spreadFactor           sdk.Dec

		expectedSqrtPriceNext           osmomath.BigDec
		expectedAmountZeroOutConsumed   sdk.Dec
		expectedAmountOneIn             sdk.Dec
		expectedSpreadRewardChargeTotal sdk.Dec
	}{
		"1: no spread reward - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			// Add 100.
			amountZeroOutRemaining: defaultAmountZero.Add(sdk.NewDec(100)),
			spreadFactor:           sdk.ZeroDec(),

			expectedSqrtPriceNext: osmomath.BigDecFromSDKDec(defaultSqrtPriceUpper),
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
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"2: no spread reward - do not reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			amountZeroOutRemaining: defaultAmountZero.Sub(sdk.NewDec(1000)),
			spreadFactor:           sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,

			expectedAmountZeroOutConsumed: defaultAmountZero.Sub(sdk.NewDec(1000)),

			expectedAmountOneIn:             amountOneTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% spread reward - reach target": {
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			amountZeroOutRemaining: defaultAmountZero.Quo(sdk.OneDec().Sub(defaultSpreadReward)),
			spreadFactor:           defaultSpreadReward,

			expectedSqrtPriceNext: osmomath.BigDecFromSDKDec(defaultSqrtPriceUpper),
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
			sqrtPriceCurrent: osmomath.BigDecFromSDKDec(defaultSqrtPriceLower),
			sqrtPriceTarget:  defaultSqrtPriceUpper,
			liquidity:        defaultLiquidity,

			amountZeroOutRemaining: defaultAmountZero.Sub(sdk.NewDec(1000)),
			spreadFactor:           defaultSpreadReward,

			expectedSqrtPriceNext:           sqrtPriceTargetNotReached,
			expectedAmountZeroOutConsumed:   defaultAmountZero.Sub(sdk.NewDec(1000)),
			expectedAmountOneIn:             amountOneTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(amountOneTargetNotReached.Ceil(), defaultSpreadReward),
		},
		"6: valid zero difference between sqrt price current and sqrt price next, amount zero in is charged": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewDecFromStr("70.710663976517714496"),
			sqrtPriceTarget:  sdk.MustNewDecFromStr("70.710663976517714496"),
			liquidity:        sdk.MustNewDecFromStr("412478955692135.521499519343199632"),

			amountZeroOutRemaining: sdk.NewDec(5416667230),
			spreadFactor:           sdk.ZeroDec(),

			expectedSqrtPriceNext: osmomath.MustNewDecFromStr("70.710663976517714496"),

			expectedAmountZeroOutConsumed:   sdk.ZeroDec(),
			expectedAmountOneIn:             sdk.ZeroDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"7: difference between sqrt prices is under BigDec ULP. Rounding causes amount consumed be greater than amount remaining": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: osmomath.MustNewDecFromStr("0.000001000049998750"),
			sqrtPriceTarget:  sdk.MustNewDecFromStr("0.000001000049998751"),
			liquidity:        sdk.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroOutRemaining: sdk.SmallestDec(),
			spreadFactor:           sdk.ZeroDec(),

			// computed with x/concentrated-liquidity/python/clmath.py
			// get_next_sqrt_price_from_amount0_round_up(liquidity, sqrtPriceCurrent, tokenOut)
			expectedSqrtPriceNext: osmomath.MustNewDecFromStr("0.000001000049998750000000000000000001"),

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
			expectedAmountZeroOutConsumed: sdk.SmallestDec(),
			// calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True)
			// math.ceil(calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True))
			expectedAmountOneIn:             sdk.OneDec(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"8: swapping 1 ULP of sdk.Dec leads to zero out being consumed (no progress made)": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: types.MaxSqrtPriceBigDec.Sub(osmomath.SmallestDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice,
			liquidity:        sdk.MustNewDecFromStr("100002498062401598791.937822606808718081"),

			amountZeroOutRemaining: sdk.SmallestDec(),
			spreadFactor:           sdk.ZeroDec(),

			// product_num = liquidity * sqrtPriceCurrent
			// product_den =  tokenOut * sqrtPriceCurrent
			// product_den = round_osmo_prec_up(product_den)
			// round_osmo_prec_up(product_num / (liquidity - product_den))
			expectedSqrtPriceNext: types.MaxSqrtPriceBigDec,

			expectedAmountZeroOutConsumed: sdk.ZeroDec(),
			// Rounded up to 1.
			expectedAmountOneIn:             sdk.NewDec(1),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"9: swapping 1 ULP of sdk.Dec with high liquidity leads to an amount consumed being greater than amount remaining": {
			// Note the numbers are hand-picked to reproduce this specific case.
			sqrtPriceCurrent: types.MaxSqrtPriceBigDec.Sub(osmomath.SmallestDec()),
			sqrtPriceTarget:  types.MaxSqrtPrice,
			// Choose large liquidity on purpose
			liquidity: sdk.MustNewDecFromStr("9999999999999999999999999999999999999999999999999999999999.937822606808718081"),

			amountZeroOutRemaining: sdk.SmallestDec(),
			spreadFactor:           sdk.ZeroDec(),

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
			expectedAmountZeroOutConsumed: sdk.SmallestDec(),

			// calc_amount_one_delta(liquidity, sqrtPriceCurrent, sqrtPriceNext, True)
			expectedAmountOneIn:             sdk.MustNewDecFromStr("10000000000000000000000"),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountZeroOutConsumed, amountOneIn, spreadRewardChargeTotal := strategy.ComputeSwapWithinBucketInGivenOut(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountZeroOutRemaining)

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
