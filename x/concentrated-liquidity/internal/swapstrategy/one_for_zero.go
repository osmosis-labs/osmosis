package swapstrategy

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
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
}

var _ swapStrategy = (*oneForZeroStrategy)(nil)

// ComputeSwapStep calculates the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity,
// and amount available to swap
//
// oneForZeroStrategy assumes moving to the right of the current square root price. amountRemaining or
// amountIn is the amount of token 1. amountOut is token 0.
func (s oneForZeroStrategy) ComputeSwapStep(sqrtPriceCurrent, nextSqrtPrice, liquidity, amountRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec) {
	// as long as the nextSqrtPrice (calculated above) is within the user defined price limit, we set it as the target sqrtPrice
	// if it is outside the user defined price limit, we set the target sqrtPrice to the user defined price limit
	if nextSqrtPrice.GT(s.sqrtPriceLimit) {
		nextSqrtPrice = s.sqrtPriceLimit
	}

	amountIn := math.CalcAmount1Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	if amountRemaining.LT(amountIn) {
		nextSqrtPrice = math.GetNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	amountIn = math.CalcAmount1Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	amountOut := math.CalcAmount0Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)

	return nextSqrtPrice, amountIn, amountOut
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
	prefixBz := types.KeyTickPrefix(poolId)
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

// ValidatePriceLimit validates the given square root price limit
// given the square root price.
//
// oneForZeroStrategy assumes moving to the right of the current square root price.
// Therefore, the following invariant must hold:
// current square root price <= sqrtPriceLimit <= types.MaxSqrtRatio
func (s oneForZeroStrategy) ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is above the current sqrt price but lower than the maximum sqrt ratio since we are swapping asset1 for asset0
	if sqrtPriceLimit.LT(currentSqrtPrice) || sqrtPriceLimit.GT(types.MaxSqrtRatio) {
		return types.InvalidPriceLimitError{SqrtPriceLimit: sqrtPriceLimit, LowerBound: currentSqrtPrice, UpperBound: types.MaxSqrtRatio}
	}
	return nil
}
