package swapstrategy

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// computeFeeChargePerSwapStepOutGivenIn returns the total fee charge per swap step given the parameters.
// Assumes swapping for token out given token in.
// - currentSqrtPrice the sqrt price at which the swap step begins.
// - nextTickSqrtPrice the next tick's sqrt price.
// - sqrtPriceLimit the sqrt price corresponding to the sqrt of the price representing price impact protection.
// - amountIn the amount of token in to be consumed during the swap step
// - amountSpecifiedRemaining is the total remaining amount of token in that needs to be consumed to complete the swap.
// - swapFee the swap fee to be charged.
//
// If swap fee is negative, it panics.
// If swap fee is 0, returns 0. Otherwise, computes and returns the fee charge per step.
// TODO: test this function.
func computeFeeChargePerSwapStepOutGivenIn(currentSqrtPrice, nextTickSqrtPrice, sqrtPriceTarget, amountIn, amountSpecifiedRemaining, swapFee sdk.Dec) sdk.Dec {
	feeChargeTotal := sdk.ZeroDec()

	if swapFee.IsNegative() {
		// This should never happen but is added as a defense-in-depth measure.
		panic(fmt.Errorf("swap fee must be non-negative, was (%s)", swapFee))
	}

	if swapFee.IsZero() {
		return feeChargeTotal
	}

	max := sqrtPriceTarget == nextTickSqrtPrice

	// In both cases, charge fee on the full amount that the tick
	// originally had.
	if max {
		feeChargeTotal = amountIn.Mul(swapFee)
	} else {
		// Otherwise, the current tick had enough liquidity to fulfill the swap
		// In that case, the fee is the difference between
		// the amount needed to fulfill and the actual amount we ended up charging.
		feeChargeTotal = amountSpecifiedRemaining.Sub(amountIn)
	}

	return feeChargeTotal
}
