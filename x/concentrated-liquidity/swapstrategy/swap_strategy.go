package swapstrategy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

// swapStrategy defines the interface for computing a swap.
// There are 2 implementations of this interface:
// - zeroForOneStrategy to provide implementations when swapping token 0 for token 1.
// - oneForZeroStrategy to provide implementations when swapping token 1 for token 0.
type SwapStrategy interface {
	// GetSqrtTargetPrice returns the target square root price given the next tick square root price
	// upon comparing it to sqrt price limit.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	GetSqrtTargetPrice(nextTickSqrtPrice sdk.Dec) sdk.Dec
	// ComputeSwapWithinBucketOutGivenIn calculates the next sqrt price, the amount of token in consumed, the amount out to return to the user, and total spread reward charge on token in.
	// This assumes swapping over a single bucket where the liqudiity stays constant until we cross the next initialized tick of the next bucket.
	// Parameters:
	//   * sqrtPriceCurrent is the current sqrt price.
	//   * sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
	//       - Next initialized tick sqrt price.
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
	ComputeSwapWithinBucketOutGivenIn(sqrtPriceCurrent osmomath.BigDec, sqrtPriceTarget, liquidity, amountRemainingIn sdk.Dec) (sqrtPriceNext osmomath.BigDec, amountInConsumed, amountOutComputed, spreadRewardChargeTotal sdk.Dec)
	// ComputeSwapWithinBucketInGivenOut calculates the next sqrt price, the amount of token out consumed, the amount in to charge to the user for requested out, and total spread reward charge on token in.
	// This assumes swapping over a single bucket where the liqudiity stays constant until we cross the next initialized tick of the next bucket.
	// Parameters:
	//   * sqrtPriceCurrent is the current sqrt price.
	//   * sqrtPriceTarget is the target sqrt price computed with GetSqrtTargetPrice(). It must be one of:
	//       - Next initialized tick sqrt price.
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
	ComputeSwapWithinBucketInGivenOut(sqrtPriceCurrent osmomath.BigDec, sqrtPriceTarget, liquidity, amountRemainingOut sdk.Dec) (sqrtPriceNext osmomath.BigDec, amountOutConsumed, amountInComputed, spreadRewardChargeTotal sdk.Dec)
	// InitializeNextTickIterator returns iterator that seeks to the next tick from the given tickIndex.
	// If nex tick relative to tickINdex does not exist in the store, it will return an invalid iterator.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	InitializeNextTickIterator(ctx sdk.Context, poolId uint64, tickIndex int64) dbm.Iterator
	// SetLiquidityDeltaSign sets the liquidity delta sign for the given liquidity delta.
	// This is called when consuming all liquidity.
	// When a position is created, we add liquidity to lower tick
	// and subtract from the upper tick to reflect that this new
	// liquidity would be added when the price crosses the lower tick
	// going up, and subtracted when the price crosses the upper tick
	// going up. As a result, the sign depends on the direction we are moving.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	SetLiquidityDeltaSign(liquidityDelta sdk.Dec) sdk.Dec
	// UpdateTickAfterCrossing updates the next tick after crossing
	// to satisfy our "position in-range" invariant which is:
	// lower tick <= current tick < upper tick
	// When crossing a tick in zero for one direction, we move
	// left on the range. As a result, we end up crossing the lower tick
	// that is inclusive. Therefore, we must decrease the next tick
	// by 1 additional unit so that it falls under the current range.
	// When crossing a tick in one for zero direction, we move
	// right on the range. As a result, we end up crossing the upper tick
	// that is exclusive. Therefore, we leave the next tick as is since
	// it is already excluded from the current range.
	UpdateTickAfterCrossing(nextTick int64) (updatedNextTick int64)
	// ValidateSqrtPrice validates the given square root price
	// relative to the current square root price on one side of the bound
	// and the min/max sqrt price on the other side.
	// See oneForZeroStrategy or zeroForOneStrategy for implementation details.
	ValidateSqrtPrice(sqrtPriceLimit sdk.Dec, currentSqrtPrice osmomath.BigDec) error

	ZeroForOne() bool
}

var (
	oneBigDec = osmomath.OneDec()
)

// New returns a swap strategy based on the provided zeroForOne parameter
// with sqrtPriceLimit for the maximum square root price until which to perform
// the swap and the stor key of the module that stores swap data.
func New(zeroForOne bool, sqrtPriceLimit sdk.Dec, storeKey sdk.StoreKey, spreadFactor sdk.Dec) SwapStrategy {
	if zeroForOne {
		return &zeroForOneStrategy{sqrtPriceLimit: sqrtPriceLimit, storeKey: storeKey, spreadFactor: spreadFactor}
	}
	return &oneForZeroStrategy{sqrtPriceLimit: sqrtPriceLimit, storeKey: storeKey, spreadFactor: spreadFactor}
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

func GetSqrtPriceLimit(priceLimit sdk.Dec, zeroForOne bool) (sdk.Dec, error) {
	if priceLimit.IsZero() {
		if zeroForOne {
			return types.MinSqrtPrice, nil
		}
		return types.MaxSqrtPrice, nil
	}

	return osmomath.MonotonicSqrt(priceLimit)
}
