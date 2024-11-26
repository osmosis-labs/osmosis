package concentrated_liquidity_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/swapstrategy"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

const (
	tickSpacingOne = 1
	tickSpacing100 = 100
)

// This configuration is expected to generate the position layout below:
//
//	                                 original_cur_tick
//		                                    ///               (NR4)
//		                                /////////             (NR3)
//		                            ////////////////          (NR2)
//	                             //////////////////////       (NR1)
//
// min tick                                                      max tick
var (
	defaultTickSpacingsAway = []uint64{4, 3, 2, 1}

	// this is chosen arbitrarily to allow tests to pass. Thee tests in this suite do not
	// intend to validate the correctness of the slippage bound. As a result, it is irrelevant here
	// and we can choose any value that works.
	defaultTokenInMaxAmount = osmomath.MustNewDecFromStr("707106781186547528576662335").TruncateInt()
)

// CreatePositionTickSpacingsFromCurrentTick creates a position with the passed in tick spacings away from the current tick.
func (s *KeeperTestSuite) CreatePositionTickSpacingsFromCurrentTick(poolId uint64, tickSpacingsAwayFromCurrentTick uint64) positionMeta {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	currentTick := pool.GetCurrentTick()

	tickSpacing := int64(pool.GetTickSpacing())

	// make sure that current tick is a multiple of tick spacing
	currentTick = currentTick - (currentTick % tickSpacing)

	lowerTick := currentTick - int64(tickSpacingsAwayFromCurrentTick)*tickSpacing
	upperTick := currentTick + int64(tickSpacingsAwayFromCurrentTick)*tickSpacing
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, pool.GetId(), s.TestAccs[0], DefaultCoins, osmomath.ZeroInt(), osmomath.ZeroInt(), lowerTick, upperTick)
	s.Require().NoError(err)

	return positionMeta{
		positionId: positionData.ID,
		lowerTick:  lowerTick,
		upperTick:  upperTick,
		liquidity:  positionData.Liquidity,
	}
}

// tickToSqrtPrice a helper to convert a tick to a sqrt price.
func (s *KeeperTestSuite) tickToSqrtPrice(tick int64) osmomath.BigDec {
	sqrtPrice, err := math.TickToSqrtPrice(tick)
	s.Require().NoError(err)
	return sqrtPrice
}

// validateIteratorLeftZeroForOne is a helper to validate the next initialized tick iterator
// in the left (zfo) direction of the swap.
func (s *KeeperTestSuite) validateIteratorLeftZeroForOne(poolId uint64, expectedTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	zeroForOneSwapStrategy, _, err := s.App.ConcentratedLiquidityKeeper.SetupSwapStrategy(s.Ctx, pool, osmomath.ZeroDec(), pool.GetToken0(), types.MinSqrtPriceBigDec)
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
func (s *KeeperTestSuite) validateIteratorRightOneForZero(poolId uint64, expectedTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	// Setup swap strategy directly as it would fail validation if constructed via SetupSwapStrategy(...)
	oneForZeroSwapStrategy := swapstrategy.New(false, osmomath.BigDecFromDec(types.MaxSqrtPrice), s.App.GetKey(types.ModuleName), osmomath.ZeroDec())
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
func (s *KeeperTestSuite) assertPositionInRange(poolId uint64, lowerTick int64, upperTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	isInRange := pool.IsCurrentTickInRange(lowerTick, upperTick)
	s.Require().True(isInRange, "currentTick: %d, lowerTick %d, upperTick: %d", pool.GetCurrentTick(), lowerTick, upperTick)
}

// assertPositionOutOfRange a helper to assert that a position with the given lowerTick and upperTick is out of range.
func (s *KeeperTestSuite) assertPositionOutOfRange(poolId uint64, lowerTick int64, upperTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	isInRange := pool.IsCurrentTickInRange(lowerTick, upperTick)
	s.Require().False(isInRange, "currentTick: %d, lowerTick %d, upperTick: %d", pool.GetCurrentTick(), lowerTick, upperTick)
}

// assertPositionRangeConditional a helper to assert that a position with the given lowerTick and upperTick is in or out of range
// depending on the isOutOfRangeExpected flag.
func (s *KeeperTestSuite) assertPositionRangeConditional(poolId uint64, isOutOfRangeExpected bool, lowerTick int64, upperTick int64) {
	if isOutOfRangeExpected {
		s.assertPositionOutOfRange(poolId, lowerTick, upperTick)
	} else {
		s.assertPositionInRange(poolId, lowerTick, upperTick)
	}
}

// swapZeroForOneLeft swaps amount in the left (zfo) direction of the swap.
// Asserts that no error is returned.
func (s *KeeperTestSuite) swapZeroForOneLeft(poolId uint64, amount sdk.Coin) {
	s.swapZeroForOneLeftWithSpread(poolId, amount, osmomath.ZeroDec())
}

// swapZeroForOneLeftWithSpread functions exactly as swapZeroForOneLeft but with a spread factor.
func (s *KeeperTestSuite) swapZeroForOneLeftWithSpread(poolId uint64, amount sdk.Coin, spreadFactor osmomath.Dec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(amount))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, amount, pool.GetToken1(), osmomath.ZeroInt(), spreadFactor)
	s.Require().NoError(err)
}

// swapOneForZeroRight swaps amount in the right (ofz) direction of the swap.
// Asserts that no error is returned.
func (s *KeeperTestSuite) swapOneForZeroRight(poolId uint64, amount sdk.Coin) {
	s.swapOneForZeroRightWithSpread(poolId, amount, osmomath.ZeroDec())
}

// swapOneForZeroRightWithSpread functions exactly as swapOneForZeroRight but with a spread factor.
func (s *KeeperTestSuite) swapOneForZeroRightWithSpread(poolId uint64, amount sdk.Coin, spreadFactor osmomath.Dec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(amount))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, amount, pool.GetToken0(), osmomath.ZeroInt(), spreadFactor)
	s.Require().NoError(err)
}

// swapInGivenOutZeroForOneLeft swaps in given out in the left (zfo) direction of the swap.
// Asserts that no error is returned.
// When swapping in given out, we provide token to swap out but eventually get charged token in.
// Therefore we must also estimate the token in amount and pre-fund the account with it.
func (s *KeeperTestSuite) swapInGivenOutZeroForOneLeft(poolId uint64, tokenOut sdk.Coin, estimatedTokenIn osmomath.Dec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	tokenInDenom := pool.GetToken0()
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(tokenInDenom, estimatedTokenIn.Ceil().TruncateInt())))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool, tokenInDenom, defaultTokenInMaxAmount, tokenOut, osmomath.ZeroDec())
	s.Require().NoError(err)
}

// swapInGivenOutOneForZeroRight swaps in given out in the right (ofz) direction of the swap.
// When swapping in given out, we provide token to swap out but eventually get charged token in.
// Therefore we must also estimate the token in amount and pre-fund the account with it.
func (s *KeeperTestSuite) swapInGivenOutOneForZeroRight(poolId uint64, tokenOut sdk.Coin, estimatedTokenIn osmomath.Dec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	tokenInDenom := pool.GetToken1()
	s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(tokenInDenom, estimatedTokenIn.Ceil().TruncateInt())))
	_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountOut(s.Ctx, s.TestAccs[0], pool, tokenInDenom, defaultTokenInMaxAmount, tokenOut, osmomath.ZeroDec())
	s.Require().NoError(err)
}

// setupPoolAndPositions creates a pool and the following positions:
// - full range
// - for every t in positionTickSpacingsFromCurrTick, it creates a narrow range position
// t tick spacings away from the current tick.
//
// Returns the pool id and the narrow range position metadata.
func (s *KeeperTestSuite) setupPoolAndPositions(testTickSpacing uint64, positionTickSpacingsFromCurrTick []uint64, initialCoins sdk.Coins) (uint64, []positionMeta) {
	pool := s.PrepareCustomConcentratedPool(s.TestAccs[0], ETH, USDC, testTickSpacing, osmomath.ZeroDec())
	poolId := pool.GetId()

	// Create a full range position
	s.FundAcc(s.TestAccs[0], DefaultCoins)
	positionData, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolId, s.TestAccs[0], initialCoins)
	s.Require().NoError(err)

	// Refetch pool as the first position updated its state.
	pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, pool.GetId())
	s.Require().NoError(err)

	// Create all narrow range positions per given tick spacings away from the current tick
	// configuration.
	positionMetas := make([]positionMeta, len(positionTickSpacingsFromCurrTick))
	liquidityAllPositions := positionData.Liquidity
	for i, tickSpacingsAway := range positionTickSpacingsFromCurrTick {
		// Create narrow range position tickSpacingsAway from the current tick
		positionMetas[i] = s.CreatePositionTickSpacingsFromCurrentTick(poolId, tickSpacingsAway)

		// Update total liquidity
		liquidityAllPositions = liquidityAllPositions.Add(positionMetas[i].liquidity)

		// Sanity check that the created position is in range
		s.assertPositionInRange(poolId, positionMetas[i].lowerTick, positionMetas[i].upperTick)
	}

	// As a sanity check confirm that current liquidity corresponds
	// to the sum of liquidities of all positions.
	s.assertPoolLiquidityEquals(poolId, liquidityAllPositions)

	return poolId, positionMetas
}

// assertPoolLiquidityEquals a helper to assert that the liquidity of a pool is equal to the expected value.
func (s *KeeperTestSuite) assertPoolLiquidityEquals(poolId uint64, expectedLiquidity osmomath.Dec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.Require().Equal(expectedLiquidity.String(), pool.GetLiquidity().String())
}

// assertPoolTickEquals a helper to assert that the current tick of a pool is equal to the expected value.
func (s *KeeperTestSuite) assertPoolTickEquals(poolId uint64, expectedTick int64) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	s.Require().Equal(expectedTick, pool.GetCurrentTick())
}

// computeSwapAmounts computes the amountIn that should be swapped to reach the expectedTickToSwapTo
// given the direction of the swap (as defined by isZeroForOne) and the current sqrt price.
// if shouldStayWithinTheSameBucket is true, the amountIn is computed such that the swap does not cross the tick.
// curSqrtPrice can be a nil dec (osmomath.Dec{}). In such a case, the system converts the current tick to a current sqrt price.
// The reason why user might want to provide a current sqrt price is when going in zero for one direction of a second swap.
// In that case, the current sqrt price is still in the domain of the previous bucket but the current tick is already in the next
// bucket.
//
// Note, that this logic runs quote estimation. Our frontend logic runs a similar algorithm.
func (s *KeeperTestSuite) computeSwapAmounts(poolId uint64, curSqrtPrice osmomath.BigDec, expectedTickToSwapTo int64, isZeroForOne bool, shouldStayWithinTheSameBucket bool) (osmomath.Dec, osmomath.Dec, osmomath.BigDec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	originalCurrentTick := pool.GetCurrentTick()

	tokenInDenom := pool.GetToken0()
	if !isZeroForOne {
		tokenInDenom = pool.GetToken1()
	}

	// Get liquidity net amounts for tokenIn estimation.
	liquidityNetAmounts, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityNetInDirection(s.Ctx, poolId, tokenInDenom, osmomath.Int{}, osmomath.Int{})
	s.Require().NoError(err)

	currentTick := originalCurrentTick
	// compute current sqrt price if not provided
	if curSqrtPrice.IsNil() {
		curSqrtPrice = s.tickToSqrtPrice(currentTick)
	}

	// Start from current pool liquidity and zero amount in.
	currentLiquidity := pool.GetLiquidity()
	amountIn := osmomath.ZeroDec()

	for i, liquidityNetEntry := range liquidityNetAmounts {
		// Initialize the next initialized tick and its sqrt price.
		nextInitializedTick := liquidityNetEntry.TickIndex
		nextInitTickSqrtPrice := s.tickToSqrtPrice(nextInitializedTick)

		// Handle swap depending on the direction.
		// Left (zero for one) or right (one for zero)
		var isWithinDesiredBucketAfterSwap bool
		if isZeroForOne {
			// Round up so that we cross the tick by default.
			curAmountIn := math.CalcAmount0Delta(currentLiquidity, curSqrtPrice, nextInitTickSqrtPrice, true).DecRoundUp()

			amountIn = amountIn.Add(curAmountIn)

			// The tick should be crossed if currentTick > expectedTickToSwapTo, unless the intention
			// is to stay within the same bucket.
			shouldCrossTick := currentTick > expectedTickToSwapTo && !shouldStayWithinTheSameBucket
			if shouldCrossTick {
				// Runs regular tick crossing logic.
				curSqrtPrice = s.tickToSqrtPrice(nextInitializedTick)
				currentLiquidity = currentLiquidity.Sub(liquidityNetEntry.LiquidityNet)
				currentTick = nextInitializedTick - 1
			}

			// Determine if we've reached the desired bucket.
			isWithinDesiredBucketAfterSwap = currentTick == expectedTickToSwapTo || shouldStayWithinTheSameBucket

			// This in an edge case when going left in second swap after previously going right
			// and indetending to stay within the same bucket.
			if amountIn.IsZero() && isWithinDesiredBucketAfterSwap {
				nextInitTickSqrtPrice := s.tickToSqrtPrice(liquidityNetAmounts[i+1].TickIndex)

				// We discount by half so that we do no cross any tick and remain in the same bucket.
				curAmountIn := math.CalcAmount0Delta(currentLiquidity, curSqrtPrice, nextInitTickSqrtPrice, true).QuoInt64(2).DecRoundUp()
				amountIn = amountIn.Add(curAmountIn)
			}
		} else {
			// Round up so that we cross the tick by default.
			curAmountIn := math.CalcAmount1Delta(currentLiquidity, curSqrtPrice, nextInitTickSqrtPrice, true).Dec()
			amountIn = amountIn.Add(curAmountIn)

			// The tick should be crossed if currentTick <= expectedTickToSwapTo, unless the intention
			// is to stay within the same bucket.
			shouldCrossTick := currentTick <= expectedTickToSwapTo && !shouldStayWithinTheSameBucket
			if shouldCrossTick {
				// Runs regular tick crossing logic.
				curSqrtPrice = s.tickToSqrtPrice(nextInitializedTick)
				currentLiquidity = currentLiquidity.Add(liquidityNetEntry.LiquidityNet)
				currentTick = nextInitializedTick
			}

			// Determine if we've reached the desired bucket.
			isWithinDesiredBucketAfterSwap = currentTick == expectedTickToSwapTo || shouldStayWithinTheSameBucket
		}

		// Stop if the desired bucket is activated.
		if isWithinDesiredBucketAfterSwap {
			break
		}
	}
	return amountIn, currentLiquidity, curSqrtPrice
}

func (s *KeeperTestSuite) computeSwapAmountsInGivenOut(poolId uint64, curSqrtPrice osmomath.BigDec, expectedTickToSwapTo int64, isZeroForOne bool, shouldStayWithinTheSameBucket bool) (osmomath.Dec, osmomath.Dec, osmomath.BigDec) {
	pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	originalCurrentTick := pool.GetCurrentTick()

	tokenOutDenom := pool.GetToken0()
	if !isZeroForOne {
		tokenOutDenom = pool.GetToken1()
	}

	// Get liquidity net amounts for tokenIn estimation.
	liquidityNetAmounts, err := s.App.ConcentratedLiquidityKeeper.GetTickLiquidityNetInDirection(s.Ctx, poolId, tokenOutDenom, osmomath.Int{}, osmomath.Int{})
	s.Require().NoError(err)

	currentTick := originalCurrentTick
	// compute current sqrt price if not provided
	if curSqrtPrice.IsNil() {
		curSqrtPrice = s.tickToSqrtPrice(currentTick)
	}

	// Start from current pool liquidity and zero amount in.
	currentLiquidity := pool.GetLiquidity()
	amountOut := osmomath.ZeroDec()

	for i, liquidityNetEntry := range liquidityNetAmounts {
		// Initialize the next initialized tick and its sqrt price.
		nextInitializedTick := liquidityNetEntry.TickIndex
		nextInitTickSqrtPrice := s.tickToSqrtPrice(nextInitializedTick)

		// Handle swap depending on the direction.
		// Left (zero for one) or right (one for zero)
		var isWithinDesiredBucketAfterSwap bool
		if isZeroForOne {
			// Round up so that we cross the tick by default.
			curAmountOut := math.CalcAmount1Delta(currentLiquidity, curSqrtPrice, nextInitTickSqrtPrice, false)

			amountOut = amountOut.Add(curAmountOut.Dec())

			// The tick should be crossed if currentTick > expectedTickToSwapTo, unless the intention
			// is to stay within the same bucket.
			shouldCrossTick := currentTick > expectedTickToSwapTo && !shouldStayWithinTheSameBucket
			if shouldCrossTick {
				// Runs regular tick crossing logic.
				curSqrtPrice = s.tickToSqrtPrice(nextInitializedTick)
				currentLiquidity = currentLiquidity.Sub(liquidityNetEntry.LiquidityNet)
				currentTick = nextInitializedTick - 1
			}

			// Determine if we've reached the desired bucket.
			isWithinDesiredBucketAfterSwap = currentTick == expectedTickToSwapTo || shouldStayWithinTheSameBucket

			// This in an edge case when going left in second swap after previously going right
			// and indetending to stay within the same bucket.
			if amountOut.IsZero() && isWithinDesiredBucketAfterSwap {
				nextInitTickSqrtPrice := s.tickToSqrtPrice(liquidityNetAmounts[i+1].TickIndex)

				// We discound by two so that we do no cross any tick and remain in the same bucket.
				curAmountIn := math.CalcAmount1Delta(currentLiquidity, curSqrtPrice, nextInitTickSqrtPrice, false).QuoInt64(2)
				amountOut = amountOut.Add(curAmountIn.DecRoundUp())
			}
		} else {
			// Round up so that we cross the tick by default.
			curAmountOut := math.CalcAmount0Delta(currentLiquidity, curSqrtPrice, nextInitTickSqrtPrice, false)
			amountOut = amountOut.Add(curAmountOut.Dec())

			// The tick should be crossed if currentTick <= expectedTickToSwapTo, unless the intention
			// is to stay within the same bucket.
			shouldCrossTick := currentTick <= expectedTickToSwapTo && !shouldStayWithinTheSameBucket
			if shouldCrossTick {
				// Runs regular tick crossing logic.
				curSqrtPrice = s.tickToSqrtPrice(nextInitializedTick)
				currentLiquidity = currentLiquidity.Add(liquidityNetEntry.LiquidityNet)
				currentTick = nextInitializedTick
			}

			// Determine if we've reached the desired bucket.
			isWithinDesiredBucketAfterSwap = currentTick == expectedTickToSwapTo || shouldStayWithinTheSameBucket
		}

		// Stop if the desired bucket is activated.
		if isWithinDesiredBucketAfterSwap {
			break
		}
	}
	return amountOut, currentLiquidity, curSqrtPrice
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
// Additionally, it sets up a case of swapping in one for zero direction to swap directly at the next initialized
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
func (s *KeeperTestSuite) TestSwapOutGivenIn_Tick_Initialization_And_Crossing() {
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
		expectedLiquidityAfterFirstAndSecondSwap osmomath.Dec
		doesFirstSwapCrossTick                   bool
		expectedNextTickAfterFirstSwapZFOLeft    int64
		expectedNextTickAfterFirstSwapOFZRight   int64
	}

	const (
		// Defines how many tick spacings away from the current
		// tick NR1 and NR2 are. Note that lower and upper ticks
		// are equally spaced around the current tick.
		nr1TickSpacingsAway = 2
		nr2TickSpacingsAway = 5
	)

	var (
		desiredPositionTickSpacingsAway = []uint64{nr1TickSpacingsAway, nr2TickSpacingsAway}
	)

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

		// First position is out if range if doesFirstSwapCrossUpperTickOne is true. Otherwise, it is in range.
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
			amountZeroIn   osmomath.Dec    = osmomath.ZeroDec()
			sqrtPriceStart osmomath.BigDec = pool.GetCurrentSqrtPrice()

			liquidity = pool.GetLiquidity()
		)

		if tickToSwapTo < nr1Position.lowerTick {
			sqrtPriceLowerTickOne := s.tickToSqrtPrice(nr1Position.lowerTick)

			amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceLowerTickOne, sqrtPriceStart, true).Dec()

			sqrtPriceStart = sqrtPriceLowerTickOne

			liquidity = liquidity.Sub(nr1Position.liquidity)
		}

		// This is the total amount necessary to cross the lower tick of narrow position.
		// Note it is rounded up to ensure that the tick is crossed.
		amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceStart, true).DecRoundUp().Add(amountZeroIn)

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
			expectedLiquidityAfterFirstAndSecondSwap osmomath.Dec
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
			amountOneIn    osmomath.Dec    = osmomath.ZeroDec()
			sqrtPriceStart osmomath.BigDec = pool.GetCurrentSqrtPrice()
			liquidity                      = pool.GetLiquidity()
		)

		if tickToSwapTo >= nr1Position.upperTick {
			sqrtPriceUpperOne := s.tickToSqrtPrice(nr1Position.upperTick)

			amountOneIn = math.CalcAmount1Delta(liquidity, sqrtPriceUpperOne, sqrtPriceStart, true).DecRoundUp()

			sqrtPriceStart = sqrtPriceUpperOne

			liquidity = liquidity.Sub(nr1Position.liquidity)
		}

		// This is the total amount necessary to cross the lower tick of narrow position.
		// Note it is rounded up to ensure that the tick is crossed.
		amountOneIn = math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceStart, true).DecRoundUp().Add(amountOneIn)

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

				poolId, positionMetas := s.setupPoolAndPositions(tc.tickSpacing, desiredPositionTickSpacingsAway, DefaultCoins)
				nr1Position := positionMetas[0]
				nr2Position := positionMetas[1]

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
				smallAmount := sdk.NewCoin(expectedAndComputedValues.tokenIn.Denom, osmomath.NewInt(10))
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

				poolId, positionMetas := s.setupPoolAndPositions(tc.tickSpacing, desiredPositionTickSpacingsAway, DefaultCoins)
				nr1Position := positionMetas[0]
				nr2Position := positionMetas[1]

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
				smallAmount := sdk.NewCoin(expectedAndComputedValues.tokenIn.Denom, osmomath.NewInt(10000))
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

// TestSwaps_Contiguous_Initialized_TickSpacingOne tests swapping multiple times in various directions
// when there are contiguous ticks initialized on a swap range in a pool with tick spacing of 1.
// For position layout, see diagram above the definition of defaultTickSpacingsAway variable.
// For specific test vectors, follow the table-driven names below.
// It uses both swap in given out and swap out given in methods.
func (s *KeeperTestSuite) TestSwaps_Contiguous_Initialized_TickSpacingOne() {
	// defines an individual test case
	type continugousTestCase struct {
		// This defines how many ticks away from current the swap should reach
		// negative values indicate ticks below current tick
		// positive values indicate ticks above current tick
		swapEndTicksAwayFromOriginalCurrent []int64
		// This flag is used to control an edge case behavior in test setup
		// when swapping right and then right again within the same tick.
		// It is only used to signal the swap direction while estimating the expected results.
		isOneForZeroWithinSameTick bool

		// isPositionActiveFlag is used to control the expected state of position
		// at the end of configured swaps of the test.
		// See diagram above the definition of defaultTickSpacingsAway variable for layout.
		// The first position is NR1, the second position is NR2 etc.
		// That is, the wider range position is precedes the narrower range position.
		isPositionActiveFlag []bool
	}

	// computeExpectedTicks estimates the expected tick values from the given test case
	// and returns them as tick slice.
	computeExpectedTicks := func(poolId uint64, tc continugousTestCase) []int64 {
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		originalCurrentTick := pool.GetCurrentTick()

		// Convert the given "ticks away from original current" to actual tick indexes.

		expectedSwapEndTicks := make([]int64, len(tc.swapEndTicksAwayFromOriginalCurrent))

		for i, swapTicksAway := range tc.swapEndTicksAwayFromOriginalCurrent {
			// We special case the min and max tick to be given by absolute values
			// rather than relative to the original current tick.
			if swapTicksAway == types.MinInitializedTick {
				// In zero for one direction, we kick the tick back by one while crossing.
				// This is to ensure that our definition of "active bucket" is correct.
				expectedSwapEndTicks[i] = types.MinCurrentTick
			} else if swapTicksAway == types.MaxTick {
				expectedSwapEndTicks[i] = types.MaxTick
			} else {
				expectedSwapEndTicks[i] = originalCurrentTick + swapTicksAway
			}
		}

		return expectedSwapEndTicks
	}

	// computeNextTickToReachAndMultiplier returns the tick to swap to during estimate computation and amountIn multiplier.
	// It most cases, the tick to swap to is the same as the expected tick to reach after the swap and the multiplier is 1.
	// The only exception is when performing a second swap in the same direction within the same tick.
	// In such a case, we need to run our estimate logic one tick further to ensure that our estimate is non-zero.
	// However, we discount the amountIn by half to ensure that the tick is not crossed.
	computeNextTickToReachAndMultiplier := func(isZeroForOne bool, expectedSwapEndTick int64, shouldStayWithinTheSameTickInCompute bool) (int64, osmomath.Dec) {
		if shouldStayWithinTheSameTickInCompute {
			nextTickToReachInCompute := expectedSwapEndTick
			if isZeroForOne {
				nextTickToReachInCompute = nextTickToReachInCompute - 1
			} else {
				nextTickToReachInCompute = nextTickToReachInCompute + 1
			}

			return nextTickToReachInCompute, osmomath.NewDecWithPrec(5, 1)
		}

		return expectedSwapEndTick, osmomath.OneDec()
	}

	validateActivePositions := func(poolId uint64, positionMeta []positionMeta, expectedIsPositionActiveFlags []bool) {
		// Refetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		// Validate the positions
		s.Require().NotEmpty(expectedIsPositionActiveFlags)
		for i, expectedActivePositionIndex := range expectedIsPositionActiveFlags {

			isInRange := pool.IsCurrentTickInRange(positionMeta[i].lowerTick, positionMeta[i].upperTick)
			s.Require().Equal(expectedActivePositionIndex, isInRange, fmt.Sprintf("position %d", i))
		}
	}

	// Note, that we use the same test cases for both kinds of swaps.
	testcases := map[string]continugousTestCase{
		"zero for one, swap to the middle tick to the left of the original current, then swap again to the leftmost tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{-3, -4},
			isPositionActiveFlag:                []bool{true, false, false, false},
		},
		"zero for one, swap to the middle tick to the left of the original current, then swap again to the rightmost tick smaller than the original current": {
			swapEndTicksAwayFromOriginalCurrent: []int64{-3, -1},
			isPositionActiveFlag:                []bool{true, true, true, true},
		},
		"zero for one, swap to the middle tick to the left of the original current and swap again but stay within the same initialized tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{-3, -3},
			isPositionActiveFlag:                []bool{true, true, false, false},
		},
		"zero for one, swap to the middle tick to the left of the original current and then swap all the way back to the right of the original current tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{-3, 1},
			isPositionActiveFlag:                []bool{true, true, true, false},
		},
		"zero for one, swap beyond the leftmost tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{-3, types.MinInitializedTick},
			isPositionActiveFlag:                []bool{false, false, false, false},
		},

		"one for zero, swap to the middle tick to the left of the original current, then swap again to the leftmost tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{2, 3},
			isPositionActiveFlag:                []bool{true, false, false, false},
		},
		"one for zero, swap to the middle tick to the left of the original current, then swap again to the rightmost tick smaller than the original current": {
			swapEndTicksAwayFromOriginalCurrent: []int64{2, 1},
			isPositionActiveFlag:                []bool{true, true, true, false},
		},
		"one for zero, swap to the middle tick to the left of the original current and swap again but stay within the same initialized tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{2, 2},
			isOneForZeroWithinSameTick:          true,
			isPositionActiveFlag:                []bool{true, true, false, false},
		},
		"one for zero, swap to the middle tick to the left of the original current and then swap all the way back to the right of the original current tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{2, -2},
			isPositionActiveFlag:                []bool{true, true, true, false},
		},
		"one for zero, swap beyond the rightmost tick": {
			swapEndTicksAwayFromOriginalCurrent: []int64{2, types.MaxTick},
			isPositionActiveFlag:                []bool{false, false, false, false},
		},
	}

	s.Run("swap out given in", func() {
		for name, tc := range testcases {
			s.Run(name, func() {
				s.SetupTest()

				poolId, positionMeta := s.setupPoolAndPositions(tickSpacingOne, defaultTickSpacingsAway, DefaultCoins)
				s.Require().Equal(len(tc.isPositionActiveFlag), len(positionMeta))

				// Refetch pool.
				pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
				s.Require().NoError(err)

				// Compute expected ticks from test case configuration.
				expectedSwapEndTicks := computeExpectedTicks(poolId, tc)

				// Determine if swaps within the same tick are expected.
				// This is so that we discount the last swap by 50% to make
				// sure to remain within the same tick and never cross.
				isWithinTheSameTick := osmoutils.ContainsDuplicate(expectedSwapEndTicks)

				// Perform the swaps
				curSqrtPrice := osmomath.BigDec{}
				curTick := pool.GetCurrentTick()
				// For every expected, swap, estimate the amount in needed
				// to reach it and run validations after the swap.
				for i, expectedSwapEndTick := range expectedSwapEndTicks {
					// Determine if we are swapping zero for one or one for zero.
					isZeroForOne := expectedSwapEndTick <= curTick && !tc.isOneForZeroWithinSameTick

					// This is used to control the computations for the last swap when swapping
					// See definition of computeNextTickToReachAndMultiplier for more details.
					shouldStayWithinTheSameTickInCompute := isWithinTheSameTick && i == len(expectedSwapEndTicks)-1
					nextTickToReachInCompute, withinTheSameTickDiscount := computeNextTickToReachAndMultiplier(isZeroForOne, expectedSwapEndTick, shouldStayWithinTheSameTickInCompute)

					// Estimate the amountIn necessary to reach the expected swap end tick, the expected liquidity and
					// the "current sqrt price" for next swap.
					amountIn, expectedLiquidity, nextSqrtPrice := s.computeSwapAmounts(poolId, curSqrtPrice, nextTickToReachInCompute, isZeroForOne, shouldStayWithinTheSameTickInCompute)

					// Discount the amount in by 50% if we are swapping within the same tick.
					amountInRoundedUp := amountIn.Mul(withinTheSameTickDiscount).Ceil().TruncateInt()

					// Perform the swap in the desired direction.
					if isZeroForOne {
						amountInCoin := sdk.NewCoin(pool.GetToken0(), amountInRoundedUp)
						s.swapZeroForOneLeft(poolId, amountInCoin)
					} else {
						amountInCoin := sdk.NewCoin(pool.GetToken1(), amountInRoundedUp)
						s.swapOneForZeroRight(poolId, amountInCoin)
					}

					// Validate that current tick and current liquidity are as expected.
					s.assertPoolTickEquals(poolId, expectedSwapEndTick)
					s.assertPoolLiquidityEquals(poolId, expectedLiquidity)

					// Update the current sqrt price and tick for next swap.
					curSqrtPrice = nextSqrtPrice
					curTick = expectedSwapEndTick
				}

				// Validate active positions.
				validateActivePositions(poolId, positionMeta, tc.isPositionActiveFlag)
			})
		}
	})

	s.Run("swap in given out", func() {
		// estimateAmountInFromRounding is a helper to estimate the impact of amountOut rounding on the amountIn and next sqrt price.
		// This is necessary for correct amount in estimation to pre-fund the swapper account to. It is also required for updating
		// the "current sqrt price" for the next swap in the sequence as defined by our test configuration.
		// TODO: Change type arg of liq
		estimateAmountInFromRounding := func(isZeroForOne bool, nextSqrtPrice osmomath.BigDec, liq osmomath.BigDec, amountOutDifference osmomath.BigDec) (osmomath.Dec, osmomath.BigDec) {
			if !liq.IsPositive() {
				return osmomath.ZeroDec(), nextSqrtPrice
			}

			if isZeroForOne {
				// Round down since we want to overestimate the change in sqrt price stemming from the amount out going right-to-left
				// from the current sqrt price.  This overestimated value is then used to calculate amount in charged on the user.
				// Since amount in is overestimated, this done in favor of the pool.
				updatedNextCurSqrtPrice := math.GetNextSqrtPriceFromAmount1OutRoundingDown(nextSqrtPrice, liq.Dec(), amountOutDifference)
				// Round up since we want to overestimate the amount in in favor of the pool.
				return math.CalcAmount0Delta(liq.Dec(), updatedNextCurSqrtPrice, nextSqrtPrice, true).DecRoundUp(), updatedNextCurSqrtPrice
			}

			// Round up since we want to overestimate the change in sqrt price stemming from the amount out going left-to-right
			// from the current sqrt price. This overestimated value is then used to calculate amount in charged on the user.
			// Since amount in is overestimated, this is done in favor of the pool.
			updatedNextCurSqrtPrice := math.GetNextSqrtPriceFromAmount0OutRoundingUp(nextSqrtPrice, liq, amountOutDifference.Dec())
			// Round up since we want to overestimate the amount in in favor of the pool.
			return math.CalcAmount1Delta(liq.Dec(), updatedNextCurSqrtPrice, nextSqrtPrice, true).DecRoundUp(), updatedNextCurSqrtPrice
		}

		for name, tc := range testcases {
			s.Run(name, func() {
				s.SetupTest()

				poolId, positionMeta := s.setupPoolAndPositions(tickSpacingOne, defaultTickSpacingsAway, DefaultCoins)
				s.Require().Equal(len(tc.isPositionActiveFlag), len(positionMeta))

				// Refetch pool.
				pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
				s.Require().NoError(err)

				// Compute expected ticks from test case configuration.
				expectedSwapEndTicks := computeExpectedTicks(poolId, tc)

				// Determine if swaps within the same tick are expected.
				// This is so that we discount the last swap by 50% to make
				// sure to remain within the same tick and never cross.
				isWithinTheSameTick := osmoutils.ContainsDuplicate(expectedSwapEndTicks)

				// Perform the swaps
				curSqrtPrice := osmomath.BigDec{}
				curTick := pool.GetCurrentTick()
				// For every expected, swap, estimate the amount in needed
				// to reach it and run validations after the swap.
				for i, expectedSwapEndTick := range expectedSwapEndTicks {
					// Determine if we are swapping zero for one or one for zero.
					isZeroForOne := expectedSwapEndTick <= curTick && !tc.isOneForZeroWithinSameTick

					// This is used to control the computations for the last swap when swapping
					// See definition of computeNextTickToReachAndMultiplier for more details.
					shouldStayWithinTheSameTickInCompute := isWithinTheSameTick && i == len(expectedSwapEndTicks)-1
					nextTickToReachInCompute, withinTheSameTickDiscount := computeNextTickToReachAndMultiplier(isZeroForOne, expectedSwapEndTick, shouldStayWithinTheSameTickInCompute)

					tokenInDenom := pool.GetToken0()
					if !isZeroForOne {
						tokenInDenom = pool.GetToken1()
					}

					// Estimate the amountOut necessary to reach the expected swap end tick, the expected liquidity and
					// the "current sqrt price" for next swap.
					amountOut, expectedLiquidity, nextSqrtPrice := s.computeSwapAmountsInGivenOut(poolId, curSqrtPrice, nextTickToReachInCompute, isZeroForOne, shouldStayWithinTheSameTickInCompute)
					amountIn, _, _ := s.computeSwapAmounts(poolId, curSqrtPrice, nextTickToReachInCompute, isZeroForOne, shouldStayWithinTheSameTickInCompute)

					s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(tokenInDenom, amountIn.Ceil().TruncateInt().Add(osmomath.OneInt()))))

					// Discount the amount in by 50% if we are swapping within the same tick.
					amountOutRoundedUp := amountOut.Mul(withinTheSameTickDiscount).Ceil().TruncateInt()

					// Notice how we round the amount out up above. This causes issues in one for zero direction as this rounding
					// makes us go further than expected with swap out given in. To compensate for this, we estimate the
					// impact of the rounding to correctly update the "curSqrtPrice" at the end of this loop for
					// properly estimating the next swap. This also allows us to precisely calculate by how many tokens in we need
					// to pre-fund the swapper account.
					amountOutDifference := amountOutRoundedUp.ToLegacyDec().Sub(amountOut)
					liqBigDec := osmomath.BigDecFromDec(expectedLiquidity) // TODO: Delete
					amountInFromRounding, updatedNextCurSqrtPrice := estimateAmountInFromRounding(isZeroForOne, nextSqrtPrice, liqBigDec, osmomath.BigDecFromDec(amountOutDifference))
					amountInToPreFund := amountIn.Add(amountInFromRounding)

					// Perform the swap in the desired direction.
					if isZeroForOne {
						amountOutCoin := sdk.NewCoin(pool.GetToken1(), amountOutRoundedUp)
						s.swapInGivenOutZeroForOneLeft(poolId, amountOutCoin, amountInToPreFund)
					} else {
						// s.FundAcc(s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(tokenInDenom, osmomath.NewInt(amountInFromRounding.Ceil().TruncateInt64()))))

						amountInCoin := sdk.NewCoin(pool.GetToken0(), amountOutRoundedUp)
						s.swapInGivenOutOneForZeroRight(poolId, amountInCoin, amountInToPreFund)
					}

					// Validate that current tick and current liquidity are as expected.
					s.assertPoolTickEquals(poolId, expectedSwapEndTick)
					s.assertPoolLiquidityEquals(poolId, expectedLiquidity)

					// Update the current sqrt price and tick for next swap.
					curSqrtPrice = updatedNextCurSqrtPrice
					curTick = expectedSwapEndTick
				}

				// Validate active positions.
				validateActivePositions(poolId, positionMeta, tc.isPositionActiveFlag)
			})
		}
	})
}

// TestSwapOutGivenIn_SwapToAllowedBoundaries tests edge case behavior of swapping
// to min and max ticks.
func (s *KeeperTestSuite) TestSwapOutGivenIn_SwapToAllowedBoundaries() {
	const shouldStayWithinTheSameBucket = false

	var (
		tokenZeroDenom       = DefaultCoin0.Denom
		tokeOneDenom         = DefaultCoin1.Denom
		smallTokenOneCoinIn  = sdk.NewCoin(tokeOneDenom, osmomath.NewInt(100))
		smallTokenZeroCoinIn = sdk.NewCoin(tokenZeroDenom, osmomath.NewInt(100))
	)

	// tests tick crossing behavior around min tick
	// It first swaps to the min tick, then tries to swap below it and fails.
	// At the same time, it validates that the liquidity and current tick are updated correctly.
	// Additionally, it validates that it is still possible to swap right.
	s.Run("to min tick", func() {
		s.SetupTest()

		poolId, _ := s.setupPoolAndPositions(tickSpacingOne, defaultTickSpacingsAway, DefaultCoins)

		// Fetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		// Compute tokenIn amount necessary to reach the min tick.
		const isZeroForOne = true
		tokenIn, _, _ := s.computeSwapAmounts(poolId, osmomath.BigDec{}, types.MinInitializedTick, isZeroForOne, shouldStayWithinTheSameBucket)

		// Swap the computed large amount.
		s.swapZeroForOneLeft(poolId, sdk.NewCoin(tokenZeroDenom, tokenIn.Ceil().TruncateInt()))

		// Assert that min tick is crossed and liquidity is zero.
		s.assertPoolTickEquals(poolId, types.MinCurrentTick)
		s.assertPoolLiquidityEquals(poolId, osmomath.ZeroDec())

		// Assert that full range positions are now inactive.
		s.assertPositionOutOfRange(poolId, types.MinInitializedTick, types.MaxTick)

		// Refetch pool
		pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		// Validate cannot swap left again
		_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, smallTokenZeroCoinIn, tokeOneDenom, osmomath.ZeroInt(), osmomath.ZeroDec())
		s.Require().Error(err)
		s.Require().ErrorContains(err, types.InvalidAmountCalculatedError{Amount: osmomath.ZeroInt()}.Error())

		// Validate the ability to swap right
		s.swapOneForZeroRight(poolId, smallTokenOneCoinIn)
	})

	// tests tick crossing behavior around max tick
	// It first swaps to the max tick, then tries to swap beyond it and fails.
	// At the same time, it validates that the liquidity and current tick are updated correctly.
	// Additionally, it validates that it is still possible to swap left.
	s.Run("to max tick", func() {
		s.SetupTest()

		poolId, _ := s.setupPoolAndPositions(tickSpacingOne, defaultTickSpacingsAway, DefaultCoins)

		// Fetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		// Compute tokenIn amount necessary to reach the max tick.
		const isZeroForOne = false
		tokenIn, _, _ := s.computeSwapAmounts(poolId, osmomath.BigDec{}, types.MaxTick, isZeroForOne, shouldStayWithinTheSameBucket)

		// Swap the computed large amount.
		s.swapOneForZeroRight(poolId, sdk.NewCoin(tokeOneDenom, tokenIn.Ceil().TruncateInt()))

		// Assert that max tick is crossed and liquidity is zero.
		// N.B.: Since we can only have upper ticks of positions be initialized on the max tick,
		// the liquidity net amounts on MaxTick are always negative. Therefore, when swapping one
		// for zero and crossing a max tick to be "within in", we always end up with a current liquidity of zero.
		s.assertPoolTickEquals(poolId, types.MaxTick)
		s.assertPoolLiquidityEquals(poolId, osmomath.ZeroDec())

		// Assert that full range positions are now inactive.
		s.assertPositionOutOfRange(poolId, types.MinInitializedTick, types.MaxTick)

		// Refetch pool
		pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		// Validate cannot swap right again
		_, err = s.App.ConcentratedLiquidityKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], pool, smallTokenOneCoinIn, tokenZeroDenom, osmomath.ZeroInt(), osmomath.ZeroDec())
		s.Require().Error(err)
		s.Require().ErrorContains(err, types.InvalidAmountCalculatedError{Amount: osmomath.ZeroInt()}.Error())

		// Validate the ability to swap left
		s.swapZeroForOneLeft(poolId, smallTokenZeroCoinIn)
	})
}

// TestSwapOutGivenIn_GetLiquidityFromAmountsPositionBounds tests edge case behavior of swapping
// in either direction and how it relates to the calculation of liquidity from amounts
// and its bound checks.
// The test generates 4 positions where each is 1 tick wider than the previous one.
// Then, it swaps to lower edge of one position and runs some assertion on the way
// that liquidiy is calculated for that position as well as the adjacent position on the swap path.
// It then repeats this for the other direction.
func (s *KeeperTestSuite) TestSwapOutGivenIn_GetLiquidityFromAmountsPositionBounds() {
	// See definition of defaultTickSpacingsAway for position layout diagram.
	poolId, positions := s.setupPoolAndPositions(tickSpacingOne, defaultTickSpacingsAway, DefaultCoins)
	var (
		// 3 tick spacings away [30999997, 31000003) (3TS) from the original current tick (31000000)
		positionThreeTS               = positions[1]
		positionThreeTSLowerTick      = positionThreeTS.lowerTick
		positionThreeTSUpperTick      = positionThreeTS.upperTick
		positionThreeTSLowerSqrtPrice = s.tickToSqrtPrice(positionThreeTSLowerTick)
		positionThreeTSUpperSqrtPrice = s.tickToSqrtPrice(positionThreeTSUpperTick)

		// 2 tick spacings away [30999998, 31000002) (2TS) from the original current tick (31000000)
		positionTwoTS               = positions[2]
		positionTwoTSLowerTick      = positionTwoTS.lowerTick
		positionTwoTSUpperTick      = positionTwoTS.upperTick
		positionTwoTSLowerSqrtPrice = s.tickToSqrtPrice(positionTwoTSLowerTick)
		positionTwoTSUpperSqrtPrice = s.tickToSqrtPrice(positionTwoTSUpperTick)
	)

	// Assert that the liquidity computed from amounts utilized the "in-range" option.
	validateInRangeLiquidityFromAmounts := func(currentSqrtPrice, lowerTickSqrtPrice, upperTickSqrtPrice osmomath.BigDec) {
		liquidity0 := math.Liquidity0(DefaultAmt0, currentSqrtPrice, upperTickSqrtPrice)
		liquidity1 := math.Liquidity1(DefaultAmt1, currentSqrtPrice, lowerTickSqrtPrice)
		expectedLiquidity := osmomath.MinDec(liquidity0, liquidity1)
		actualLiquidity := math.GetLiquidityFromAmounts(currentSqrtPrice, lowerTickSqrtPrice, upperTickSqrtPrice, DefaultAmt0, DefaultAmt1)
		s.Require().Equal(expectedLiquidity, actualLiquidity)
	}

	// Assert one tick difference
	s.Require().Equal(positionThreeTSLowerTick, positionTwoTSLowerTick-tickSpacingOne)

	// Fetch pool
	p, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
	s.Require().NoError(err)

	tokenZero := p.GetToken0()
	tokenOne := p.GetToken1()

	// Persisting setup context to use
	// for both zero for one and one for zero tests.
	setupCtx := s.Ctx

	s.Run("zero for one", func() {
		// Set temporary cache context on the suite
		// so that we can reuse test helpers without mutating
		// the setup
		cacheCtx, _ := s.Ctx.CacheContext()
		s.Ctx = cacheCtx

		// Compute tokenIn amount necessary to reach the lower tick of the 3TS position.
		const (
			isZeroForOne                  = true
			shouldStayWithinTheSameBucket = false
		)
		tokenIn, _, _ := s.computeSwapAmounts(poolId, osmomath.BigDec{}, positionThreeTSLowerTick, isZeroForOne, shouldStayWithinTheSameBucket)

		// Swap until the lower tick of the 3TS position is reached.
		s.swapZeroForOneLeft(poolId, sdk.NewCoin(tokenZero, tokenIn.TruncateInt()))

		// Assert that the desired tick is reached.
		s.assertPoolTickEquals(poolId, positionThreeTSLowerTick)

		// Refetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		currentSqrtPrice := pool.GetCurrentSqrtPrice()

		// Current sqrt price is still above the lower tick of the 3TS position.
		s.Require().True(currentSqrtPrice.GT(positionThreeTSLowerSqrtPrice))

		// Current sqrt price is directly on the lower tick sqrt price of the 2TS position.
		s.Require().True(currentSqrtPrice.Equal(positionTwoTSLowerSqrtPrice))

		// 3TS position is in range.
		s.assertPositionInRange(poolId, positionThreeTSLowerTick, positionThreeTS.upperTick)

		// Liquidity from amounts for position 3TS is computed using the in-range option.
		validateInRangeLiquidityFromAmounts(currentSqrtPrice, positionThreeTSLowerSqrtPrice, positionThreeTSUpperSqrtPrice)

		// 2TS position is out of range.
		s.assertPositionOutOfRange(poolId, positionTwoTSLowerTick, positionTwoTS.upperTick)

		// 2TS position should consist of token zero only as it is to the right of the active range.
		liquidity02TS := math.Liquidity0(DefaultAmt0, currentSqrtPrice, positionTwoTSUpperSqrtPrice)
		actualLiquidity2Ts := math.GetLiquidityFromAmounts(pool.GetCurrentSqrtPrice(), positionTwoTSLowerSqrtPrice, s.tickToSqrtPrice(positionTwoTS.upperTick), DefaultAmt0, DefaultAmt1)
		s.Require().Equal(liquidity02TS, actualLiquidity2Ts)

		// Reset suite context
		s.Ctx = setupCtx
	})

	s.Run("one for zero", func() {
		// Set temporary cache context on the suite
		// so that we can reuse test helpers without mutating
		// the setup
		cacheCtx, _ := s.Ctx.CacheContext()
		s.Ctx = cacheCtx

		/// Compute tokenIn amount necessary to reach the upper tick of the 3TS position.
		const (
			isZeroForOne                  = false
			shouldStayWithinTheSameBucket = false
		)
		tokenIn, _, _ := s.computeSwapAmounts(poolId, osmomath.BigDec{}, positionTwoTSUpperTick, isZeroForOne, shouldStayWithinTheSameBucket)

		s.swapOneForZeroRight(poolId, sdk.NewCoin(tokenOne, tokenIn.TruncateInt()))

		// Assert that the desired tick is reached.
		s.assertPoolTickEquals(poolId, positionTwoTSUpperTick)

		// Refetch pool
		pool, err := s.App.ConcentratedLiquidityKeeper.GetPoolById(s.Ctx, poolId)
		s.Require().NoError(err)

		currentSqrtPrice := pool.GetCurrentSqrtPrice()

		// Current sqrt price equals to 2TS position's upper tick sqrt price.
		s.Require().True(currentSqrtPrice.Equal(positionTwoTSUpperSqrtPrice))

		// Current sqrt price is below the upper tick sqrt price of the 3TS position.
		s.Require().True(currentSqrtPrice.LT(positionThreeTSUpperSqrtPrice))

		// 3 TS position is in range.
		s.assertPositionInRange(poolId, positionThreeTSLowerTick, positionThreeTS.upperTick)

		// Liquidity from amounts for 3TS position is computed using the in-range option.
		validateInRangeLiquidityFromAmounts(currentSqrtPrice, positionThreeTSLowerSqrtPrice, positionThreeTSUpperSqrtPrice)

		// 2TS position is out of range.
		s.assertPositionOutOfRange(poolId, positionTwoTSLowerTick, positionTwoTS.upperTick)

		// 2TS should consist of token one only as it is to the right of the active range.
		liquidity1Pos3 := math.Liquidity1(DefaultAmt1, currentSqrtPrice, positionTwoTSLowerSqrtPrice)
		actualLiquidityPos3 := math.GetLiquidityFromAmounts(pool.GetCurrentSqrtPrice(), positionTwoTSLowerSqrtPrice, positionTwoTSUpperSqrtPrice, DefaultAmt0, DefaultAmt1)
		s.Require().Equal(liquidity1Pos3, actualLiquidityPos3)

		// Reset suite context
		s.Ctx = setupCtx
	})
}
