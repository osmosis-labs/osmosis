package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
func liquidity0(amount sdk.Int, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	product := sqrtPriceA.Mul(sqrtPriceB)
	diff := sqrtPriceB.Sub(sqrtPriceA)
	return amount.ToDec().Mul(product.Quo(diff))
}

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity1 = amount / (sqrtPriceB - sqrtPriceA)
func liquidity1(amount sdk.Int, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	return amount.ToDec().Quo(diff)
}

// calcAmount0 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount0Delta = (liquidity * (sqrtPriceA - sqrtPriceB)) / (sqrtPriceA * sqrtPriceB)
func calcAmount0Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	mult := liq
	return (mult.Mul(diff)).Quo(sqrtPriceB).Quo(sqrtPriceA)
}

// calcAmount1 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount1Delta = liq * (sqrtPriceA - sqrtPriceB)
func calcAmount1Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	return liq.Mul(diff)
}

// computeSwapStep calculates the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity, and amount available to swap
// lte is reference to "less than or equal", which determines if we are moving left or right of the current price to find the next initialized tick with liquidity
func computeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec, zeroForOne bool) (sqrtPriceNext, amountIn, amountOut sdk.Dec) {
	if zeroForOne {
		amountIn = calcAmount0Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent)
	} else {
		amountIn = calcAmount1Delta(liquidity, sqrtPriceTarget, sqrtPriceCurrent)
	}

	if amountRemaining.GTE(amountIn) {
		sqrtPriceNext = sqrtPriceTarget
	} else {
		sqrtPriceNext = getNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining, zeroForOne)
	}

	if zeroForOne {
		amountIn = calcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
		amountOut = calcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
	} else {
		amountIn = calcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
		amountOut = calcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
	}

	return sqrtPriceNext, amountIn, amountOut
}

func getNextSqrtPriceFromInput(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec, zeroForOne bool) (sqrtPriceNext sdk.Dec) {
	if zeroForOne {
		sqrtPriceNext = getNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining)
	} else {
		sqrtPriceNext = getNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining)
	}
	return sqrtPriceNext
}

// getNextSqrtPriceFromAmount0RoundingUp utilizes the current squareRootPrice, liquidity of denom0, and amount of denom0 that still needs
// to be swapped in order to determine the next squareRootPrice
func getNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	numerator := liquidity.Mul(sdk.NewDec(2))
	product := amountRemaining.Mul(sqrtPriceCurrent)

	if product.Quo(amountRemaining).Equal(sqrtPriceCurrent) {
		denominator := numerator.Add(product)
		if denominator.GTE(numerator) {
			numerator = numerator.Mul(sqrtPriceCurrent)
			sqrtPriceNext = numerator.QuoRoundUp(denominator)
			return sqrtPriceNext
		}
	}
	denominator := numerator.Quo(sqrtPriceCurrent).Add(amountRemaining)
	sqrtPriceNext = numerator.QuoRoundUp(denominator)
	return sqrtPriceNext
}

// getNextSqrtPriceFromAmount1RoundingDown utilizes the current squareRootPrice, liquidity of denom1, and amount of denom1 that still needs
// to be swapped in order to determine the next squareRootPrice
func getNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return sqrtPriceCurrent.Add(amountRemaining.Quo(liquidity))
}

// getLiquidityFromAmounts takes the current sqrtPrice and the sqrtPrice for the upper and lower ticks as well as the amounts of asset0 and asset1
// in return, liquidity is calculated from these inputs
func getLiquidityFromAmounts(sqrtPrice, sqrtPriceA, sqrtPriceB sdk.Dec, amount0, amount1 sdk.Int) (liquidity sdk.Dec) {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	if sqrtPrice.LTE(sqrtPriceA) {
		liquidity = liquidity0(amount0, sqrtPrice, sqrtPriceB)
	} else if sqrtPrice.LTE(sqrtPriceB) {
		liquidity0 := liquidity0(amount0, sqrtPrice, sqrtPriceB)
		liquidity1 := liquidity1(amount1, sqrtPrice, sqrtPriceA)
		liquidity = sdk.MinDec(liquidity0, liquidity1)
	} else {
		liquidity = liquidity1(amount1, sqrtPrice, sqrtPriceA)
	}
	return liquidity
}
