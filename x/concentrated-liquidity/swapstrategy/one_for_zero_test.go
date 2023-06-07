package swapstrategy_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

func (suite *StrategyTestSuite) setupNewOneForZeroSwapStrategy(sqrtPriceLimit sdk.Dec, spread sdk.Dec) swapstrategy.SwapStrategy {
	suite.SetupTest()
	return swapstrategy.New(false, sqrtPriceLimit, suite.App.GetKey(types.ModuleName), spread, defaultTickSpacing)
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

func (suite *StrategyTestSuite) TestComputeSwapStepOutGivenIn_OneForZero() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceLower
		sqrtPriceNext    = defaultSqrtPriceUpper

		// sqrt_price_current + token_in / liquidity
		sqrtPriceTargetNotReached = sdk.MustNewDecFromStr("70.710678085714122880")
		// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
		amountZeroTargetNotReached = sdk.MustNewDecFromStr("13369.979999999989129753")

		sqrt = func(x int64) sdk.Dec {
			sqrt, _ := sdk.NewDec(x).ApproxSqrt()
			return sqrt
		}
	)

	tests := map[string]struct {
		sqrtPriceCurrent     sdk.Dec
		sqrtPriceTarget      sdk.Dec
		liquidity            sdk.Dec
		amountOneInRemaining sdk.Dec
		spreadFactor         sdk.Dec

		expectedSqrtPriceNext           sdk.Dec
		expectedAmountInConsumed        sdk.Dec
		expectedAmountOut               sdk.Dec
		expectedSpreadRewardChargeTotal sdk.Dec

		expectError error
	}{
		"1: no spread factor - reach target": {
			sqrtPriceCurrent: sqrtPriceCurrent,
			sqrtPriceTarget:  sqrtPriceNext,
			liquidity:        defaultLiquidity,
			// Add 100.
			amountOneInRemaining: defaultAmountOne.Add(sdk.NewDec(100)),
			spreadFactor:         sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Reached target, so 100 is not consumed.
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:               defaultAmountZero.Sub(sdk.SmallestDec()), // subtracting smallest dec to account for truncations in favor of the pool.
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"2: no spread factor - do not reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(sdk.NewDec(100)),
			spreadFactor:         sdk.ZeroDec(),

			// sqrt_price_current + token_in / liquidity
			expectedSqrtPriceNext:    sqrtPriceTargetNotReached,
			expectedAmountInConsumed: defaultAmountOne.Sub(sdk.NewDec(100)).Ceil(),
			// subtracting 3 * smallest dec to account for truncations in favor of the pool.
			expectedAmountOut:               amountZeroTargetNotReached.Sub(sdk.SmallestDec().MulInt64(3)),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% spread factor - reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Add(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultSpreadReward)),
			spreadFactor:         defaultSpreadReward,

			expectedSqrtPriceNext:    sqrtPriceNext,
			expectedAmountInConsumed: defaultAmountOne.Ceil(),
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent) / (sqrtPriceNext * sqrtPriceCurrent)
			expectedAmountOut:               defaultAmountZero.Sub(sdk.SmallestDec()), // subtracting smallest dec to account for truncations in favor of the pool.
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountOne.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread factor - do not reach target": {
			sqrtPriceCurrent:     sqrtPriceCurrent,
			sqrtPriceTarget:      sqrtPriceNext,
			liquidity:            defaultLiquidity,
			amountOneInRemaining: defaultAmountOne.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultSpreadReward)),
			spreadFactor:         defaultSpreadReward,

			expectedSqrtPriceNext:    sqrtPriceTargetNotReached,
			expectedAmountInConsumed: defaultAmountOne.Sub(sdk.NewDec(100)).Ceil(),
			// subtracting 3 * smallest dec to account for truncations in favor of the pool.
			expectedAmountOut: amountZeroTargetNotReached.Sub(sdk.SmallestDec().MulInt64(3)),
			// Difference between given amount remaining in and amount in actually consumed which qpproximately equals to spread factor.
			expectedSpreadRewardChargeTotal: defaultAmountOne.Sub(sdk.NewDec(100)).Quo(sdk.OneDec().Sub(defaultSpreadReward)).Sub(defaultAmountOne.Sub(sdk.NewDec(100)).Ceil()),
		},
		"5: custom amounts at high price levels - reach target": {
			sqrtPriceCurrent: sqrt(100_000_000),
			sqrtPriceTarget:  sqrt(100_000_100),
			liquidity:        math.GetLiquidityFromAmounts(sqrt(1), sqrt(100_000_000), sqrt(100_000_100), defaultAmountZero.TruncateInt(), defaultAmountOne.TruncateInt()),

			// this value is exactly enough to reach the target
			amountOneInRemaining: sdk.NewDec(1336900668450),
			spreadFactor:         sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrt(100_000_100),

			expectedAmountInConsumed: sdk.NewDec(1336900668450),
			// subtracting smallest dec as a rounding error in favor of the pool.
			expectedAmountOut:               defaultAmountZero.TruncateDec().Sub(sdk.SmallestDec()),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountInConsumed, amountZeroOut, spreadRewardChargeTotal := strategy.ComputeSwapStepOutGivenIn(tc.sqrtPriceCurrent, tc.sqrtPriceTarget, tc.liquidity, tc.amountOneInRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountInConsumed, amountInConsumed)
			suite.Require().Equal(tc.expectedAmountOut, amountZeroOut)
			suite.Require().Equal(tc.expectedSpreadRewardChargeTotal, spreadRewardChargeTotal)
		})
	}
}

func (suite *StrategyTestSuite) TestComputeSwapStepInGivenOut_OneForZero() {
	var (
		sqrtPriceCurrent = defaultSqrtPriceLower
		sqrtPriceNext    = defaultSqrtPriceUpper
		sqrtPriceTarget  = sqrtPriceNext
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

	// sqrtPriceCurrent, sqrtPriceTarget, liquidity are all set to defaults defined above.
	tests := map[string]struct {
		amountZeroOutRemaining sdk.Dec
		spreadFactor           sdk.Dec

		expectedSqrtPriceNext           sdk.Dec
		expectedAmountZeroOutConsumed   sdk.Dec
		expectedAmountOneIn             sdk.Dec
		expectedSpreadRewardChargeTotal sdk.Dec

		expectError error
	}{
		"1: no spread reward - reach target": {
			// Add 100.
			amountZeroOutRemaining: defaultAmountZero.Add(sdk.NewDec(100)),
			spreadFactor:           sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceNext,
			// Reached target, so 100 is not consumed.
			expectedAmountZeroOutConsumed: defaultAmountZero.Sub(sdk.SmallestDec()), // subtracting smallest dec to account for truncations in favor of the pool.
			// liquidity * (sqrtPriceNext - sqrtPriceCurrent)
			expectedAmountOneIn:             defaultAmountOne.Ceil(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"2: no spread reward - do not reach target": {
			amountZeroOutRemaining: defaultAmountZero.Sub(sdk.NewDec(1000)),
			spreadFactor:           sdk.ZeroDec(),

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,

			// subtracting 3 * smallest dec to account for truncations in favor of the pool.
			expectedAmountZeroOutConsumed: amountZeroTargetNotReached.Sub(sdk.SmallestDec().MulInt64(3)),

			expectedAmountOneIn:             amountOneTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: sdk.ZeroDec(),
		},
		"3: 3% spread reward - reach target": {
			amountZeroOutRemaining: defaultAmountZero.Quo(sdk.OneDec().Sub(defaultSpreadReward)),
			spreadFactor:           defaultSpreadReward,

			expectedSqrtPriceNext:           sqrtPriceNext,
			expectedAmountZeroOutConsumed:   defaultAmountZero.Sub(sdk.SmallestDec()), // subtracting smallest dec to account for truncations in favor of the pool.
			expectedAmountOneIn:             defaultAmountOne.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(defaultAmountOne.Ceil(), defaultSpreadReward),
		},
		"4: 3% spread reward - do not reach target": {
			amountZeroOutRemaining: defaultAmountZero.Sub(sdk.NewDec(1000)),
			spreadFactor:           defaultSpreadReward,

			expectedSqrtPriceNext: sqrtPriceTargetNotReached,
			// subtracting 3 * smallest dec to account for truncations in favor of the pool.
			expectedAmountZeroOutConsumed:   amountZeroTargetNotReached.Sub(sdk.SmallestDec().MulInt64(3)),
			expectedAmountOneIn:             amountOneTargetNotReached.Ceil(),
			expectedSpreadRewardChargeTotal: swapstrategy.ComputeSpreadRewardChargeFromAmountIn(amountOneTargetNotReached.Ceil(), defaultSpreadReward),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, tc.spreadFactor)
			sqrtPriceNext, amountZeroOutConsumed, amountOneIn, spreadRewardChargeTotal := strategy.ComputeSwapStepInGivenOut(sqrtPriceCurrent, sqrtPriceTarget, defaultLiquidity, tc.amountZeroOutRemaining)

			suite.Require().Equal(tc.expectedSqrtPriceNext, sqrtPriceNext)
			suite.Require().Equal(tc.expectedAmountZeroOutConsumed, amountZeroOutConsumed)
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
			expectIsValid:  true,
			expectNextTick: 100,
		},
		"no ticks, one for zero": {
			expectIsValid: false,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			strategy := suite.setupNewOneForZeroSwapStrategy(types.MaxSqrtPrice, zero)
			suite.runTickIteratorTest(strategy, tc)
		})
	}
}
