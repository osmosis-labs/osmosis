package swapstrategy

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// zeroForOneStrategy implements the swapStrategy interface.
// This implementation assumes that we are swapping token 0 for
// token 1 and performs calculations accordingly.
//
// With this strategy, we are moving to the left of the current
// tick index and square root price.
type zeroForOneStrategy struct {
	isOutGivenIn   bool
	sqrtPriceLimit sdk.Dec
	storeKey       sdk.StoreKey
	swapFee        sdk.Dec
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

// ComputeSwapStep calculates the next sqrt price, the new amount remaining, the amount of the token other than remaining given current price, and total fee charge.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// amountRemaining is the amount of token in when swapping out given in and token out when swapping in given out.
// amountRemaining is token zero.
// amountZero is token out when swapping in given out and token in when swapping out given in.
// amountOne is token in when swapping in given out and token out when swapping out given in.
// TODO: improve tests
func (s zeroForOneStrategy) ComputeSwapStep(sqrtPriceCurrent, sqrtPriceNextTick, liquidity, amountRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// sqrtPriceTarget is the maximum of sqrtPriceNextTick or sqrtPriceLimit.
	sqrtPriceTarget := s.GetSqrtTargetPrice(sqrtPriceNextTick)

	// Estimate the amount of token zero needed until the target sqrt price is reached.
	amountZero := math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop

	// Calculate sqrtPriceNext on the amount of token remaining after fee.
	amountRemainingLessFee := getAmountRemainingLessFee(amountRemaining, s.swapFee, s.isOutGivenIn)
	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after fee than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountRemainingLessFee.GTE(amountZero) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after fee.
		// TODO: when swapping in given out, GetNextSqrtPriceFromAmount0OutRoundingUp must be used.
		// To be addressed in: https://github.com/osmosis-labs/osmosis/issues/4427
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount0InRoundingUp(sqrtPriceCurrent, liquidity, amountRemainingLessFee)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after fee was needed
	// to complete the swap step. This implies that some of the amount remaining after fee is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountZero = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountOne := math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	// Handle fees.
	// Note that fee is always charged on the amount in.
	var feeChargeTotal sdk.Dec
	if s.isOutGivenIn {
		// amountZero is amount in.
		feeChargeTotal = computeFeeChargePerSwapStepOutGivenIn(sqrtPriceCurrent, hasReachedTarget, amountZero, amountRemaining, s.swapFee)
	} else {
		// amountOne is amount in.
		// TODO: multiplication with rounding up at precision end.
		feeChargeTotal = amountOne.Mul(s.swapFee).Quo(sdk.OneDec().Sub(s.swapFee))
	}

	return sqrtPriceNext, amountZero, amountOne, feeChargeTotal
}

// InitializeTickValue returns the initial tick value for computing swaps based
// on the actual current tick.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// As a result, we use reverse iterator in NextInitializedTick to find the next
// tick to the left of current. The end cursor for reverse iteration is non-inclusive
// so must add one here to make sure that the current tick is included in the search.
func (s zeroForOneStrategy) InitializeTickValue(currentTick sdk.Int) sdk.Int {
	return currentTick.Add(sdk.OneInt())
}

// NextInitializedTick returns the next initialized tick index based on the
// provided tickindex. If no initialized tick exists, <0, false>
// will be returned.
//
// zeroForOneStrategy searches for the next tick to the left of the current tickIndex.
func (s zeroForOneStrategy) NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64) (next sdk.Int, initialized bool) {
	store := ctx.KVStore(s.storeKey)

	// Construct a prefix store with a prefix of <TickPrefix | poolID>, allowing
	// us to retrieve the next initialized tick without having to scan all ticks.
	prefixBz := types.KeyTickPrefix(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	startKey := types.TickIndexToBytes(tickIndex)

	iter := prefixStore.ReverseIterator(nil, startKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}
		if tick <= tickIndex {
			return sdk.NewInt(tick), true
		}
	}
	return sdk.ZeroInt(), false
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

// ValidatePriceLimit validates the given square root price limit
// given the square root price.
//
// zeroForOneStrategy assumes moving to the left of the current square root price.
// Therefore, the following invariant must hold:
// types.MinSqrtRatio <= sqrtPriceLimit <= current square root price
func (s zeroForOneStrategy) ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt ratio if we are swapping asset0 for asset1
	if sqrtPriceLimit.GT(currentSqrtPrice) || sqrtPriceLimit.LT(types.MinSqrtRatio) {
		return types.InvalidPriceLimitError{SqrtPriceLimit: sqrtPriceLimit, LowerBound: types.MinSqrtRatio, UpperBound: sqrtPriceLimit}
	}
	return nil
}
