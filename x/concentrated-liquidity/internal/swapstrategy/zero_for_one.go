package swapstrategy

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/osmomath"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
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
}

var _ swapStrategy = (*zeroForOneStrategy)(nil)

// ComputeSwapStep calculates the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity,
// and amount available to swap
//
// zeroForOneStrategy assumes moving to the left of the current square root price. amountRemaining or
// amountIn is the amount of token 0. amountOut is token 1. The calculations and validations are performed
// accordingly.
func (s zeroForOneStrategy) ComputeSwapStep(sqrtPriceCurrent, nextSqrtPrice, liquidity, amountRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec) {
	// as long as the nextSqrtPrice (calculated above) is within the user defined price limit, we set it as the target sqrtPrice
	// if it is outside the user defined price limit, we set the target sqrtPrice to the user defined price limit
	if nextSqrtPrice.LT(s.sqrtPriceLimit) {
		nextSqrtPrice = s.sqrtPriceLimit
	}

	amountIn := math.CalcAmount0Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	if amountRemaining.LT(amountIn) {
		nextSqrtPrice = math.GetNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	amountIn = math.CalcAmount0Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	amountOut := math.CalcAmount1Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)

	return nextSqrtPrice, amountIn, amountOut
}

// NextInitializedTick returns the next initialized tick index based on the
// provided tickindex. If no initialized tick exists, <0, false>
// will be returned.
//
// zeroForOneStrategy searches for the next tick to the left of the current tickIndex.
func (s zeroForOneStrategy) NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64) (next int64, initialized bool) {
	store := ctx.KVStore(s.storeKey)

	// Construct a prefix store with a prefix of <TickPrefix | poolID>, allowing
	// us to retrieve the next initialized tick without having to scan all ticks.
	prefixBz := types.KeyTickPrefix(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	// When looking to the left of the current tick, we need to evaluate the
	// current tick as well. The end cursor for reverse iteration is non-inclusive
	// so must add one and handle overflow.
	startKey := types.TickIndexToBytes(osmomath.Max(tickIndex, tickIndex+1))

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
			return tick, true
		}
	}
	return 0, false
}

func (s zeroForOneStrategy) SetLiquidityDeltaSign(deltaLiquidity sdk.Dec) sdk.Dec {
	return deltaLiquidity.Neg()
}

func (s zeroForOneStrategy) SetNextTick(nextTick int64) sdk.Int {
	return sdk.NewInt(nextTick - 1)
}

func (s zeroForOneStrategy) ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt ratio if we are swapping asset0 for asset1
	if sqrtPriceLimit.GT(currentSqrtPrice) || sqrtPriceLimit.LT(types.MinSqrtRatio) {
		return types.InvalidPriceLimitError{SqrtPriceLimit: sqrtPriceLimit, LowerBound: types.MinSqrtRatio, UpperBound: sqrtPriceLimit}
	}
	return nil
}
