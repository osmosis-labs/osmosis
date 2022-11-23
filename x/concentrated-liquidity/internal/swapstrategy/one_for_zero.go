package swapstrategy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

type oneForZeroStrategy struct {
	sqrtPriceLimit sdk.Dec
}

var _ swapStrategy = (*oneForZeroStrategy)(nil)

func (s *oneForZeroStrategy) ComputeSwapStep3(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext, amountIn, amountOut sdk.Dec) {
	amountIn = math.CalcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent, false)
	if amountRemaining.GTE(amountIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		sqrtPriceNext = s.GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	amountIn = math.CalcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)
	amountOut = math.CalcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent, false)

	return sqrtPriceNext, amountIn, amountOut
}

func (s *oneForZeroStrategy) ComputeSwapStep(sqrtPriceCurrent, nextSqrtPrice, liquidity, amountRemaining sdk.Dec) (sdk.Dec, sdk.Dec, sdk.Dec) {
	// as long as the nextSqrtPrice (calculated above) is within the user defined price limit, we set it as the target sqrtPrice
	// if it is outside the user defined price limit, we set the target sqrtPrice to the user defined price limit
	if nextSqrtPrice.GT(s.sqrtPriceLimit) {
		nextSqrtPrice = s.sqrtPriceLimit
	}

	amountIn := math.CalcAmount1Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	if amountRemaining.LT(amountIn) {
		nextSqrtPrice = s.GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	amountIn = math.CalcAmount1Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)
	amountOut := math.CalcAmount0Delta(liquidity, nextSqrtPrice, sqrtPriceCurrent, false)

	return nextSqrtPrice, amountIn, amountOut
}

func (s *oneForZeroStrategy) GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return math.GetNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining)
}

func (s *oneForZeroStrategy) SetLiquidityDeltaSign(deltaLiquidity sdk.Dec) sdk.Dec {
	return deltaLiquidity
}

func (s *oneForZeroStrategy) SetNextTick(nextTick int64) sdk.Int {
	return sdk.NewInt(nextTick)
}

func (s *oneForZeroStrategy) ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error {
	// check that the price limit is below the current sqrt price but not lower than the minimum sqrt ratio if we are swapping asset0 for asset1
	if sqrtPriceLimit.LT(currentSqrtPrice) || sqrtPriceLimit.GT(types.MaxSqrtRatio) {
		return types.InvalidPriceLimitError{SqrtPriceLimit: sqrtPriceLimit, LowerBound: currentSqrtPrice, UpperBound: types.MaxSqrtRatio}
	}
	return nil
}
