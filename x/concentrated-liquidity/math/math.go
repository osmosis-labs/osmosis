package math

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// Liquidity0 = amount0 * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)
// TODO: Define rounding properties we expect to hold for this function.
func Liquidity0(amount osmomath.Int, sqrtPriceA, sqrtPriceB osmomath.BigDec) osmomath.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}

	// We convert to BigDec to avoid precision loss when calculating liquidity. Without doing this,
	// our liquidity calculations will be off from our theoretical calculations within our tests.
	amountBigDec := osmomath.BigDecFromDec(amount.ToLegacyDec())

	product := sqrtPriceA.Mul(sqrtPriceB)
	diff := sqrtPriceB.Sub(sqrtPriceA)
	if diff.IsZero() {
		panic(fmt.Sprintf("liquidity0 diff is zero: sqrtPriceA %s sqrtPriceB %s", sqrtPriceA, sqrtPriceB))
	}

	return amountBigDec.MulMut(product).QuoMut(diff).Dec()
}

// Liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// Liquidity1 = amount1 / (sqrtPriceB - sqrtPriceA)
func Liquidity1(amount osmomath.Int, sqrtPriceA, sqrtPriceB osmomath.BigDec) osmomath.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}

	// We convert to BigDec to avoid precision loss when calculating liquidity. Without doing this,
	// our liquidity calculations will be off from our theoretical calculations within our tests.
	amountBigDec := osmomath.BigDecFromDec(amount.ToLegacyDec())
	diff := sqrtPriceB.Sub(sqrtPriceA)
	if diff.IsZero() {
		panic(fmt.Sprintf("liquidity1 diff is zero: sqrtPriceA %s sqrtPriceB %s", sqrtPriceA, sqrtPriceB))
	}

	return amountBigDec.QuoMut(diff).Dec()
}

// CalcAmount0Delta takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// CalcAmount0Delta = (liquidity * (sqrtPriceB - sqrtPriceA)) / (sqrtPriceB * sqrtPriceA)
func CalcAmount0Delta(liq, sqrtPriceA, sqrtPriceB osmomath.BigDec, roundUp bool) osmomath.BigDec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	// if calculating for amountIn, we round up
	// if calculating for amountOut, we round down at precision end
	// this is to prevent removing more from the pool than expected due to rounding
	// example: we calculate 1000000.9999999 uusdc (~$1) amountIn and 2000000.999999 uosmo amountOut
	// we would want the user to put in 1000001 uusdc rather than 1000000 uusdc to ensure we are charging enough for the amount they are removing
	// additionally, without rounding, there exists cases where the swapState.amountSpecifiedRemaining.IsPositive() for loop within
	// the CalcOut/In functions never actually reach zero due to dust that would have never gotten counted towards the amount (numbers after the 10^6 place)
	if roundUp {
		// Note that we do MulTruncate so that the denominator is smaller as this is
		// the case where we want to round up to favor the pool.
		// Examples include:
		// - calculating amountIn during swap
		// - adding liquidity (request user to provide more tokens in in favor of the pool)
		// The denominator is truncated to get a higher final amount.
		denom := sqrtPriceA.MulTruncate(sqrtPriceB)
		return liq.Mul(diff).QuoMut(denom).Ceil()
	}
	// These are truncated at precision end to round in favor of the pool when:
	// - calculating amount out during swap
	// - withdrawing liquidity
	// The denominator is rounded up to get a smaller final amount.
	denom := sqrtPriceA.MulRoundUp(sqrtPriceB)

	return liq.MulTruncate(diff).QuoTruncate(denom)
}

// CalcAmount1Delta takes the asset with the smaller liquidity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// CalcAmount1Delta = liq * (sqrtPriceB - sqrtPriceA)
func CalcAmount1Delta(liq, sqrtPriceA, sqrtPriceB osmomath.BigDec, roundUp bool) osmomath.BigDec {
	// make sqrtPriceA the smaller value amongst sqrtPriceA and sqrtPriceB
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	// if calculating for amountIn, we round up
	// if calculating for amountOut, we don't round at all
	// this is to prevent removing more from the pool than expected due to rounding
	// example: we calculate 1000000.9999999 uusdc (~$1) amountIn and 2000000.999999 uosmo amountOut
	// we would want the used to put in 1000001 uusdc rather than 1000000 uusdc to ensure we are charging enough for the amount they are removing
	// additionally, without rounding, there exists cases where the swapState.amountSpecifiedRemaining.IsPositive() for loop within
	// the CalcOut/In functions never actually reach zero due to dust that would have never gotten counted towards the amount (numbers after the 10^6 place)
	if roundUp {
		// Note that we do MulRoundUp so that the end result is larger as this is
		// the case where we want to round up to favor the pool.
		// Examples include:
		// - calculating amountIn during swap
		// - adding liquidity (request user to provide more tokens in in favor of the pool)
		return liq.Mul(diff).Ceil()
	}
	// This is truncated at precision end to round in favor of the pool when:
	// - calculating amount out during swap
	// - withdrawing liquidity
	// The denominator is rounded up to get a higher final amount.
	return liq.MulTruncate(diff)
}

// GetNextSqrtPriceFromAmount0InRoundingUp utilizes sqrtPriceCurrent, liquidity, and amount of denom0 that still needs
// to be swapped in order to determine the sqrtPriceNext.
// When we swap for token one out given token zero in, the price is decreasing, and we need to move the sqrt price (decrease it) less
// to avoid overpaying the amount out of the pool. Therefore, we round up.
// sqrt_next = liq * sqrt_cur / (liq + token_in * sqrt_cur)
func GetNextSqrtPriceFromAmount0InRoundingUp(sqrtPriceCurrent, liquidity, amountZeroRemainingIn osmomath.BigDec) (sqrtPriceNext osmomath.BigDec) {
	if amountZeroRemainingIn.IsZero() {
		return sqrtPriceCurrent
	}

	product := amountZeroRemainingIn.Mul(sqrtPriceCurrent)
	// denominator = product + liquidity
	denominator := product
	denominator.AddMut(liquidity)
	return liquidity.Mul(sqrtPriceCurrent).QuoRoundUp(denominator)
}

// GetNextSqrtPriceFromAmount0OutRoundingUp utilizes sqrtPriceCurrent, liquidity, and amount of denom0 that still needs
// to be swapped out order to determine the sqrtPriceNext.
// When we swap for token one in given token zero out, the price is increasing and we need to move the price up enough
// so that we get the desired output amount out. Therefore, we round up.
// sqrt_next = liq * sqrt_cur / (liq - token_out * sqrt_cur)
func GetNextSqrtPriceFromAmount0OutRoundingUp(sqrtPriceCurrent, liquidity, amountZeroRemainingOut osmomath.BigDec) (sqrtPriceNext osmomath.BigDec) {
	if amountZeroRemainingOut.IsZero() {
		return sqrtPriceCurrent
	}

	// mul round up to make the final denominator smaller and final result larger
	product := amountZeroRemainingOut.MulRoundUp(sqrtPriceCurrent)
	denominator := liquidity.Sub(product)
	// mul round up numerator to make the final result larger
	// quo round up to make the final result larger
	return liquidity.MulRoundUp(sqrtPriceCurrent).QuoRoundUp(denominator)
}

// GetNextSqrtPriceFromAmount1InRoundingDown utilizes the current sqrtPriceCurrent, liquidity, and amount of denom1 that still needs
// to be swapped in order to determine the sqrtPriceNext.
// When we swap for token zero out given token one in, the price is increasing and we need to move the sqrt price (increase it) less to
// avoid overpaying out of the pool. Therefore, we round down.
// sqrt_next = sqrt_cur + token_in / liq
func GetNextSqrtPriceFromAmount1InRoundingDown(sqrtPriceCurrent, liquidity, amountOneRemainingIn osmomath.BigDec) (sqrtPriceNext osmomath.BigDec) {
	return sqrtPriceCurrent.Add(amountOneRemainingIn.QuoTruncate(liquidity))
}

// GetNextSqrtPriceFromAmount1OutRoundingDown utilizes the current sqrtPriceCurrent, liquidity, and amount of denom1 that still needs
// to be swapped out order to determine the sqrtPriceNext.
// When we swap for token zero in given token one out, the price is decrearing and we need to move the price down enough
// so that we get the desired output amount out.
// sqrt_next = sqrt_cur - token_out / liq
func GetNextSqrtPriceFromAmount1OutRoundingDown(sqrtPriceCurrent, liquidity, amountOneRemainingOut osmomath.BigDec) (sqrtPriceNext osmomath.BigDec) {
	return sqrtPriceCurrent.Sub(amountOneRemainingOut.QuoRoundUp(liquidity))
}

// GetLiquidityFromAmounts takes the current sqrtPrice and the sqrtPrice for the upper and lower ticks as well as the amounts of asset0 and asset1
// and returns the resulting liquidity from these inputs.
func GetLiquidityFromAmounts(sqrtPrice osmomath.BigDec, sqrtPriceA, sqrtPriceB osmomath.BigDec, amount0, amount1 osmomath.Int) (liquidity osmomath.Dec) {
	// Reorder the prices so that sqrtPriceA is the smaller of the two.
	// todo: Remove this.
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}

	if sqrtPrice.LTE(sqrtPriceA) {
		// If the current price is less than or equal to the lower tick, then we use the liquidity0 formula.
		liquidity = Liquidity0(amount0, sqrtPriceA, sqrtPriceB)
	} else if sqrtPrice.LT(sqrtPriceB) {
		// If the current price is between the lower and upper ticks (exclusive of both the lower and upper ticks,
		// as both would trigger a division by zero), then we use the minimum of the liquidity0 and liquidity1 formulas.
		liquidity0 := Liquidity0(amount0, sqrtPrice, sqrtPriceB)
		liquidity1 := Liquidity1(amount1, sqrtPrice, sqrtPriceA)
		liquidity = osmomath.MinDec(liquidity0, liquidity1)
	} else {
		// If the current price is greater than the upper tick, then we use the liquidity1 formula.
		liquidity = Liquidity1(amount1, sqrtPriceB, sqrtPriceA)
	}

	return liquidity
}

// SquareRoundUp squares and rounds up at precision end.
func SquareRoundUp(sqrtPrice osmomath.Dec) osmomath.Dec {
	return sqrtPrice.MulRoundUp(sqrtPrice)
}

// SquareTruncate squares and truncates at precision end.
func SquareTruncate(sqrtPrice osmomath.Dec) osmomath.Dec {
	return sqrtPrice.MulTruncate(sqrtPrice)
}
