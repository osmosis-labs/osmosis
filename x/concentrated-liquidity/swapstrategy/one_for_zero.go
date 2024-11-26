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

// oneForZeroStrategy implements the swapStrategy interface.
// This implementation assumes that we are swapping token 1 for
// token 0 and performs calculations accordingly.
//
// With this strategy, we are moving to the right of the current
// tick index and square root price.
type oneForZeroStrategy struct {
	sqrtPriceLimit osmomath.BigDec
	storeKey       storetypes.StoreKey
	spreadFactor   osmomath.Dec

	// oneMinusSpreadFactor is 1 - spreadFactor
	oneMinusSpreadFactor osmomath.Dec
	// spfOverOneMinusSpf is spreadFactor / (1 - spreadFactor)
	spfOverOneMinusSpf osmomath.Dec
}

var _ SwapStrategy = (*oneForZeroStrategy)(nil)

func (s oneForZeroStrategy) ZeroForOne() bool { return false }

// GetSqrtTargetPrice returns the target square root price given the next tick square root price.
// If the given nextTickSqrtPrice is greater than the sqrt price limit, the sqrt price limit is returned.
// Otherwise, the input nextTickSqrtPrice is returned.
func (s oneForZeroStrategy) GetSqrtTargetPrice(nextTickSqrtPrice osmomath.BigDec) osmomath.BigDec {
	if nextTickSqrtPrice.GT(s.sqrtPriceLimit) {
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
//   - amountOneRemainingIn is the amount of token one in remaining to be swapped. This amount is fully consumed
//     if sqrt price target is not reached. In that case, the returned amountOne is the amount remaining given.
//     Otherwise, the returned amountOneIn will be smaller than amountOneRemainingIn given.
//
// Returns:
//   - sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
//   - amountOneIn is the amount of token in consumed. It equals amountRemainingIn if target is reached. Otherwise, it is less than amountOneRemainingIn.
//   - amountZeroOut the amount of token out computed. It is the amount of token out to return to the user.
//   - spreadRewardChargeTotal is the total spread reward charge. The spread reward is charged on the amount of token in.
//
// OneForZero details:
// - oneForZeroStrategy assumes moving to the right of the current square root price.
func (s oneForZeroStrategy) ComputeSwapWithinBucketOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget osmomath.BigDec, liquidity, amountOneInRemaining osmomath.Dec) (osmomath.BigDec, osmomath.Dec, osmomath.Dec, osmomath.Dec) {
	// Estimate the amount of token one needed until the target sqrt price is reached.
	amountOneIn := math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, true)

	// Calculate sqrtPriceNext on the amount of token remaining after spread reward.
	oneMinusTakerFee := s.getOneMinusSpreadFactor()
	amountOneInRemainingLessSpreadReward := osmomath.NewBigDecFromDecMulDec(amountOneInRemaining, oneMinusTakerFee)

	var sqrtPriceNext osmomath.BigDec
	// If have more of the amount remaining after spread reward than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountOneInRemainingLessSpreadReward.GTE(amountOneIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after spread reward.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount1InRoundingDown(sqrtPriceCurrent, liquidity, amountOneInRemainingLessSpreadReward)
	}

	hasReachedTarget := sqrtPriceTarget.Equal(sqrtPriceNext)

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after spread reward was needed
	// to complete the swap step. This implies that some of the amount remaining after spread reward is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountOneIn = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountZeroOut := math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	// Round up to charge user more in pool's favor.
	amountInDecFinal := amountOneIn.DecRoundUp()

	// Handle spread rewards.
	// Note that spread reward is always charged on the amount in.
	spreadRewardChargeTotal := computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountInDecFinal, amountOneInRemaining, s.spreadFactor, s.getSpfOverOneMinusSpf)

	// Round down amount out to give user less in pool's favor.
	return sqrtPriceNext, amountInDecFinal, amountZeroOut.Dec(), spreadRewardChargeTotal
}

// ComputeSwapWithinBucketInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total spread reward charge on token in.
// This assumes swapping over a single bucket where the liqudiity stays constant until we cross the next initialized tick of the next bucket.
// Parameters:
//   - sqrtPriceCurrent is the current sqrt price.
//   - sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
//     1. Next initialized tick sqrt price.
//     2. Sqrt price limit representing price impact protection.
//   - liquidity is the amount of liquidity between the sqrt price current and sqrt price target.
//   - amountZeroRemainingOut is the amount of token zero out remaining to be swapped to estimate how much of token one in is needed to be charged.
//     This amount is fully consumed if sqrt price target is not reached. In that case, the returned amountOut is the amount zero remaining given.
//     Otherwise, the returned amountOut will be smaller than amountZeroRemainingOut given.
//
// Returns:
//   - sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
//   - amountZeroOut is the amount of token zero out consumed. It equals amountZeroRemainingOut if target is reached. Otherwise, it is less than amountZeroRemainingOut.
//   - amountIn is the amount of token in computed. It is the amount of token one in to charge to the user for the desired amount out.
//   - spreadRewardChargeTotal is the total spread reward charge. The spread reward is charged on the amount of token in.
//
// OneForZero details:
// - oneForZeroStrategy assumes moving to the right of the current square root price.
func (s oneForZeroStrategy) ComputeSwapWithinBucketInGivenOut(sqrtPriceCurrent, sqrtPriceTarget osmomath.BigDec, liquidity, amountZeroRemainingOut osmomath.Dec) (osmomath.BigDec, osmomath.Dec, osmomath.Dec, osmomath.Dec) {
	liquidityBigDec := osmomath.BigDecFromDec(liquidity)
	amountZeroRemainingOutBigDec := osmomath.BigDecFromDec(amountZeroRemainingOut)

	// Estimate the amount of token zero needed until the target sqrt price is reached.
	// N.B.: contrary to out given in, we do not round up because we do not want to exceed the initial amount out at the end.
	amountZeroOut := math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)

	// Calculate sqrtPriceNext on the amount of token remaining. Note that the
	// spread reward is not charged as amountRemaining is amountOut, and we only charge spread reward on
	// amount in.
	var sqrtPriceNext osmomath.BigDec
	// If have more of the amount remaining after spread reward than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountZeroRemainingOutBigDec.GTE(amountZeroOut) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after spread reward.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount0OutRoundingUp(sqrtPriceCurrent, liquidityBigDec, amountZeroRemainingOut)
	}

	hasReachedTarget := sqrtPriceTarget.Equal(sqrtPriceNext)

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after spread reward was needed
	// to complete the swap step. This implies that some of the amount remaining after spread reward is left over after the
	// current swap step.
	if !hasReachedTarget {
		// N.B.: contrary to out given in, we do not round up because we do not want to exceed the initial amount out at the end.
		amountZeroOut = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountOneIn := math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true)

	// Round up to charge user more in pool's favor.
	amountOneInFinal := amountOneIn.DecRoundUp()

	// Handle spread rewards.
	// Note that spread reward is always charged on the amount in.
	spreadRewardChargeTotal := computeSpreadRewardChargeFromAmountIn(amountOneInFinal, s.getSpfOverOneMinusSpf())

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
	if amountZeroOut.GT(amountZeroRemainingOutBigDec) {
		amountZeroOut = amountZeroRemainingOutBigDec
	}

	// Round down amount out to give user less in pool's favor.
	return sqrtPriceNext, amountZeroOut.Dec(), amountOneInFinal, spreadRewardChargeTotal
}

func (s oneForZeroStrategy) getOneMinusSpreadFactor() osmomath.Dec {
	if s.oneMinusSpreadFactor.IsNil() {
		s.oneMinusSpreadFactor = oneDec.Sub(s.spreadFactor)
	}
	return s.oneMinusSpreadFactor
}

func (s oneForZeroStrategy) getSpfOverOneMinusSpf() osmomath.Dec {
	if s.spfOverOneMinusSpf.IsNil() {
		s.spfOverOneMinusSpf = s.spreadFactor.QuoRoundUp(s.getOneMinusSpreadFactor())
	}
	return s.spfOverOneMinusSpf
}

// InitializeNextTickIterator returns iterator that seeks to the next tick from the given tickIndex.
// In one for zero direction, the search is EXCLUSIVE of the current tick index.
// If next tick relative to currentTickIndex is not initialized (does not exist in the store),
// it will return an invalid iterator.
// This is a requirement to satisfy our "active range" invariant of "lower tick <= current tick < upper tick".
// If we swap twice and the first swap crosses tick X, we do not want the second swap to cross tick X again
// so we search from X + 1.
//
// oneForZeroStrategy assumes moving to the right of the current square root price.
// As a result, we use forward iterator to seek to the next tick index relative to the currentTickIndex.
// Since start key of the forward iterator is inclusive, we search directly from the currentTickIndex
// forwards in increasing lexicographic order until a tick greater than currentTickIndex is found.
// Returns an invalid iterator if no tick greater than currentTickIndex is found in the store.
// Panics if fails to parse tick index from bytes.
// The caller is responsible for closing the iterator on success.
func (s oneForZeroStrategy) InitializeNextTickIterator(ctx sdk.Context, poolId uint64, currentTickIndex int64) dbm.Iterator {
	store := ctx.KVStore(s.storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)
	startKey := types.TickIndexToBytes(currentTickIndex)
	iter := prefixStore.Iterator(startKey, nil)

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			iter.Close()
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}

		if tick > currentTickIndex {
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
// oneForZeroStrategy assumes moving to the right of the current square root price.
// When we move to the right, we must be crossing lower ticks first where
// liqudiity delta tracks the amount of liquidity being added. So the sign must be
// positive.
func (s oneForZeroStrategy) SetLiquidityDeltaSign(deltaLiquidity osmomath.Dec) osmomath.Dec {
	return deltaLiquidity
}

// UpdateTickAfterCrossing updates the next tick after crossing
// to satisfy our "position in-range" invariant which is:
// lower tick <= current tick < upper tick.
// When crossing a tick in one for zero direction, we move
// right on the range. As a result, we end up crossing the upper tick
// that is exclusive. Therefore, we leave the next tick as is since
// it is already excluded from the current range.
func (s oneForZeroStrategy) UpdateTickAfterCrossing(nextTick int64) int64 {
	return nextTick
}

// ValidateSqrtPrice validates the given square root price
// relative to the current square root price on one side of the bound
// and the min/max sqrt price on the other side.
//
// oneForZeroStrategy assumes moving to the right of the current square root price.
// Therefore, the following invariant must hold:
// current square root price <= sqrtPrice <= types.MaxSqrtRatio
func (s oneForZeroStrategy) ValidateSqrtPrice(sqrtPrice osmomath.BigDec, currentSqrtPrice osmomath.BigDec) error {
	// check that the price limit is above the current sqrt price but lower than the maximum sqrt price since we are swapping asset1 for asset0
	if sqrtPrice.LT(currentSqrtPrice) || sqrtPrice.GT(types.MaxSqrtPriceBigDec) {
		return types.SqrtPriceValidationError{SqrtPriceLimit: sqrtPrice, LowerBound: currentSqrtPrice, UpperBound: types.MaxSqrtPriceBigDec}
	}
	return nil
}
