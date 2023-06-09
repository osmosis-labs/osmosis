package swapstrategy

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

// zeroForOneStrategy implements the swapStrategy interface.
// This implementation assumes that we are swapping token 0 for
// token 1 and performs calculations accordingly.
//
// With this strategy, we are moving to the left of the current
// tick index and square root price.
type zeroForOneStrategy struct {
	sqrtPriceLimit sdk.Dec
	storeKey       sdk.StoreKey
	spreadFactor   sdk.Dec
	tickSpacing    uint64
}

var _ swapStrategy = (*zeroForOneStrategy)(nil)

// GetSqrtTargetPrice returns the target square root price given the next tick square root price.
// If the given nextTickSqrtPrice is less than the sqrt price limit, the sqrt price limit is returned.
// Otherwise, the input nextTickSqrtPrice is returned.
func (s zeroForOneStrategy) GetSqrtTargetPrice(nextTickSqrtPrice sdk.Dec) sdk.Dec {
	if nextTickSqrtPrice.LT(s.sqrtPriceLimit) {
		return s.sqrtPriceLimit
	}
	return nextTickSqrtPrice
}

// ComputeSwapStepOutGivenIn calculates the next sqrt price, the amount of token in consumed, the amount out to return to the user, and total spread reward charge on token in.
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
func (s zeroForOneStrategy) ComputeSwapStepOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountZeroInRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// Estimate the amount of token zero needed until the target sqrt price is reached.
	amountZeroIn := math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop

	// Calculate sqrtPriceNext on the amount of token remaining after spread reward.
	amountZeroInRemainingLessSpreadReward := amountZeroInRemaining.Mul(sdk.OneDec().Sub(s.spreadFactor))
	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after spread reward than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountZeroInRemainingLessSpreadReward.GTE(amountZeroIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after spread reward.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount0InRoundingUp(sqrtPriceCurrent, liquidity, amountZeroInRemainingLessSpreadReward)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after spread reward was needed
	// to complete the swap step. This implies that some of the amount remaining after spread reward is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountOneOut := math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	// Handle spread rewards.
	// Note that spread reward is always charged on the amount in.
	spreadRewardChargeTotal := computeSpreadRewardChargePerSwapStepOutGivenIn(hasReachedTarget, amountZeroIn, amountZeroInRemaining, s.spreadFactor)

	return sqrtPriceNext, amountZeroIn, amountOneOut, spreadRewardChargeTotal
}

// ComputeSwapStepInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total spread reward charge on token in.
// Parameters:
//   - sqrtPriceCurrent is the current sqrt price.
//   - sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
//     1. Next tick sqrt price.
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
func (s zeroForOneStrategy) ComputeSwapStepInGivenOut(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountOneRemainingOut sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// Estimate the amount of token one needed until the target sqrt price is reached.
	amountOneOut := math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)

	// Calculate sqrtPriceNext on the amount of token remaining. Note that the
	// spread reward is not charged as amountRemaining is amountOut, and we only charge spread reward on
	// amount in.
	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after spread reward than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountOneRemainingOut.GTE(amountOneOut) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after spread reward.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount1OutRoundingDown(sqrtPriceCurrent, liquidity, amountOneRemainingOut)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after spread reward was needed
	// to complete the swap step. This implies that some of the amount remaining after spread reward is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountOneOut = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountZeroIn := math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true)

	// Handle spread rewards.
	// Note that spread reward is always charged on the amount in.
	spreadRewardChargeTotal := computeSpreadRewardChargeFromAmountIn(amountZeroIn, s.spreadFactor)

	return sqrtPriceNext, amountOneOut, amountZeroIn, spreadRewardChargeTotal
}

// InitializeNextTickIterator returns iterator that seeks to the next tick from the given tickIndex.
// If nex tick relative to tickINdex does not exist in the store, it will return an invalid iterator.
//
// oneForZeroStrategy assumes moving to the left of the current square root price.
// As a result, we use a reverse iterator to seek to the next tick index relative to the currentTickIndexPlusOne.
// Since end key of the reverse iterator is exclusive, we search from current + 1 tick index.
// in decrasing lexicographic order until a tick one smaller than current is found.
// Returns an invalid iterator if currentTickIndexPlusOne is not in the store.
// Panics if fails to parse tick index from bytes.
// The caller is responsible for closing the iterator on success.
func (s zeroForOneStrategy) InitializeNextTickIterator(ctx sdk.Context, poolId uint64, currentTickIndexPlusOne int64) dbm.Iterator {
	store := ctx.KVStore(s.storeKey)
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)
	startKey := types.TickIndexToBytes(currentTickIndexPlusOne)

	iter := prefixStore.ReverseIterator(nil, startKey)

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			iter.Close()
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}
		if tick < currentTickIndexPlusOne {
			break
		}
	}
	return iter
}

// InitializeTickValue returns the initial tick value for computing swaps based
// on the actual current tick.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// As a result, we use reverse iterator in InitializeNextTickIterator to find the next
// tick to the left of current. The end cursor for reverse iteration is non-inclusive
// so must add one here to make sure that the current tick is included in the search.
func (s zeroForOneStrategy) InitializeTickValue(currentTick int64) int64 {
	return currentTick + 1
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
func (s zeroForOneStrategy) SetLiquidityDeltaSign(deltaLiquidity sdk.Dec) sdk.Dec {
	return deltaLiquidity.Neg()
}

// ValidateSqrtPrice validates the given square root price
// relative to the current square root price on one side of the bound
// and the min/max sqrt price on the other side.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// Therefore, the following invariant must hold:
// types.MinSqrtRatio <= sqrtPrice <= current square root price
func (s zeroForOneStrategy) ValidateSqrtPrice(sqrtPrice, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt price if we are swapping asset0 for asset1
	if sqrtPrice.GT(currentSqrtPrice) || sqrtPrice.LT(types.MinSqrtPrice) {
		return types.SqrtPriceValidationError{SqrtPriceLimit: sqrtPrice, LowerBound: types.MinSqrtPrice, UpperBound: currentSqrtPrice}
	}
	return nil
}
