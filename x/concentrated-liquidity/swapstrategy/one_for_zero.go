package swapstrategy

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// oneForZeroStrategy implements the swapStrategy interface.
// This implementation assumes that we are swapping token 1 for
// token 0 and performs calculations accordingly.
//
// With this strategy, we are moving to the right of the current
// tick index and square root price.
type oneForZeroStrategy struct {
	sqrtPriceLimit sdk.Dec
	storeKey       sdk.StoreKey
	swapFee        sdk.Dec
	tickSpacing    uint64
}

var _ swapStrategy = (*oneForZeroStrategy)(nil)

// GetSqrtTargetPrice returns the target square root price given the next tick square root price.
// If the given nextTickSqrtPrice is greater than the sqrt price limit, the sqrt price limit is returned.
// Otherwise, the input nextTickSqrtPrice is returned.
func (s oneForZeroStrategy) GetSqrtTargetPrice(nextTickSqrtPrice sdk.Dec) sdk.Dec {
	if nextTickSqrtPrice.GT(s.sqrtPriceLimit) {
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
//   - amountOneRemainingIn is the amount of token one in remaining to be swapped. This amount is fully consumed
//     if sqrt price target is not reached. In that case, the returned amountOne is the amount remaining given.
//     Otherwise, the returned amountOneIn will be smaller than amountOneRemainingIn given.
//
// Returns:
//   - sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
//   - amountOneIn is the amount of token in consumed. It equals amountRemainingIn if target is reached. Otherwise, it is less than amountOneRemainingIn.
//   - amountZeroOut the amount of token out computed. It is the amount of token out to return to the user.
//   - feeChargeTotal is the total fee charge. The fee is charged on the amount of token in.
//
// OneForZero details:
// - oneForZeroStrategy assumes moving to the right of the current square root price.
func (s oneForZeroStrategy) ComputeSwapStepOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountOneInRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// Estimate the amount of token one needed until the target sqrt price is reached.
	amountOneIn := math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop

	// Calculate sqrtPriceNext on the amount of token remaining after fee.
	amountOneInRemainingLessFee := amountOneInRemaining.Mul(sdk.OneDec().Sub(s.swapFee))

	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after fee than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountOneInRemainingLessFee.GTE(amountOneIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after fee.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount1InRoundingDown(sqrtPriceCurrent, liquidity, amountOneInRemainingLessFee)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after fee was needed
	// to complete the swap step. This implies that some of the amount remaining after fee is left over after the
	// current swap step.
	if !hasReachedTarget {
		amountOneIn = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true) // N.B.: if this is false, causes infinite loop
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountZeroOut := math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	// Handle fees.
	// Note that fee is always charged on the amount in.
	feeChargeTotal := computeFeeChargePerSwapStepOutGivenIn(hasReachedTarget, amountOneIn, amountOneInRemaining, s.swapFee)

	return sqrtPriceNext, amountOneIn, amountZeroOut, feeChargeTotal
}

// ComputeSwapStepInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total fee charge on token in.
// Parameters:
//   - sqrtPriceCurrent is the current sqrt price.
//   - sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
//     1. Next tick sqrt price.
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
//   - feeChargeTotal is the total fee charge. The fee is charged on the amount of token in.
//
// OneForZero details:
// - oneForZeroStrategy assumes moving to the right of the current square root price.
func (s oneForZeroStrategy) ComputeSwapStepInGivenOut(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountZeroRemainingOut sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	// Estimate the amount of token zero needed until the target sqrt price is reached.
	// N.B.: contrary to out given in, we do not round up because we do not want to exceed the initial amount out at the end.
	amountZeroOut := math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)

	// Calculate sqrtPriceNext on the amount of token remaining. Note that the
	// fee is not charged as amountRemaining is amountOut, and we only charge fee on
	// amount in.
	var sqrtPriceNext sdk.Dec
	// If have more of the amount remaining after fee than estimated until target,
	// bound the next sqrtPriceNext by the target sqrt price.
	if amountZeroRemainingOut.GTE(amountZeroOut) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		// Otherwise, compute the next sqrt price based on the amount remaining after fee.
		sqrtPriceNext = math.GetNextSqrtPriceFromAmount0OutRoundingUp(sqrtPriceCurrent, liquidity, amountZeroRemainingOut)
	}

	hasReachedTarget := sqrtPriceTarget == sqrtPriceNext

	// If the sqrt price target was not reached, recalculate how much of the amount remaining after fee was needed
	// to complete the swap step. This implies that some of the amount remaining after fee is left over after the
	// current swap step.
	if !hasReachedTarget {
		// N.B.: contrary to out given in, we do not round up because we do not want to exceed the initial amount out at the end.
		amountZeroOut = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	}

	// Calculate the amount of the other token given the sqrt price range.
	amountOneIn := math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, true)

	// Handle fees.
	// Note that fee is always charged on the amount in.
	feeChargeTotal := computeFeeChargeFromAmountIn(amountOneIn, s.swapFee)

	return sqrtPriceNext, amountZeroOut, amountOneIn, feeChargeTotal
}

// InitializeTickValue returns the initial tick value for computing swaps based
// on the actual current tick.
//
// oneForZeroStrategy assumes moving to the right of the current square root price.
// As a result, we use forward iterator in NextInitializedTick to find the next
// tick to the left of current. The end cursor for forward iteration is inclusive.
// Therefore, this method is, essentially a no-op. The logic is reversed for
// zeroForOneStrategy where we use reverse iterator and have to add one to
// the input. Therefore, we define this method to account for different strategies.
func (s oneForZeroStrategy) InitializeTickValue(currentTick sdk.Int) sdk.Int {
	return currentTick
}

// NextInitializedTick returns the next initialized tick index based on the
// provided tickindex. If no initialized tick exists, <0, false>
// will be returned.
//
// oneForZerostrategy searches for the next tick to the right of the current tickIndex.
func (s oneForZeroStrategy) NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64) (next sdk.Int, initialized bool) {
	store := ctx.KVStore(s.storeKey)

	// Construct a prefix store with a prefix of <TickPrefix | poolID>, allowing
	// us to retrieve the next initialized tick without having to scan all ticks.
	prefixBz := types.KeyTickPrefixByPoolId(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	startKey := types.TickIndexToBytes(tickIndex)

	iter := prefixStore.Iterator(startKey, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}

		if tick > tickIndex {
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
// oneForZeroStrategy assumes moving to the right of the current square root price.
// When we move to the right, we must be crossing lower ticks first where
// liqudiity delta tracks the amount of liquidity being added. So the sign must be
// positive.
func (s oneForZeroStrategy) SetLiquidityDeltaSign(deltaLiquidity sdk.Dec) sdk.Dec {
	return deltaLiquidity
}

// ValidateSqrtPrice validates the given square root price
// relative to the current square root price on one side of the bound
// and the min/max sqrt price on the other side.
//
// oneForZeroStrategy assumes moving to the right of the current square root price.
// Therefore, the following invariant must hold:
// current square root price <= sqrtPrice <= types.MaxSqrtRatio
func (s oneForZeroStrategy) ValidateSqrtPrice(sqrtPrice, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is above the current sqrt price but lower than the maximum sqrt price since we are swapping asset1 for asset0
	if sqrtPrice.LT(currentSqrtPrice) || sqrtPrice.GT(types.MaxSqrtPrice) {
		return types.SqrtPriceValidationError{SqrtPriceLimit: sqrtPrice, LowerBound: currentSqrtPrice, UpperBound: types.MaxSqrtPrice}
	}
	return nil
}
