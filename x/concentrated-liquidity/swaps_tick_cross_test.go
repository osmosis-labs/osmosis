package concentrated_liquidity_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

// SwapTickCrossTestSuite tests that the ticks are
// updated correctly after swapping in various scenarios.
type SwapTickCrossTestSuite struct {
	KeeperTestSuite
}

func TestSwapTickCrossTestSuite(t *testing.T) {
	suite.Run(t, new(SwapTickCrossTestSuite))
}

// CreatePositionTickSpacingsFromCurrentTick creates a position with the passed in tick spacings away from the current tick.
func (s *SwapTickCrossTestSuite) CreatePositionTickSpacingsFromCurrentTick(poolId uint64, tickSpacingsAwayFromCurrentTick uint64) positionMeta {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	currentTick := pool.GetCurrentTick()

	tickSpacing := int64(pool.GetTickSpacing())

	// make sure that current tick is a multiple of tick spacing
	currentTick = currentTick - (currentTick % tickSpacing)

	lowerTick := currentTick - int64(tickSpacingsAwayFromCurrentTick)*tickSpacing
	upperTick := currentTick + int64(tickSpacingsAwayFromCurrentTick)*tickSpacing
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	positionId, _, _, liquidityNarrowRangeTwo, _, _, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[0], DefaultCoins, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)

	return positionMeta{
		positionId: positionId,
		lowerTick:  lowerTick,
		upperTick:  upperTick,
		liquidity:  liquidityNarrowRangeTwo,
	}
}

// tickToSqrtPrice a helper to convert a tick to a sqrt price.
func (s *SwapTickCrossTestSuite) tickToSqrtPrice(tick int64) sdk.Dec {
	_, sqrtPrice, err := math.TickToSqrtPrice(tick)
	s.Require().NoError(err)
	return sqrtPrice
}

// validateIteratorLeftZeroForOne is a helper to validate the next initialized tick iterator
// in the left (zfo) direction of the swap.
func (s *SwapTickCrossTestSuite) validateIteratorLeftZeroForOne(poolId uint64, expectedTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	zeroForOneSwapStrategy, _, err := s.App.ConcentratedLiquidityKeeper.SetupSwapStrategy(s.Ctx, pool, sdk.ZeroDec(), pool.GetToken0(), types.MinSqrtPrice)
	s.Require().NoError(err)
	initializedTickValue := pool.GetCurrentTick()
	iter := zeroForOneSwapStrategy.InitializeNextTickIterator(s.Ctx, pool.GetId(), initializedTickValue)
	s.Require().True(iter.Valid())
	nextTick, err := types.TickIndexFromBytes(iter.Key())
	s.Require().NoError(err)
	s.Require().NoError(iter.Close())

	s.Require().Equal(expectedTick, nextTick)
}

// validateIteratorRightOneForZero is a helper to validate the next initialized tick iterator
// in the right (ofz) direction of the swap.
func (s *SwapTickCrossTestSuite) validateIteratorRightOneForZero(poolId uint64, expectedTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Setup swap strategy directly as it would fail validation if constructed via SetupSwapStrategy(...)
	oneForZeroSwapStrategy := swapstrategy.New(false, types.MaxSqrtPrice, s.App.GetKey(types.ModuleName), sdk.ZeroDec())
	s.Require().NoError(err)
	initializedTickValue := pool.GetCurrentTick()
	iter := oneForZeroSwapStrategy.InitializeNextTickIterator(s.Ctx, pool.GetId(), initializedTickValue)
	s.Require().True(iter.Valid())
	nextTick, err := types.TickIndexFromBytes(iter.Key())
	s.Require().NoError(err)
	s.Require().NoError(iter.Close())

	s.Require().Equal(expectedTick, nextTick)
}

// assertPositionInRange a helper to assert that a position with the given lowerTick and upperTick is in range.
func (s *SwapTickCrossTestSuite) assertPositionInRange(poolId uint64, lowerTick int64, upperTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	isInRange := pool.IsCurrentTickInRange(lowerTick, upperTick)
	s.Require().True(isInRange, "currentTick: %d, lowerTick %d, upperTick: %d", pool.GetCurrentTick(), lowerTick, upperTick)
}

// assertPositionOutOfRange a helper to assert that a position with the given lowerTick and upperTick is out of range.
func (s *SwapTickCrossTestSuite) assertPositionOutOfRange(poolId uint64, lowerTick int64, upperTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	isInRange := pool.IsCurrentTickInRange(lowerTick, upperTick)
	s.Require().False(isInRange, "currentTick: %d, lowerTick %d, upperTick: %d", pool.GetCurrentTick(), lowerTick, upperTick)
}

// assertPositionRangeConditional a helper to assert that a position with the given lowerTick and upperTick is in or out of range
// depending on the isOutOfRangeExpected flag.
func (s *SwapTickCrossTestSuite) assertPositionRangeConditional(poolId uint64, isOutOfRangeExpected bool, lowerTick int64, upperTick int64) {
	if isOutOfRangeExpected {
		s.assertPositionOutOfRange(poolId, lowerTick, upperTick)
	} else {
		s.assertPositionInRange(poolId, lowerTick, upperTick)
	}
}

// swapZeroForOneLeft swaps amount in the left (zfo) direction of the swap.
// Asserts that no error is returned.
func (s *SwapTickCrossTestSuite) swapZeroForOneLeft(poolId uint64, amount sdk.Coin) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(amount))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, amount, pool.GetToken1(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().NoError(err)
}

// swapOneForZeroRight swaps amount in the right (ofz) direction of the swap.
// Asserts that no error is returned.
func (s *SwapTickCrossTestSuite) swapOneForZeroRight(poolId uint64, amount sdk.Coin) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(amount))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, amount, pool.GetToken0(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().NoError(err)
}

// assertPoolLiquidityEquals a helper to assert that the liquidity of a pool is equal to the expected value.
func (s *SwapTickCrossTestSuite) assertPoolLiquidityEquals(poolId uint64, expectedLiquidity sdk.Dec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.Require().Equal(expectedLiquidity, pool.GetLiquidity())
}

// assertPoolTickEquals a helper to assert that the current tick of a pool is equal to the expected value.
func (s *SwapTickCrossTestSuite) assertPoolTickEquals(poolId uint64, expectedTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.Require().Equal(expectedTick, pool.GetCurrentTick())
}

// TestSwapOutGivenIn_Tick_Initialization_And_Crossing tests that ticks are initialized and updated correctly
// across multiple swaps. In particular, this test does 2 swaps.
// For every test case, the following invariants hold:
// * first swap MAY cross a tick depending on test case configuration
// * second swap MUST NOT cross a tick (only swap in-between ticks)
// * both swaps are in the same direction
//
// It creates 3 positions:
// - (FR) full range
// - narrow range where
//   - (NR1) narrow range is 2 tick spacings around the current tick
//   - (NR2) narrow range'is 5 tick spacings above current tick
//
// Position setup:
//
//	                                                                              cur tick
//	                                                                                  |
//	narrow range 1  (NR1)                                                        //////////
//	narrow range 2  (NR2)                                                  //////////////////////
//	full range:     (FR)      //////////////////////////////////////////////////////////////////////////////////////////
//	                          MinTick                                                                            MaxTick
//
// Both directions are tested. Zero for one (left). One for zero (right).
//
// For every test case, we set it up with 1 and 100 tick spacing.
//
// This test helped to identify 2 high severity bugs:
//
// BUG 1. Wrong "current tick" update when swapping in zero for one (left) direction.
//
// The repro of the bug: if we perform a swap and cross a tick, the subsequent swap in the same direction
// would cross the same tick again and mistakenly kick in liquidity, completely invalidating pool state.
//
// Initial guess was that we should not search inclusive of the current tick when initializing the next tick iterator
// in zero for one direction. Otherwise, it would be possible to initialize nextTick to the already crossed currenTick.
// and cross it twice. Once during the first swap, and once during the second swap.
// However, when initializing the tick for the second swap, it is correct to search inclusive of the current
// tick. To understand this, consider another case where we swap right (zfo) and cross tick X, then when we swap
// left (ofz). In such a case, we must cross tick X in the opposite direction when going left.
// Therefore, the realization concluded that we need to special case the "update of the current tick after tick crossing
// when swapping in zero for one direction". In such a case, we should kick current tick 1 unit to the left (outside of
// the current range before first swap). This way, during second swap, we do not cross the already crossed tick again.
// While with the sequence of 2 swaps (zfo, ofz), we do cross the same tick twice as expected.
//
// This test is set up to reproduce all of the above.
//
// It does the first swap that stops exactly after crossing the NR1's lower tick.
// Next, it continues with the second swap that we do not expect to cross any ticks
// and remain in the bucket between NR2 lower tick and NR1 lower tick.
//
// Once the second swap is completed, we validate that the liquidity in the pool corresponds to
// full range position only. If the tick were to be crossed twice, we would have mistakenly
// subtracted liquidity net from the lower tick of the narrow position twice, making the
// current liquidity be zero.
//
// Prior to adding this test and implementing a fix, the system would have panicked with:
// "negative coin amount: -7070961745605321174329" at the end of the second swap. This stems
// from the invalid liquidity of zero that is incorrectly computed due to crossing the tick twice.
//
// Additionally, it sets up a case of swapping in one for zero direction to swap directly at the next initialzed
// tick. After the swap completes, we manually validate that next tick iterators return the correct values
// in both directions (left and right).
//
// BUG 2. Banker's rounding in sqrt price to tick conversion for zero for one swap makes current tick be off by 1,
// leading to tick being crossed twice.
//
// The consequences of this bug are similar to the first where we cross the tick twice. However, in
// this case the error occurs at the end of the second swap, not the first. The reason is that
// second swap takes in such a small amount as to barely move the sqrt price. Then, at the end it
// converts the sqrt price to tick and rounds up. Rounding up, makes current tick be equal to the
// tick we already crossed (be off by 1). As a result, inactive positions are treated as active
// The solution is to avoid tick update if the swap state's tick is smaller than the tick computed
// from the sqrt price. That is, if we already seen a further tick, we do not update it to an earlier one.
func (s *SwapTickCrossTestSuite) TestSwapOutGivenIn_Tick_Initialization_And_Crossing() {
	s.Setup()

	// testCase defines the test case configuration
	type testCase struct {
		name        string
		tickSpacing uint64

		swapTicksAway                  int64
		expectedTickAwayAfterFirstSwap int64
	}

	// expectedAndComputedValues defines the expected and computed values
	// that are derived from testCase parameters during the execution
	// of the test case.
	type expectedAndComputedValues struct {
		// computed inputs
		tokenIn sdk.Coin

		// computed expected outputs
		expectedTickAfterFirstSwap               int64
		expectedLiquidityAfterFirstAndSecondSwap sdk.Dec
		doesFirstSwapCrossTick                   bool
		expectedNextTickAfterFirstSwapZFOLeft    int64
		expectedNextTickAfterFirstSwapOFZRight   int64
	}

	const (
		tickSpacingOne = 1
		tickSpacing100 = 100

		// Defines how many tick spacings away from the current
		// tick NR1 and NR2 are. Note that lower and upper ticks
		// are equally spaced around the current tick.
		nr1TickSpacingsAway = 2
		nr2TickSpacingsAway = 5
	)

	// setupPoolAndPositions creates a pool and 3 positions:
	// - full range
	// - NR1: narrow range one with nr1TickSpacingsAway around the current tick
	// - NR2: narrow range two with nr2TickSpacingsAway around the current tick
	//
	// Returns the pool id and the 2 positions metatadata to be used in tests.
	setupPoolAndPositions := func(testTickSpacing uint64) (uint64, positionMeta, positionMeta) {
		pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, testTickSpacing, sdk.ZeroDec())
		poolId := pool.GetId()

		// Create a full range position
		s.FundAcc(s.TestAccs[0], DefaultCoins)
		_, _, _, liquidityFullRange, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolId, s.TestAccs[0], DefaultCoins)
		s.Require().NoError(err)

		// Refetch pool as the first position updated its state.
		pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
		s.Require().NoError(err)

		// Create narrow range position 2 tick spacings away the current tick
		narrowRangeOnePosition := s.CreatePositionTickSpacingsFromCurrentTick(poolId, nr1TickSpacingsAway)

		// Create narrow range position 5 tick spacings away from the current.
		narrowRangeTwoPosition := s.CreatePositionTickSpacingsFromCurrentTick(poolId, nr2TickSpacingsAway)

		// As a sanity check confirm that current liquidity corresponds
		// to the sum of liquidities of all positions.
		liquidityAllPositions := liquidityFullRange.Add(narrowRangeOnePosition.liquidity.Add(narrowRangeTwoPosition.liquidity))
		s.assertPoolLiquidityEquals(poolId, liquidityAllPositions)

		// Sanity check that that NR1 is in range prior to swap.
		s.assertPositionInRange(poolId, narrowRangeOnePosition.lowerTick, narrowRangeOnePosition.upperTick)

		// Sanity check that that NR2 is in range prior to swap.
		s.assertPositionInRange(poolId, narrowRangeTwoPosition.lowerTick, narrowRangeTwoPosition.upperTick)

		return poolId, narrowRangeOnePosition, narrowRangeTwoPosition
	}

	// validateAfterFirstSwap runs validation logic of the system's state after the first swap executes.
	// It validates the following:
	// - pool's current tick equals the expected value
	// - pool's liquidity equals the expected value
	// - NR1 position is either in or out of range depending on the test configuration
	// - NR2 position is always in range by construction of the test case
	// - next tick iterator towards left (zero for one) is correct
	// - next tick iterator towards right (one for zero) is correct
	validateAfterFirstSwap := func(poolId uint64, expectedValues expectedAndComputedValues, nr1Position positionMeta, nr2Position positionMeta) {
		// Assert that pool tick and liquidity correspond to the expected values.
		s.assertPoolTickEquals(poolId, expectedValues.expectedTickAfterFirstSwap)
		s.assertPoolLiquidityEquals(poolId, expectedValues.expectedLiquidityAfterFirstAndSecondSwap)

		// Check position ranges

		// First position is out of range if doesFirstSwapCrossUpperTickOne is true. Otherwise, it is in range.
		s.assertPositionRangeConditional(poolId, expectedValues.doesFirstSwapCrossTick, nr1Position.lowerTick, nr1Position.upperTick)

		// Second position is always in range by test construction.
		s.assertPositionInRange(poolId, nr2Position.lowerTick, nr2Position.upperTick)

		// Confirm that the next tick to be returned is correct for both zero for one and one for zero directions.

		// Validate iterator in the zero for one direction (left).
		s.validateIteratorLeftZeroForOne(poolId, expectedValues.expectedNextTickAfterFirstSwapZFOLeft)

		// Validate iterator in the one for zero direction (right).
		s.validateIteratorRightOneForZero(poolId, expectedValues.expectedNextTickAfterFirstSwapOFZRight)
	}

	// validateAfterSecondSwap runs validation logic of the system's state after the second swap executes.
	// It validates the following:
	// - pool's liquidity equals the expected value
	// - NR1 position is either in or out of range depending on the test configuration
	// - NR2 position is always in range by construction of the test case
	validateAfterSecondSwap := func(poolId uint64, expectedValues expectedAndComputedValues, nr1Position positionMeta, nr2Position positionMeta) {
		// Liquidity should remain the same by construction as we do not expect second swap to cross the tick.
		// This check helps validate that we start from the correct tick on second swap and do not cross
		// a tick twice inter-swap. Otherwise, if we were to cross it twice, the liquidity would be zero.
		s.assertPoolLiquidityEquals(poolId, expectedValues.expectedLiquidityAfterFirstAndSecondSwap)

		// Check position ranges

		// First position is out if range if doesFirstSwapCrossTick is true. Otherwise, it is in range.
		s.assertPositionRangeConditional(poolId, expectedValues.doesFirstSwapCrossTick, nr1Position.lowerTick, nr1Position.upperTick)

		// Second position is always in range by test construction.
		s.assertPositionInRange(poolId, nr2Position.lowerTick, nr2Position.upperTick)
	}

	// computeValuesForTestZeroForOne computes the expected values for the test case when swapping zero for one.
	// It does so by determining whether the first swap crosses NR1's lower tick from tc parameter.
	// Next, it determines the appropriate amount to swap in for the first swap as well as the
	// expected liquidity, current tick, and next tick in either direction to be returned by the iterator after first swap.
	computeValuesForTestZeroForOne := func(poolId uint64, tc testCase, nr1Position positionMeta, nr2Position positionMeta) expectedAndComputedValues {
		// Fetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		var (
			originalCurrentTick = pool.GetCurrentTick()
			// Note: it already validated to be correct in setupPoolAndPositionsCommon(...)
			liquidityAllPositions = pool.GetLiquidity()
		)

		tickToSwapTo := originalCurrentTick + tc.swapTicksAway
		expectedTickAfterFirstSwap := originalCurrentTick + tc.expectedTickAwayAfterFirstSwap

		doesFirstSwapCrossLowerTickOne := expectedTickAfterFirstSwap < nr1Position.lowerTick
		expectedLiquidityAfterFirstAndSecondSwap := liquidityAllPositions.Sub(nr1Position.liquidity)

		expectedNextTickAfterFirstSwapZFOLeft := nr2Position.lowerTick
		expectedNextTickAfterFirstSwapOFZRight := nr1Position.lowerTick
		if !doesFirstSwapCrossLowerTickOne {
			expectedLiquidityAfterFirstAndSecondSwap = expectedLiquidityAfterFirstAndSecondSwap.Add(nr1Position.liquidity)
			expectedNextTickAfterFirstSwapZFOLeft = nr1Position.lowerTick
			expectedNextTickAfterFirstSwapOFZRight = nr1Position.upperTick
		}

		// Compute the sqrt price corresponding to the lower tick
		// of the narrow range position.
		sqrtPriceTarget := s.tickToSqrtPrice(tickToSwapTo)

		// Check that narrow range position is considered in range
		isNarrowInRange := pool.IsCurrentTickInRange(nr1Position.lowerTick, nr1Position.upperTick)
		s.Require().True(isNarrowInRange)

		var (
			amountZeroIn   sdk.Dec = sdk.ZeroDec()
			sqrtPriceStart sdk.Dec = pool.GetCurrentSqrtPrice()

			liquidity = pool.GetLiquidity()
		)

		if tickToSwapTo < nr1Position.lowerTick {
			sqrtPriceLowerTickOne := s.tickToSqrtPrice(nr1Position.lowerTick)

			amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceLowerTickOne, sqrtPriceStart, true)

			sqrtPriceStart = sqrtPriceLowerTickOne

			liquidity = liquidity.Sub(nr1Position.liquidity)
		}

		// This is the total amount necessary to cross the lower tick of narrow position.
		// Note it is rounded up to ensure that the tick is crossed.
		amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceStart, true).Add(amountZeroIn)

		tokenZeroIn := sdk.NewCoin(pool.GetToken0(), amountZeroIn.Ceil().TruncateInt())

		expectedValues := expectedAndComputedValues{
			tokenIn: tokenZeroIn,

			expectedTickAfterFirstSwap:               expectedTickAfterFirstSwap,
			expectedLiquidityAfterFirstAndSecondSwap: expectedLiquidityAfterFirstAndSecondSwap,
			doesFirstSwapCrossTick:                   doesFirstSwapCrossLowerTickOne,
			expectedNextTickAfterFirstSwapZFOLeft:    expectedNextTickAfterFirstSwapZFOLeft,
			expectedNextTickAfterFirstSwapOFZRight:   expectedNextTickAfterFirstSwapOFZRight,
		}

		return expectedValues
	}

	// computeExpectedValuesForTestOneForZero computes the expected values for the test case when swapping one for zero.
	// It does so by determining whether the first swap crosses NR1's upper tick from tc parameter.
	// Next, it determines the appropriate amount to swap in for the first swap as well as the
	// expected liquidity, current tick, and next tick in either direction to be returned by the iterator after first swap.
	computeExpectedValuesForTestOneForZero := func(poolId uint64, tc testCase, nr1Position positionMeta, nr2Position positionMeta) expectedAndComputedValues {
		// Fetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		var (
			originalCurrentTick = pool.GetCurrentTick()
			// Note: it already validated to be correct in setupPoolAndPositionsCommon(...)
			liquidityAllPositions = pool.GetLiquidity()
		)

		// Compute the tick to swap to relative to the current tick depending on test case configuration.
		tickToSwapTo := originalCurrentTick + tc.swapTicksAway
		// Compute the expected tick after the first swap relative to the current tick depending on test case configuration.
		expectedTickAfterFirstSwap := originalCurrentTick + tc.expectedTickAwayAfterFirstSwap

		// If expected tick after the first swap is greater than or equal to the upper tick of the narrow range position
		// then the first swap will cross the upper tick of the narrow range position.
		// N.B.: equals matters since our "active bucket" definition is inclusive of the lower tick and exclusive of the upper tick.
		// One for zero swap swaps to the right.
		doesFirstSwapCrossUpperTickOne := expectedTickAfterFirstSwap >= nr1Position.upperTick

		var (
			expectedLiquidityAfterFirstAndSecondSwap sdk.Dec
			expectedNextTickAfterFirstSwapZFOLeft    int64
			expectedNextTickAfterFirstSwapOFZRight   int64
		)

		// If we expect swap to cross the tick, then we must be in the bucket
		// between the upper tick of NR1 and lower tick of the NR2.
		if doesFirstSwapCrossUpperTickOne {
			expectedNextTickAfterFirstSwapZFOLeft = nr1Position.upperTick
			expectedNextTickAfterFirstSwapOFZRight = nr2Position.upperTick
			expectedLiquidityAfterFirstAndSecondSwap = liquidityAllPositions.Sub(nr1Position.liquidity)
		} else {
			// If we expect swap to NOT cross the tick, then we must be in the original bucket
			// between the lower tick of NR1 and upper tick the NR1.
			expectedLiquidityAfterFirstAndSecondSwap = liquidityAllPositions
			expectedNextTickAfterFirstSwapZFOLeft = nr1Position.lowerTick
			expectedNextTickAfterFirstSwapOFZRight = nr1Position.upperTick
		}

		// Compute the sqrt price corresponding to the tick we want to swap to.
		sqrtPriceTarget := s.tickToSqrtPrice(tickToSwapTo)

		////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// 2. Estimate the amount of token 1 to swap to depending on test configuration.

		// Refetch pool as the second position creation updated its state.
		pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		var (
			amountOneIn    sdk.Dec = sdk.ZeroDec()
			sqrtPriceStart sdk.Dec = pool.GetCurrentSqrtPrice()
			liquidity              = pool.GetLiquidity()
		)

		if tickToSwapTo >= nr1Position.upperTick {
			_, sqrtPriceUpperOne, err := math.TickToSqrtPrice(nr1Position.upperTick)
			s.Require().NoError(err)

			amountOneIn = math.CalcAmount1Delta(liquidity, sqrtPriceUpperOne, sqrtPriceStart, true)

			sqrtPriceStart = sqrtPriceUpperOne

			liquidity = liquidity.Sub(nr1Position.liquidity)
		}

		// This is the total amount necessary to cross the lower tick of narrow position.
		// Note it is rounded up to ensure that the tick is crossed.
		amountOneIn = math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceStart, true).Add(amountOneIn)

		tokenOneIn := sdk.NewCoin(pool.GetToken1(), amountOneIn.Ceil().TruncateInt())

		expectedValues := expectedAndComputedValues{
			tokenIn: tokenOneIn,

			expectedTickAfterFirstSwap:               expectedTickAfterFirstSwap,
			expectedLiquidityAfterFirstAndSecondSwap: expectedLiquidityAfterFirstAndSecondSwap,
			doesFirstSwapCrossTick:                   doesFirstSwapCrossUpperTickOne,
			expectedNextTickAfterFirstSwapZFOLeft:    expectedNextTickAfterFirstSwapZFOLeft,
			expectedNextTickAfterFirstSwapOFZRight:   expectedNextTickAfterFirstSwapOFZRight,
		}

		return expectedValues
	}

	s.Run("zero for one", func() {
		testCases := map[string]testCase{
			// Group 1:
			// Test setup:
			// swap 1: just enough to cross lower tick of NR1
			// swap 2: stop between NR2 lower and lower NR1
			"group1 100 tick spacing, first swap crosses tick, second swap in same direction": {
				tickSpacing: tickSpacing100,

				// 200 ticks to the left of current or 2 tick spacings away
				swapTicksAway: -nr1TickSpacingsAway * tickSpacing100,
				// Note, we expect the system to kick as to the left by 1 tick since we cross the NR1 lower tick.
				// The definition of the active range is exclusive of the upper tick and inclusive of the lower tick.
				expectedTickAwayAfterFirstSwap: -nr1TickSpacingsAway*tickSpacing100 - 1,
			},
			"group1 1 tick spacing, first swap cross tick, second swap in same direction": {
				tickSpacing: tickSpacingOne,

				// 2 ticks to the left of current or 2 tick spacings away
				swapTicksAway: -nr1TickSpacingsAway,
				// Note, we expect the system to kick us to the left by 1 tick since we cross the NR1 lower tick.
				// The definition of the active range is exclusive of the upper tick and inclusive of the lower tick.
				expectedTickAwayAfterFirstSwap: -nr1TickSpacingsAway - 1,
			},

			// Group 2:
			// Test setup:
			// swap 1: stop right before lower tick of NR1
			// swap 2: stop right before lower tick of NR1
			"group2 100 tick spacing, first swap does not cross tick, second swap in same direction": {
				tickSpacing: tickSpacing100,

				// two tick spacings to the left + 1 tick from current tick
				swapTicksAway: -nr1TickSpacingsAway*tickSpacing100 + 1,

				// N.B.: rounding takes our current tick to the left by 1 tick.
				// This is acceptable as we stay in the same bucket.
				expectedTickAwayAfterFirstSwap: -nr1TickSpacingsAway * tickSpacing100,
			},
			"group2 1 tick spacing, first swap does not cross tick, second swap in same direction": {
				tickSpacing: 1,

				// two tick spacings to the left + 1 tick from current
				swapTicksAway: -nr1TickSpacingsAway + 1,

				// N.B.: rounding takes our current tick to the left by 1 tick.
				// This is acceptable as we stay in the same bucket.
				expectedTickAwayAfterFirstSwap: -nr1TickSpacingsAway,
			},

			// Group 3:
			// Test setup:
			// swap 1: stop to the right of lower tick of NR1
			// swap 2: stop to the right of lower tick of NR1
			"group3 100 tick spacing, first swap does not cross tick, second swap in same direction": {
				tickSpacing: tickSpacing100,

				// 200 ticks (or 2 tick spacings to the left) + 1 tick from current
				swapTicksAway: -(nr1TickSpacingsAway - 1) * tickSpacing100,

				// N.B.: rounding takes our current tick to the left by 1 tick.
				// This is acceptable as we stay in the same bucket.
				expectedTickAwayAfterFirstSwap: -(nr1TickSpacingsAway-1)*tickSpacing100 - 1,
			},
			"group3 1 tick spacing, first swap does not cross tick, second swap in same direction": {
				tickSpacing: tickSpacingOne,

				// 2 tick s(or 2 tick spacings to the left) + 1 tick from current
				swapTicksAway: -(nr1TickSpacingsAway + 1),

				// N.B.: rounding takes our current tick to the left by 1 tick.
				// This is acceptable as we stay in the same bucket.
				expectedTickAwayAfterFirstSwap: -(nr1TickSpacingsAway + 1) - 1,
			},
		}

		for name, tc := range testCases {
			tc := tc
			s.Run(name, func() {
				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 1. Prepare pool and positions for test

				poolId, nr1Position, nr2Position := setupPoolAndPositions(tc.tickSpacing)

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 2. Estimate the amount of tokenIn to swap in to depending on test configuration.
				// Also, estimated the expected values depending on test configuration.
				expectedAndComputedValues := computeValuesForTestZeroForOne(poolId, tc, nr1Position, nr2Position)

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 3. Run the first swap and validate results.

				// Perform the swap that should cross the next NR1 tick in the direction of the swap.
				s.swapZeroForOneLeft(poolId, expectedAndComputedValues.tokenIn)

				// Perform validations after the first swap.
				validateAfterFirstSwap(
					poolId,
					expectedAndComputedValues,
					nr1Position,
					nr2Position)

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 4. Run the second swap and validate results.

				// If we swap a small amount in the same direction again, we do not expect
				// any tick to be crossed again.
				smallAmount := sdk.NewCoin(expectedAndComputedValues.tokenIn.Denom, sdk.NewInt(10))
				s.swapZeroForOneLeft(poolId, smallAmount)

				// Perform validation after the second swap.
				validateAfterSecondSwap(
					poolId,
					expectedAndComputedValues,
					nr1Position,
					nr2Position,
				)
			})
		}
	})

	s.Run("one for zero", func() {
		testCases := map[string]testCase{

			// Group 1:
			// Test setup:
			// swap 1: just enough to cross upper tick of NR1
			// swap 2: stop between NR1 upper and NR 2 upper
			"group1 100 tick spacing, first swap cross tick, second swap in same direction": {
				tickSpacing: tickSpacing100,

				// 2 tick spacings away from current tick.
				swapTicksAway: nr1TickSpacingsAway * tickSpacing100,

				expectedTickAwayAfterFirstSwap: nr1TickSpacingsAway * tickSpacing100,
			},
			"group1 1 tick spacing, first swap cross tick, second swap in same direction": {
				tickSpacing: tickSpacingOne,

				// 2 ticks (tick spacings) away from current tick.
				swapTicksAway: nr1TickSpacingsAway,

				expectedTickAwayAfterFirstSwap: nr1TickSpacingsAway,
			},

			// Group 2:
			// Test setup:
			// swap 1: stop right before upper tick of NR1
			// swap 2: stop right before upper tick of NR1
			"group2 100 tick spacing, first swap does not cross tick, second swap in same direction": {
				tickSpacing: tickSpacing100,

				// 200 ticks (or 2 tick spacings away) + 1 tick from current
				swapTicksAway: nr1TickSpacingsAway*tickSpacing100 - 1,

				expectedTickAwayAfterFirstSwap: nr1TickSpacingsAway*tickSpacing100 - 1,
			},
			"group2 1 tick spacing, first swap does not cross tick, second swap in same direction": {
				tickSpacing: tickSpacingOne,

				// 2 ticks (or 2 tick spacings away) - 1
				swapTicksAway: nr1TickSpacingsAway*tickSpacingOne - 1,

				expectedTickAwayAfterFirstSwap: nr1TickSpacingsAway*tickSpacingOne - 1,
			},

			// Group 3:
			// Test setup:
			// swap 1: stop right after lower tick of NR1
			// swap 2: stop right before lower tick of NR1
			"group3 100 tick spacing, first swap crosses tick, second swap in same direction": {
				tickSpacing: tickSpacing100,

				// 200 ticks (or 2 tick spacings away) + 1 tick from current
				swapTicksAway: (nr1TickSpacingsAway + 1) * tickSpacing100,

				expectedTickAwayAfterFirstSwap: (nr1TickSpacingsAway + 1) * tickSpacing100,
			},
			"group3 1 tick spacing, first swap crosses tick, second swap in same direction": {
				tickSpacing: tickSpacingOne,

				// 2 ticks (or 2 tick spacings away) + 1 tick from current
				swapTicksAway: nr1TickSpacingsAway + 1,

				expectedTickAwayAfterFirstSwap: nr1TickSpacingsAway + 1,
			},
		}

		for name, tc := range testCases {
			tc := tc
			s.Run(name, func() {

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 1. Prepare pool and positions for test

				poolId, nr1Position, nr2Position := setupPoolAndPositions(tc.tickSpacing)

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 2. Estimate the amount of token 1 to swap to depending on test configuration.
				expectedAndComputedValues := computeExpectedValuesForTestOneForZero(poolId, tc, nr1Position, nr2Position)

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 3. Run the first swap and validate results.

				// Perform the swap that should cross the next NR1 tick in the direction of the swap.
				s.swapOneForZeroRight(poolId, expectedAndComputedValues.tokenIn)

				// Perform validations after the first swap.
				validateAfterFirstSwap(
					poolId,
					expectedAndComputedValues,
					nr1Position,
					nr2Position)

				////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
				// 4. Run the second swap and validate results.

				// If we swap a small amount in the same direction again, we do not expect
				// any tick to be crossed again.
				smallAmount := sdk.NewCoin(expectedAndComputedValues.tokenIn.Denom, sdk.NewInt(10000))
				s.swapOneForZeroRight(poolId, smallAmount)

				// Perform validation after the second swap.
				validateAfterSecondSwap(
					poolId,
					expectedAndComputedValues,
					nr1Position,
					nr2Position,
				)
			})
		}
	})
}
