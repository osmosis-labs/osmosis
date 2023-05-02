package swapstrategy

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
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
	swapFee        sdk.Dec
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

// ComputeSwapStepOutGivenIn calculates the next sqrt price, the amount of token in consumed, the amount out to return to the user, and total fee charge on token in.
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
//   - feeChargeTotal is the total fee charge. The fee is charged on the amount of token in.
//
// ZeroForOne details:
// - zeroForOneStrategy assumes moving to the left of the current square root price.
func (s zeroForOneStrategy) ComputeSwapStepOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountZeroInRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// Estimate the amount of token zero needed until the target sqrt price is reached.
	amountZeroIn := math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop

	// Calculate sqrtPriceNext on the amount of token remaining after fee.
	amountZeroInRemainingLessFee := amountZeroInRemaining.Mul(sdk.OneDec().Sub(s.swapFee))
	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after fee than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountZeroInRemainingLessFee.GTE(amountZeroIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after fee.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount0InRoundingUp(sqrtPriceCurrent, liquidity, amountZeroInRemainingLessFee)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after fee was needed
	// to complete the swap step. This implies that some of the amount remaining after fee is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountZeroIn = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountOneOut := math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	// Handle fees.
	// Note that fee is always charged on the amount in.
	feeChargeTotal := computeFeeChargePerSwapStepOutGivenIn(hasReachedTarget, amountZeroIn, amountZeroInRemaining, s.swapFee)

	return sqrtPriceNext, amountZeroIn, amountOneOut, feeChargeTotal
}

// ComputeSwapStepInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total fee charge on token in.
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
//   - feeChargeTotal is the total fee charge. The fee is charged on the amount of token in.
//
// ZeroForOne details:
// - zeroForOneStrategy assumes moving to the left of the current square root price.
func (s zeroForOneStrategy) ComputeSwapStepInGivenOut(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountOneRemainingOut sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// Estimate the amount of token one needed until the target sqrt price is reached.
	amountOneOut := math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)

	// Calculate sqrtPriceNext on the amount of token remaining. Note that the
	// fee is not charged as amountRemaining is amountOut, and we only charge fee on
	// amount in.
	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after fee than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountOneRemainingOut.GTE(amountOneOut) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after fee.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount1OutRoundingDown(sqrtPriceCurrent, liquidity, amountOneRemainingOut)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after fee was needed
	// to complete the swap step. This implies that some of the amount remaining after fee is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountOneOut = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountZeroIn := math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true)

	// Handle fees.
	// Note that fee is always charged on the amount in.
	feeChargeTotal := computeFeeChargeFromAmountIn(amountZeroIn, s.swapFee)

	return sqrtPriceNext, amountOneOut, amountZeroIn, feeChargeTotal
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
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
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
