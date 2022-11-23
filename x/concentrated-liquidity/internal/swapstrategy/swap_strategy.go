package swapstrategy

import sdk "github.com/cosmos/cosmos-sdk/types"

type swapStrategy interface {
	GetNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec)
	ComputeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext, amountIn, amountOut sdk.Dec)
	SetLiquidityDeltaSign(sdk.Dec) sdk.Dec
	SetNextTick(int64) sdk.Int
	ValidatePriceLimit(sqrtPriceLimit, currentSqrtPrice sdk.Dec) error
}

func New(zeroForOne bool, sqrtPriceLimit sdk.Dec) swapStrategy {
	if zeroForOne {
		return &zeroForOneStrategy{sqrtPriceLimit: sqrtPriceLimit}
	}
	return &oneForZeroStrategy{sqrtPriceLimit: sqrtPriceLimit}
}
