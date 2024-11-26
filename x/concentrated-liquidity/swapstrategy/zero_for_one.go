package swapstrategy

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"

	storetypes "cosmossdk.io/store/types"
)

// zeroForOneStrategy implements the swapStrategy interface.
// This implementation assumes that we are swapping token 0 for
// token 1 and performs calculations accordingly.
//
// With this strategy, we are moving to the left of the current
// tick index and square root price.
type zeroForOneStrategy struct {
	sqrtPriceLimit osmomath.BigDec
	storeKey       storetypes.StoreKey
	spreadFactor   osmomath.Dec

	// oneMinusSpreadFactor is 1 - spreadFactor
	oneMinusSpreadFactor osmomath.Dec
	// spfOverOneMinusSpf is spreadFactor / (1 - spreadFactor)
	spfOverOneMinusSpf osmomath.Dec
}

var _ SwapStrategy = (*zeroForOneStrategy)(nil)

func (s zeroForOneStrategy) ZeroForOne() bool { return true }

// GetSqrtTargetPrice returns the target square root price given the next tick square root price.
// If the given nextTickSqrtPrice is less than the sqrt price limit, the sqrt price limit is returned.
// Otherwise, the input nextTickSqrtPrice is returned.
func (s zeroForOneStrategy) GetSqrtTargetPrice(nextTickSqrtPrice osmomath.BigDec) osmomath.BigDec {
	if nextTickSqrtPrice.LT(s.sqrtPriceLimit) {
		return s.sqrtPriceLimit
	}
	return nextTickSqrtPrice
}

// ComputeSwapWithinBucketOutGivenIn calculates the next sqrt price, the amount of token in consumed, the amount out to return to the user, and total spread reward charge on token in.
// Parameters:
//   - sqrtPriceCurrent is the current sqrt price.
//   - sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
//     1. Next tick sqrt price.
//     2. Sqrt price limit representing price impact protection.
//   - liquidity is the amount of liquidity between the sqrt price current and sqrt price target.
//   - amountZeroInRemaining is the amount of token zero in remaining to be swapped. This amount is fully consumed
//     if sqrt price target is not reached. In that case, the returned amountZeroIn is the amount remaining given.
//     Otherwise, the returned amountIn will be smaller than amountZeroInRemaining given.
//
// Returns:
//   - sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
//   - amountZeroIn is the amount of token zero in consumed. It equals amountZeroInRemaining if target is reached. Otherwise, it is less than amountZeroInRemaining.
//   - amountOutComputed is the amount of token out computed. It is the amount of token out to return to the user.
//   - spreadRewardChargeTotal is the total spread reward charge. The spread reward is charged on the amount of token in.
//
// ZeroForOne details:
// - zeroForOneStrategy assumes moving to the left of the current square root price.
func (s zeroForOneStrategy) ComputeSwapWithinBucketOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget osmomath.BigDec, liquidity, amountZeroInRemaining osmomath.Dec) (osmomath.BigDec, osmomath.Dec, osmomath.Dec, osmomath.Dec) {
	liquidityBigDec := osmomath.BigDecFromDec(liquidity)

	// Estimate the amount of token zero needed until the target sqrt price is reached.
	amountZeroIn := math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop

	// Calculate sqrtPriceNext on the amount of token remaining after spread reward.
	oneMinusTakerFee := s.getOneMinusSpreadFactor()
	amountZeroInRemainingLessSpreadReward := osmomath.NewBigDecFromDecMulDec(amountZeroInRemaining, oneMinusTakerFee)

	var sqrtPriceNext osmomath.BigDec
	// If have more of the amount remaining after spread reward than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountZeroInRemainingLessSpreadReward.GTE(amountZeroIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after spread reward.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount0InRoundingUp(sqrtPriceCurrent, liquidityBigDec, amountZeroInRemainingLessSpreadReward)
	}

	hasReachedTarget := sqrtPriceTarget.Equal(sqrtPriceNext)

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after spread reward was needed
	// to complete the swap step. This implies that some of the amount remaining after spread reward is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountOneOut := math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	// Round up to charge user more in pool's favor.
	amountZeroInFinal := amountZeroIn.DecRoundUp()

	// Handle spread rewards.
	// Note that spread reward is always charged on the amount in.
	spreadRewardChargeTotal := computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountZeroInFinal, amountZeroInRemaining, s.spreadFactor, s.getSpfOverOneMinusSpf)

	// Round down amount out to give user less in pool's favor.
	return sqrtPriceNext, amountZeroInFinal, amountOneOut.Dec(), spreadRewardChargeTotal
}

// ComputeSwapWithinBucketInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total spread reward charge on token in.
// This assumes swapping over a single bucket where the liqudiity stays constant until we cross the next initialized tick of the next bucket.
// Parameters:
//   - sqrtPriceCurrent is the current sqrt price.
//   - sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
//     1. Next initialized tick sqrt price.
//     2. Sqrt price limit representing price impact protection.
//   - liquidity is the amount of liquidity between the sqrt price current and sqrt price target.
//   - amountOneRemainingOut is the amount of token one out remaining to be swapped to estimate how much of token zero in is needed to be charged.
//     This amount is fully consumed if sqrt price target is not reached. In that case, the returned amountOneOut is the amount remaining given.
//     Otherwise, the returned amountOneOut will be smaller than amountOneRemainingOut given.
//
// Returns:
//   - sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
//   - amountOneOut is the amount of token one out consumed. It equals amountOneRemainingOut if target is reached. Otherwise, it is less than amountOneRemainingOut.
//   - amountZeroIn is the amount of token zero in computed. It is the amount of token in to charge to the user for the desired amount out.
//   - spreadRewardChargeTotal is the total spread reward charge. The spread reward is charged on the amount of token in.
//
// ZeroForOne details:
// - zeroForOneStrategy assumes moving to the left of the current square root price.
func (s zeroForOneStrategy) ComputeSwapWithinBucketInGivenOut(sqrtPriceCurrent, sqrtPriceTarget osmomath.BigDec, liquidity, amountOneRemainingOut osmomath.Dec) (osmomath.BigDec, osmomath.Dec, osmomath.Dec, osmomath.Dec) {
	amountOneRemainingOutBigDec := osmomath.BigDecFromDec(amountOneRemainingOut)

	// Estimate the amount of token one needed until the target sqrt price is reached.
	amountOneOut := math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)

	// Calculate sqrtPriceNext on the amount of token remaining. Note that the
	// spread reward is not charged as amountRemaining is amountOut, and we only charge spread reward on
	// amount in.
	var sqrtPriceNext osmomath.BigDec
	// If have more of the amount remaining after spread reward than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountOneRemainingOutBigDec.GTE(amountOneOut) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after spread reward.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount1OutRoundingDown(sqrtPriceCurrent, liquidity, amountOneRemainingOutBigDec)
	}

	hasReachedTarget := sqrtPriceTarget.Equal(sqrtPriceNext)

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after spread reward was needed
	// to complete the swap step. This implies that some of the amount remaining after spread reward is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountOneOut = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountZeroIn := math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true)

	// Round up to charge user more in pool's favor.
	amountZeroInFinal := amountZeroIn.DecRoundUp()

	// Handle spread rewards.
	// Note that spread reward is always charged on the amount in.
	spreadRewardChargeTotal := computeSpreadRewardChargeFromAmountIn(amountZeroInFinal, s.getSpfOverOneMinusSpf())

	// Cap the output amount to not exceed the remaining output amount.
	// The reason why we must do this for in given out and NOT out given in is the following:
	// When swapping for exact out while not reaching sqrtPriceTarget, we calculate  sqrtPriceNext from the
	// amountRemainingOut. While calculating it, we round sqrtPriceNext in the direction opposite from the sqrtPriceCurrent.
	// This is because we need to move the price up enough so that we get the desired output amount out.
	// From newly calculate sqrtPriceNext, we then re-calculate the amountOut actually consumed. In certain cases, this
	// recalculation might lead to a slightly greater amount than remaining due to sqrtPriceNext having been rounded in
	// the opposite direction of the sqrtPriceCurrent. Therefore, we force the amountOut consumed to equal to amountRemaining.
	// This is acceptable since the former is calculated from the latter, and the only possible source of difference is rounding.
	// Going back to the exact in swap, when calculating the sqrtPriceNext, we round it in the direction of the sqrtPriceCurrent.
	// As a result, this rounding error should not be possible in its case.
	if amountOneOut.GT(amountOneRemainingOutBigDec) {
		amountOneOut = amountOneRemainingOutBigDec
	}

	// Round down amount out to give user less in pool's favor.
	return sqrtPriceNext, amountOneOut.Dec(), amountZeroInFinal, spreadRewardChargeTotal
}

func (s zeroForOneStrategy) getOneMinusSpreadFactor() osmomath.Dec {
	if s.oneMinusSpreadFactor.IsNil() {
		s.oneMinusSpreadFactor = oneDec.Sub(s.spreadFactor)
	}
	return s.oneMinusSpreadFactor
}

func (s zeroForOneStrategy) getSpfOverOneMinusSpf() osmomath.Dec {
	if s.spfOverOneMinusSpf.IsNil() {
		s.spfOverOneMinusSpf = s.spreadFactor.QuoRoundUp(s.getOneMinusSpreadFactor())
	}
	return s.spfOverOneMinusSpf
}

// InitializeNextTickIterator returns iterator that searches for the next tick given currentTickIndex.
// In zero for one direction, the search is INCLUSIVE of the current tick index.
// If next tick relative to currentTickIndex is not initialized (does not exist in the store),
// it will return an invalid iterator.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// As a result, we use a reverse iterator to seek to the next tick index relative to the currentTickIndex.
// Since end key of the reverse iterator is exclusive, we search from currentTickIndex + 1 tick index
// in decreasing lexicographic order until a tick one smaller than current is found.
// We add 1 so that the current tick index is included in the search. This is a requirement to satisfy our
// "active range" invariant of "lower tick <= current tick < upper tick". If we swapr right (zfo) and
// cross tick X, then immediately start swapping left (zfo), we should be able to cross tick X in the other direction.
// Returns an invalid iterator if no ticks smaller than currentTickIndex are initialized in the the store.
// Panics if fails to parse tick index from bytes.
// The caller is responsible for closing the iterator on success.
func (s zeroForOneStrategy) InitializeNextTickIterator(ctx sdk.Context, poolId uint64, currentTickIndex int64) dbm.Iterator {
	store := ctx.KVStore(s.storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)
	startKey := types.TickIndexToBytes(currentTickIndex + 1)

	iter := prefixStore.ReverseIterator(nil, startKey)

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			iter.Close()
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}
		if tick <= currentTickIndex {
			break
		}
	}
	return iter
}

// SetLiquidityDeltaSign sets the liquidity delta sign for the given liquidity delta.
// This is called when consuming all liquidity.
// When a position is created, we add liquidity to lower tick
// and subtract from the upper tick to reflect that this new
// liquidity would be added when the price crosses the lower tick
// going up, and subtracted when the price crosses the upper tick
// going up. As a result, the sign depend on the direction we are moving.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// When we move to the left, we must be crossing upper ticks first where
// liquidity delta tracks the amount of liquidity being removed. So the sign must be
// negative.
func (s zeroForOneStrategy) SetLiquidityDeltaSign(deltaLiquidity osmomath.Dec) osmomath.Dec {
	return deltaLiquidity.Neg()
}

// UpdateTickAfterCrossing updates the next tick after crossing
// to satisfy our "position in-range" invariant which is:
// lower tick <= current tick < upper tick
// When crossing a tick in zero for one direction, we move
// left on the range. As a result, we end up crossing the lower tick
// that is inclusive. Therefore, we must decrease the next tick
// by 1 additional unit so that it falls under the current range.
func (s zeroForOneStrategy) UpdateTickAfterCrossing(nextTick int64) int64 {
	return nextTick - 1
}

// ValidateSqrtPrice validates the given square root price
// relative to the current square root price on one side of the bound
// and the min/max sqrt price on the other side.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// Therefore, the following invariant must hold:
// types.MinSqrtRatio <= sqrtPrice <= current square root price
func (s zeroForOneStrategy) ValidateSqrtPrice(sqrtPrice osmomath.BigDec, currentSqrtPrice osmomath.BigDec) error {
	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt price if we are swapping asset0 for asset1
	if sqrtPrice.GT(currentSqrtPrice) || sqrtPrice.LT(types.MinSqrtPriceBigDec) {
		return types.SqrtPriceValidationError{SqrtPriceLimit: sqrtPrice, LowerBound: types.MinSqrtPriceBigDec, UpperBound: currentSqrtPrice}
	}
	return nil
}
