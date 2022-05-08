package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Don't EVER change after initializing
// TODO: Analyze choice here
var powPrecision, _ = sdk.NewDecFromStr("0.00000001")

// Singletons
var zero sdk.Dec = sdk.ZeroDec()

var (
	one_half sdk.Dec = sdk.MustNewDecFromStr("0.5")
	one      sdk.Dec = sdk.OneDec()
	two      sdk.Dec = sdk.MustNewDecFromStr("2")
)

// calcSpotPrice returns the spot price of the pool
// This is the weight-adjusted balance of the tokens in the pool.
// so spot_price = (B_in / W_in) / (B_out / W_out)
func calcSpotPrice(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut sdk.Dec,
) sdk.Dec {
	number := tokenBalanceIn.Quo(tokenWeightIn)
	denom := tokenBalanceOut.Quo(tokenWeightOut)
	ratio := number.Quo(denom)

	return ratio
}

// calcSpotPriceWithSwapFee returns the spot price of the pool accounting for
// the input taken by the swap fee.
// This is the weight-adjusted balance of the tokens in the pool.
// so spot_price = (B_in / W_in) / (B_out / W_out)
// and spot_price_with_fee = spot_price / (1 - swapfee)
func calcSpotPriceWithSwapFee(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut,
	swapFee sdk.Dec,
) sdk.Dec {
	spotPrice := calcSpotPrice(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut)
	// Q: Why is this not just (1 - swapfee)
	// A: Its because its being applied to the other asset.
	// TODO: write this up more coherently
	// 1 / (1 - swapfee)
	scale := sdk.OneDec().Quo(sdk.OneDec().Sub(swapFee))

	return spotPrice.Mul(scale)
}

// aO
func calcOutGivenIn(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut,
	tokenAmountIn,
	swapFee sdk.Dec,
) sdk.Dec {
	weightRatio := tokenWeightIn.Quo(tokenWeightOut)
	adjustedIn := sdk.OneDec().Sub(swapFee)
	adjustedIn = tokenAmountIn.Mul(adjustedIn)
	y := tokenBalanceIn.Quo(tokenBalanceIn.Add(adjustedIn))
	foo := pow(y, weightRatio)
	bar := sdk.OneDec().Sub(foo)
	return tokenBalanceOut.Mul(bar)
}

// aI
func calcInGivenOut(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut,
	tokenAmountOut,
	swapFee sdk.Dec,
) sdk.Dec {
	weightRatio := tokenWeightOut.Quo(tokenWeightIn)
	diff := tokenBalanceOut.Sub(tokenAmountOut)
	y := tokenBalanceOut.Quo(diff)
	foo := pow(y, weightRatio)
	foo = foo.Sub(one)
	tokenAmountIn := sdk.OneDec().Sub(swapFee)
	return (tokenBalanceIn.Mul(foo)).Quo(tokenAmountIn)
}

// pAo
func calcPoolOutGivenSingleIn(
	tokenBalanceIn,
	tokenWeightIn,
	poolSupply,
	totalWeight,
	tokenAmountIn,
	swapFee sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeightIn.Quo(totalWeight)
	zaz := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	tokenAmountInAfterFee := tokenAmountIn.Mul(sdk.OneDec().Sub(zaz))

	newTokenBalanceIn := tokenBalanceIn.Add(tokenAmountInAfterFee)
	tokenInRatio := newTokenBalanceIn.Quo(tokenBalanceIn)

	// uint newPoolSupply = (ratioTi ^ weightTi) * poolSupply;
	poolRatio := pow(tokenInRatio, normalizedWeight)
	newPoolSupply := poolRatio.Mul(poolSupply)
	return newPoolSupply.Sub(poolSupply)
}

// tAi
func calcSingleInGivenPoolOut(
	tokenBalanceIn,
	tokenWeightIn,
	poolSupply,
	totalWeight,
	poolAmountOut,
	swapFee sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeightIn.Quo(totalWeight)
	newPoolSupply := poolSupply.Add(poolAmountOut)
	poolRatio := newPoolSupply.Quo(poolSupply)

	// uint newBalTi = poolRatio^(1/weightTi) * balTi;
	boo := sdk.OneDec().Quo(normalizedWeight)
	tokenInRatio := pow(poolRatio, boo)
	newTokenBalanceIn := tokenInRatio.Mul(tokenBalanceIn)
	tokenAmountInAfterFee := newTokenBalanceIn.Sub(tokenBalanceIn)
	// Do reverse order of fees charged in joinswap_ExternAmountIn, this way
	//     ``` pAo == joinswap_ExternAmountIn(Ti, joinswap_PoolAmountOut(pAo, Ti)) ```
	// uint tAi = tAiAfterFee / (1 - (1-weightTi) * swapFee) ;
	zar := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	return tokenAmountInAfterFee.Quo(sdk.OneDec().Sub(zar))
}

// tAo
func calcSingleOutGivenPoolIn(
	tokenBalanceOut,
	tokenWeightOut,
	poolSupply,
	totalWeight,
	poolAmountIn,
	swapFee sdk.Dec,
	exitFee sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeightOut.Quo(totalWeight)
	// charge exit fee on the pool token side
	// pAiAfterExitFee = pAi*(1-exitFee)
	poolAmountInAfterExitFee := poolAmountIn.Mul(sdk.OneDec().Sub(exitFee))
	newPoolSupply := poolSupply.Sub(poolAmountInAfterExitFee)
	poolRatio := newPoolSupply.Quo(poolSupply)

	// newBalTo = poolRatio^(1/weightTo) * balTo;

	tokenOutRatio := pow(poolRatio, sdk.OneDec().Quo(normalizedWeight))
	newTokenBalanceOut := tokenOutRatio.Mul(tokenBalanceOut)

	tokenAmountOutBeforeSwapFee := tokenBalanceOut.Sub(newTokenBalanceOut)

	// charge swap fee on the output token side
	// uint tAo = tAoBeforeSwapFee * (1 - (1-weightTo) * swapFee)
	zaz := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	tokenAmountOut := tokenAmountOutBeforeSwapFee.Mul(sdk.OneDec().Sub(zaz))
	return tokenAmountOut
}

// pAi
func calcPoolInGivenSingleOut(
	tokenBalanceOut,
	tokenWeightOut,
	poolSupply,
	totalWeight,
	tokenAmountOut,
	swapFee sdk.Dec,
	exitFee sdk.Dec,
) sdk.Dec {
	// charge swap fee on the output token side
	normalizedWeight := tokenWeightOut.Quo(totalWeight)
	// uint tAoBeforeSwapFee = tAo / (1 - (1-weightTo) * swapFee) ;
	zoo := sdk.OneDec().Sub(normalizedWeight)
	zar := zoo.Mul(swapFee)
	tokenAmountOutBeforeSwapFee := tokenAmountOut.Quo(sdk.OneDec().Sub(zar))

	newTokenBalanceOut := tokenBalanceOut.Sub(tokenAmountOutBeforeSwapFee)
	tokenOutRatio := newTokenBalanceOut.Quo(tokenBalanceOut)

	// uint newPoolSupply = (ratioTo ^ weightTo) * poolSupply;
	poolRatio := pow(tokenOutRatio, normalizedWeight)
	newPoolSupply := poolRatio.Mul(poolSupply)
	poolAmountInAfterExitFee := poolSupply.Sub(newPoolSupply)

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	return poolAmountInAfterExitFee.Quo(sdk.OneDec().Sub(exitFee))
}

/*********************************************************/

// absDifferenceWithSign returns | a - b |, (a - b).sign()
func absDifferenceWithSign(a, b sdk.Dec) (sdk.Dec, bool) {
	if a.GTE(b) {
		return a.Sub(b), false
	} else {
		return b.Sub(a), true
	}
}

// func largeBasePow(base sdk.Dec, exp sdk.Dec) sdk.Dec {
// 	// pow requires the base to be <= 2
// }

// pow computes base^(exp)
// However since the exponent is not an integer, we must do an approximation algorithm.
// TODO: In the future, lets add some optimized routines for common exponents, e.g. for common wIn / wOut ratios
// Many simple exponents like 2:1 pools
func pow(base sdk.Dec, exp sdk.Dec) sdk.Dec {
	// Exponentiation of a negative base with an arbitrary real exponent is not closed within the reals.
	// You can see this by recalling that `i = (-1)^(.5)`. We have to go to complex numbers to define this.
	// (And would have to implement complex logarithms)
	// We don't have a need for negative bases, so we don't include any such logic.
	if !base.IsPositive() {
		panic(fmt.Errorf("base must be greater than 0"))
	}
	// TODO: Remove this if we want to generalize the function,
	// we can adjust the algorithm in this setting.
	if base.GTE(two) {
		panic(fmt.Errorf("base must be lesser than two"))
	}

	// We will use an approximation algorithm to compute the power.
	// Since computing an integer power is easy, we split up the exponent into
	// an integer component and a fractional component.
	integer := exp.TruncateDec()
	fractional := exp.Sub(integer)

	integerPow := base.Power(uint64(integer.TruncateInt64()))

	if fractional.IsZero() {
		return integerPow
	}

	fractionalPow := powApprox(base, fractional, powPrecision)

	return integerPow.Mul(fractionalPow)
}

// Contract: 0 < base <= 2
// 0 < exp < 1
func powApprox(base sdk.Dec, exp sdk.Dec, precision sdk.Dec) sdk.Dec {
	if exp.IsZero() {
		return sdk.ZeroDec()
	}

	// Common case optimization
	// Optimize for it being equal to one-half
	if exp.Equal(one_half) {
		output, err := base.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		return output
	}
	// TODO: Make an approx-equal function, and then check if exp * 3 = 1, and do a check accordingly

	// We compute this via taking the maclaurin series of (1 + x)^a
	// where x = base - 1.
	// The maclaurin series of (1 + x)^a = sum_{k=0}^{infty} binom(a, k) x^k
	// Binom(a, k) takes the natural continuation on the first parameter, namely that
	// Binom(a, k) = N/D, where D = k!, and N = a(a-1)(a-2)...(a-k+1)
	// Next we show that the absolute value of each term is less than the last term.
	// Note that the change in term n's value vs term n + 1 is a multiplicative factor of
	// v_n = x(a - n) / (n+1)
	// So if |v_n| < 1, we know that each term has a lesser impact on the result than the last.
	// For our bounds on |x| < 1, |a| < 1,
	// it suffices to see for what n is |v_n| < 1,
	// in the worst parameterization of x = 1, a = -1.
	// v_n = |(-1 + epsilon - n) / (n+1)|
	// So |v_n| is always less than 1, as n ranges over the integers.
	//
	// Note that term_n of the expansion is 1 * prod_{i=0}^{n-1} v_i
	// The error if we stop the expansion at term_n is:
	// error_n = sum_{k=n+1}^{infty} term_k
	// At this point we further restrict a >= 0, so 0 <= a < 1.
	// Now we take the _INCORRECT_ assumption that if term_n < p, then
	// error_n < p.
	// This assumption is obviously wrong.
	// However our usages of this function don't use the full domain.
	// With a > 0, |x| << 1, and p sufficiently low, perhaps this actually is true.

	// TODO: Check with our parameterization
	// TODO: If theres a bug, balancer is also wrong here :thonk:
	a := exp
	x, xneg := absDifferenceWithSign(base, one)
	term := sdk.OneDec()
	sum := sdk.OneDec()
	negative := false

	// TODO: Document this computation via taylor expansion
	for i := 1; term.GTE(precision); i++ {
		bigK := sdk.OneDec().MulInt64(int64(i))
		c, cneg := absDifferenceWithSign(a, bigK.Sub(one))
		term = term.Mul(c.Mul(x))
		term = term.Quo(bigK)

		if term.IsZero() {
			break
		}
		if xneg {
			negative = !negative
		}

		if cneg {
			negative = !negative
		}

		if negative {
			sum = sum.Sub(term)
		} else {
			sum = sum.Add(term)
		}
	}
	return sum
}
