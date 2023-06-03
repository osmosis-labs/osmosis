package swapstrategy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

// swapStrategy defines the interface for computing a swap.
// There are 2 implementations of this interface:
// - zeroForOneStrategy to provide implementations when swapping token 0 for token 1.
// - oneForZeroStrategy to provide implementations when swapping token 1 for token 0.
type swapStrategy interface {
	// GetSqrtTargetPrice returns the target square root price given the next tick square root price
	// upon comparing it to sqrt price limit.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	GetSqrtTargetPrice(nextTickSqrtPrice sdk.Dec) sdk.Dec
	// ComputeSwapStepOutGivenIn calculates the next sqrt price, the amount of token in consumed, the amount out to return to the user, and total spread reward charge on token in.
	// Parameters:
	//   * sqrtPriceCurrent is the current sqrt price.
	//   * sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
	//       - Next tick sqrt price.
	//       - Sqrt price limit representing price impact protection.
	//   * liquidity is the amount of liquidity between the sqrt price current and sqrt price target.
	//   * amountRemainingIn is the amount of token in remaining to be swapped. This amount is fully consumed
	//   if sqrt price target is not reached. In that case, the returned amountInConsumed is the amount remaining given.
	//   Otherwise, the returned amountInConsumed will be smaller than amountRemainingIn given.
	// Returns:
	//   * sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
	//   * amountInConsumed is the amount of token in consumed. It equals amountRemainingIn if target is reached. Otherwise, it is less than amountRemainingIn.
	//   * amountOutComputed is the amount of token out computed. It is the amount of token out to return to the user.
	//   * spreadRewardChargeTotal is the total spread reward charge. The spread reward is charged on the amount of token in.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	ComputeSwapStepOutGivenIn(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemainingIn sdk.Dec) (sqrtPriceNext, amountInConsumed, amountOutComputed, spreadRewardChargeTotal sdk.Dec)
	// ComputeSwapStepInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total spread reward charge on token in.
	// Parameters:
	//   * sqrtPriceCurrent is the current sqrt price.
	//   * sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
	//       - Next tick sqrt price.
	//       - Sqrt price limit representing price impact protection.
	//   * liquidity is the amount of liquidity between the sqrt price current and sqrt price target.
	//   * amountRemainingOut is the amount of token out remaining to be swapped to estimate how much of token in is needed to be charged.
	//   This amount is fully consumed if sqrt price target is not reached. In that case, the returned amountOutConsumed is the amount remaining given.
	//   Otherwise, the returned amountOutConsumed will be smaller than amountRemainingOut given.
	// Returns:
	//   * sqrtPriceNext is the next sqrt price. It equals sqrt price target if target is reached. Otherwise, it is in-between sqrt price current and target.
	//   * amountOutConsumed is the amount of token out consumed. It equals amountRemainingOut if target is reached. Otherwise, it is less than amountRemainingOut.
	//   * amountInComputed is the amount of token in computed. It is the amount of token in to charge to the user for the desired amount out.
	//   * spreadRewardChargeTotal is the total spread reward charge. The spread reward is charged on the amount of token in.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	ComputeSwapStepInGivenOut(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemainingOut sdk.Dec) (sqrtPriceNext, amountOutConsumed, amountInComputed, spreadRewardChargeTotal sdk.Dec)
	// InitializeNextTickIterator returns iterator that seeks to the next tick from the given tickIndex.
	// If nex tick relative to tickINdex does not exist in the store, it will return an invalid iterator.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	InitializeNextTickIterator(ctx sdk.Context, poolId uint64, tickIndex int64) dbm.Iterator
	// InitializeTickValue returns the initial tick value for computing swaps based
	// on the actual current tick.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	InitializeTickValue(currentTick int64) int64
	// SetLiquidityDeltaSign sets the liquidity delta sign for the given liquidity delta.
	// This is called when consuming all liquidity.
	// When a position is created, we add liquidity to lower tick
	// and subtract from the upper tick to reflect that this new
	// liquidity would be added when the price crosses the lower tick
	// going up, and subtracted when the price crosses the upper tick
	// going up. As a result, the sign depends on the direction we are moving.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	SetLiquidityDeltaSign(liquidityDelta sdk.Dec) sdk.Dec
	// ValidateSqrtPrice validates the given square root price
	// relative to the current square root price on one side of the bound
	// and the min/max sqrt price on the other side.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	ValidateSqrtPrice(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error
}

// New returns a swap strategy based on the provided zeroForOne parameter
// with sqrtPriceLimit for the maximum square root price until which to perform
// the swap and the stor key of the module that stores swap data.
func New(zeroForOne bool, sqrtPriceLimit sdk.Dec, storeKey sdk.StoreKey, spreadFactor sdk.Dec, tickSpacing uint64) swapStrategy {
	if zeroForOne {
		return &zeroForOneStrategy{sqrtPriceLimit: sqrtPriceLimit, storeKey: storeKey, spreadFactor: spreadFactor, tickSpacing: tickSpacing}
	}
	return &oneForZeroStrategy{sqrtPriceLimit: sqrtPriceLimit, storeKey: storeKey, spreadFactor: spreadFactor, tickSpacing: tickSpacing}
}

// GetPriceLimit returns the price limit based on which token is being swapped in.
// If zero in for one out, the price is decreasing. Therefore, min spot price is the limit.
// If one in for zero out, the price is increasing. Therefore, max spot price is the limit.
func GetPriceLimit(zeroForOne bool) sdk.Dec {
	if zeroForOne {
		return types.MinSpotPrice
	}
	return types.MaxSpotPrice
}
