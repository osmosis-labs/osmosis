package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

var (
	smallestDec = sdk.SmallestDec()
)

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// Liquidity0 = amount0 * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)
func Liquidity0(amount sdk.Int, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}

	// We convert to BigDec to avoid precision loss when calculating liquidity. Without doing this,
	// our liquidity calculations will be off from our theoretical calculations within our tests.
	amountBigDec := osmomath.BigDecFromSDKDec(amount.ToDec())
	sqrtPriceABigDec := osmomath.BigDecFromSDKDec(sqrtPriceA)
	sqrtPriceBBigDec := osmomath.BigDecFromSDKDec(sqrtPriceB)

	product := sqrtPriceABigDec.Mul(sqrtPriceBBigDec)
	diff := sqrtPriceBBigDec.Sub(sqrtPriceABigDec)
	return amountBigDec.Mul(product).Quo(diff).SDKDec()
}

// Liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// Liquidity1 = amount1 / (sqrtPriceB - sqrtPriceA)
func Liquidity1(amount sdk.Int, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}

	// We convert to BigDec to avoid precision loss when calculating liquidity. Without doing this,
	// our liquidity calculations will be off from our theoretical calculations within our tests.
	amountBigDec := osmomath.BigDecFromSDKDec(amount.ToDec())
	sqrtPriceABigDec := osmomath.BigDecFromSDKDec(sqrtPriceA)
	sqrtPriceBBigDec := osmomath.BigDecFromSDKDec(sqrtPriceB)

	diff := sqrtPriceBBigDec.Sub(sqrtPriceABigDec)
	return amountBigDec.Quo(diff).SDKDec()
}

// CalcAmount0 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// CalcAmount0Delta = (liquidity * (sqrtPriceB - sqrtPriceA)) / (sqrtPriceB * sqrtPriceA)
func CalcAmount0Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec, roundUp bool) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	denom := sqrtPriceA.Mul(sqrtPriceB)
	// if calculating for amountIn, we round up
	// if calculating for amountOut, we don't round at all
	// this is to prevent removing more from the pool than expected due to rounding
	// example: we calculate 1000000.9999999 uusdc (~$1) amountIn and 2000000.999999 uosmo amountOut
	// we would want the user to put in 1000001 uusdc rather than 1000000 uusdc to ensure we are charging enough for the amount they are removing
	// additionally, without rounding, there exists cases where the swapState.amountSpecifiedRemaining.GT(sdk.ZeroDec()) for loop within
	// the CalcOut/In functions never actually reach zero due to dust that would have never gotten counted towards the amount (numbers after the 10^6 place)
	if roundUp {
		return liq.Mul(diff).Quo(denom).Ceil()
	}
	return liq.Mul(diff).Quo(denom)
}

// CalcAmount1 takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// CalcAmount1Delta = liq * (sqrtPriceB - sqrtPriceA)
func CalcAmount1Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec, roundUp bool) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	// if calculating for amountIn, we round up
	// if calculating for amountOut, we don't round at all
	// this is to prevent removing more from the pool than expected due to rounding
	// example: we calculate 1000000.9999999 uusdc (~$1) amountIn and 2000000.999999 uosmo amountOut
	// we would want the used to put in 1000001 uusdc rather than 1000000 uusdc to ensure we are charging enough for the amount they are removing
	// additionally, without rounding, there exists cases where the swapState.amountSpecifiedRemaining.GT(sdk.ZeroDec()) for loop within
	// the CalcOut/In functions never actually reach zero due to dust that would have never gotten counted towards the amount (numbers after the 10^6 place)
	if roundUp {
		return liq.Mul(diff).Ceil()
	}
	return liq.Mul(diff)
}

// getNextSqrtPriceFromAmount0RoundingUp utilizes the current squareRootPrice, liquidity of denom0, and amount of denom0 that still needs
// to be swapped in order to determine the next squareRootPrice
// Note: we are using only using the precise formula here.
func GetNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	if amountRemaining.Equal(sdk.ZeroDec()) {
		return sqrtPriceCurrent
	}

	product := amountRemaining.Mul(sqrtPriceCurrent)
	denominator := liquidity.Add(product)
	return liquidity.Mul(sqrtPriceCurrent).QuoRoundUp(denominator)
}

// getNextSqrtPriceFromAmount1RoundingDown utilizes the current squareRootPrice, liquidity of denom1, and amount of denom1 that still needs
// to be swapped in order to determine the next squareRootPrice
func GetNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return sqrtPriceCurrent.Add(amountRemaining.QuoTruncate(liquidity))
}

// getLiquidityFromAmounts takes the current sqrtPrice and the sqrtPrice for the upper and lower ticks as well as the amounts of asset0 and asset1
// in return, liquidity is calculated from these inputs
func GetLiquidityFromAmounts(sqrtPrice, sqrtPriceA, sqrtPriceB sdk.Dec, amount0, amount1 sdk.Int) (liquidity sdk.Dec) {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	if sqrtPrice.LTE(sqrtPriceA) {
		liquidity = Liquidity0(amount0, sqrtPriceA, sqrtPriceB)
	} else if sqrtPrice.LTE(sqrtPriceB) {
		liquidity0 := Liquidity0(amount0, sqrtPrice, sqrtPriceB)
		liquidity1 := Liquidity1(amount1, sqrtPrice, sqrtPriceA)
		liquidity = sdk.MinDec(liquidity0, liquidity1)
	} else {
		liquidity = Liquidity1(amount1, sqrtPriceB, sqrtPriceA)
	}
	return liquidity
}

func AddLiquidity(liquidityA, liquidityB sdk.Dec) (finalLiquidity sdk.Dec) {
	if liquidityB.LT(sdk.ZeroDec()) {
		return liquidityA.Sub(liquidityB.Abs())
	}
	return liquidityA.Add(liquidityB)
}

// MulRoundUp multiplies a by b and rounds up to the nearest integer
// at precision end.
func MulRoundUp(a, b sdk.Dec) sdk.Dec {
	return a.MulTruncate(b).Add(smallestDec)
}
