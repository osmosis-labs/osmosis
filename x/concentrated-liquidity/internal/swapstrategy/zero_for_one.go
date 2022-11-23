package swapstrategy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

type zeroForOneStrategy struct {
	sqrtPriceLimit sdk.Dec
}

var _ swapStrategy = (*zeroForOneStrategy)(nil)

// ComputeSwapStep calculates the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity, and amount available to swap
// lte is reference to "less than or equal", which determines if we are moving left or right of the current price to find the next initialized tick with liquidity
func (s *zeroForOneStrategy) ComputeSwapStep3(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext, amountIn, amountOut sdk.Dec) {
	amountIn = math.CalcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)
	if amountRemaining.GTE(amountIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		sqrtPriceNext = s.GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	amountIn = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	amountOut = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	return sqrtPriceNext, amountIn, amountOut
}

// ComputeSwapStep calculates the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity, and amount available to swap
// lte is reference to "less than or equal", which determines if we are moving left or right of the current price to find the next initialized tick with liquidity
func (s *zeroForOneStrategy) ComputeSwapStep(sqrtPriceCurrent, nextSqrtPrice, liquidity, amountRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec) {
	// as long as the nextSqrtPrice (calculated above) is within the user defined price limit, we set it as the target sqrtPrice
	// if it is outside the user defined price limit, we set the target sqrtPrice to the user defined price limit
	if nextSqrtPrice.LT(s.sqrtPriceLimit) {
		nextSqrtPrice = s.sqrtPriceLimit
	}

	amountIn := math.CalcAmount0Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	if amountRemaining.LT(amountIn) {
		nextSqrtPrice = s.GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	amountIn = math.CalcAmount0Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	amountOut := math.CalcAmount1Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)

	return nextSqrtPrice, amountIn, amountOut
}

func (s *zeroForOneStrategy) GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return math.GetNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining)
}

func (s *zeroForOneStrategy) SetLiquidityDeltaSign(deltaLiquidity sdk.Dec) sdk.Dec {
	return deltaLiquidity.Neg()
}

func (s *zeroForOneStrategy) SetNextTick(nextTick int64) sdk.Int {
	return sdk.NewInt(nextTick - 1)
}

func (s *zeroForOneStrategy) ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt ratio if we are swapping asset0 for asset1
	if sqrtPriceLimit.GT(currentSqrtPrice) || sqrtPriceLimit.LT(types.MinSqrtRatio) {
		return types.InvalidPriceLimitError{SqrtPriceLimit: sqrtPriceLimit, LowerBound: types.MinSqrtRatio, UpperBound: sqrtPriceLimit}
	}
	return nil
}
