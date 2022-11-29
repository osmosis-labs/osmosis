package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

// swapStrategy defines the interface for computing a swap.
// There are 2 implementations of this interface:
// - zeroForOneStrategy to provide implementations when swapping token 0 for token 1.
// - oneForZeroStrategy to provide implementations when swapping token 1 for token 0.
type swapStrategy interface {
	// ComputeSwapStep calculates the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity,
	// and amount available to swap
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	ComputeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext, amountIn, amountOut sdk.Dec)
	// NextInitializedTick returns the next initialized tick index based on the
	// provided tickindex. If no initialized tick exists, <0, false>
	// will be returned. The zeroForOne argument indicates if we need to find the next
	// initialized tick to the left or right of the current tick index, where true
	// indicates searching to the left.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64) (next int64, initialized bool)
	// SetLiquidityDeltaSign sets the liquidity delta sign for the given liquidity delta.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	SetLiquidityDeltaSign(liquidityDelta sdk.Dec) sdk.Dec
	// SetNextTick returns the next tick after performing a swap step.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	SetNextTick(nextTick int64) sdk.Int
	// ValidatePriceLimit validates the given square root price limit
	// given the square root price.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error
}

// New returns a swap strategy based on the provided zeroForOne parameter
// with sqrtPriceLimit for the maximum square root price until which to perform
// the swap and the stor key of the module that stores swap data.
func New(zeroForOne bool, sqrtPriceLimit sdk.Dec, storeKey sdk.StoreKey) swapStrategy {
	if zeroForOne {
		return &zeroForOneStrategy{sqrtPriceLimit: sqrtPriceLimit, storeKey: storeKey}
	}
	return &oneForZeroStrategy{sqrtPriceLimit: sqrtPriceLimit, storeKey: storeKey}
}
