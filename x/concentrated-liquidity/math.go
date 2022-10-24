package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
func liquidity0(amount int64, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	product := sqrtPriceA.Mul(sqrtPriceB)
	diff := sqrtPriceB.Sub(sqrtPriceA)
	amt := sdk.NewDec(amount)
	return amt.Mul(product.Quo(diff))
}

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity1 = amount / (sqrtPriceB - sqrtPriceA)
func liquidity1(amount int64, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	amt := sdk.NewDec(amount)
	return amt.Quo(diff)
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
func computeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec, lte bool) (sqrtPriceNext sdk.Dec, amountIn sdk.Dec, amountOut sdk.Dec) {
	if lte {
		sqrtPriceNext = getNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining)
		amountIn = calcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
		amountOut = calcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
	} else {
		sqrtPriceNext = getNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining)
		amountIn = calcAmount0Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
		amountOut = calcAmount1Delta(liquidity, sqrtPriceNext, sqrtPriceCurrent)
	}
	return sqrtPriceNext, amountIn, amountOut
}

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

func getNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return sqrtPriceCurrent.Add(amountRemaining.Quo(liquidity))
}
