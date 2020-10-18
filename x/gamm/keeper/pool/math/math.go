package math

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Don't EVER change after initializing
var powPrecision, _ = sdk.NewDecFromStr("0.00000001")

//  sP
func calcSpotPrice(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut,
	swapFee sdk.Dec,
) sdk.Dec {
	number := tokenBalanceIn.Quo(tokenWeightIn)
	denom := tokenBalanceOut.Quo(tokenWeightOut)
	ratio := number.Quo(denom)
	scale := sdk.OneDec().Quo(sdk.OneDec().Sub(swapFee))

	return ratio.Mul(scale)
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
	foo = foo.Sub(sdk.OneDec())
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

//tAi
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

	//uint newBalTi = poolRatio^(1/weightTi) * balTi;
	boo := sdk.OneDec().Quo(normalizedWeight)
	tokenInRatio := pow(poolRatio, boo)
	newTokenBalanceIn := tokenInRatio.Mul(tokenBalanceIn)
	tokenAmountInAfterFee := newTokenBalanceIn.Sub(tokenBalanceIn)
	// Do reverse order of fees charged in joinswap_ExternAmountIn, this way
	//     ``` pAo == joinswap_ExternAmountIn(Ti, joinswap_PoolAmountOut(pAo, Ti)) ```
	//uint tAi = tAiAfterFee / (1 - (1-weightTi) * swapFee) ;
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
) sdk.Dec {
	normalizedWeight := tokenWeightOut.Quo(totalWeight)
	// charge exit fee on the pool token side
	// pAiAfterExitFee = pAi*(1-exitFee)
	poolAmountInAfterExitFee := poolAmountIn.Mul(sdk.OneDec())
	newPoolSupply := poolSupply.Sub(poolAmountInAfterExitFee)
	poolRatio := newPoolSupply.Quo(poolSupply)

	// newBalTo = poolRatio^(1/weightTo) * balTo;

	tokenOutRatio := pow(poolRatio, sdk.OneDec().Quo(normalizedWeight))
	newTokenBalanceOut := tokenOutRatio.Mul(tokenBalanceOut)

	tokenAmountOutBeforeSwapFee := tokenBalanceOut.Sub(newTokenBalanceOut)

	// charge swap fee on the output token side
	//uint tAo = tAoBeforeSwapFee * (1 - (1-weightTo) * swapFee)
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
) sdk.Dec {
	// charge swap fee on the output token side
	normalizedWeight := tokenWeightOut.Quo(totalWeight)
	//uint tAoBeforeSwapFee = tAo / (1 - (1-weightTo) * swapFee) ;
	zoo := sdk.OneDec().Sub(normalizedWeight)
	zar := zoo.Mul(swapFee)
	tokenAmountOutBeforeSwapFee := tokenAmountOut.Quo(sdk.OneDec().Sub(zar))

	newTokenBalanceOut := tokenBalanceOut.Sub(tokenAmountOutBeforeSwapFee)
	tokenOutRatio := newTokenBalanceOut.Quo(tokenBalanceOut)

	//uint newPoolSupply = (ratioTo ^ weightTo) * poolSupply;
	poolRatio := pow(tokenOutRatio, normalizedWeight)
	newPoolSupply := poolRatio.Mul(poolSupply)
	poolAmountInAfterExitFee := poolSupply.Sub(newPoolSupply)

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	return poolAmountInAfterExitFee.Quo(sdk.OneDec())
}

/*********************************************************/

func subSign(a, b sdk.Dec) (sdk.Dec, bool) {
	if a.GTE(b) {
		return a.Sub(b), false
	} else {
		return b.Sub(a), true
	}
}

func pow(base sdk.Dec, exp sdk.Dec) sdk.Dec {
	if base.LTE(sdk.ZeroDec()) {
		panic(fmt.Errorf("base have to be greater than zero"))
	}
	if base.GTE(sdk.OneDec().MulInt64(2)) {
		panic(fmt.Errorf("base have to be lesser than two"))
	}

	whole := exp.TruncateDec()
	remain := exp.Sub(whole)

	wholePow := base.Power(uint64(whole.TruncateInt64()))

	if remain.IsZero() {
		return wholePow
	}

	partialResult := powApprox(base, remain, powPrecision)

	return wholePow.Mul(partialResult)
}

func powApprox(base sdk.Dec, exp sdk.Dec, precision sdk.Dec) sdk.Dec {
	if base.LTE(sdk.ZeroDec()) {
		panic(fmt.Errorf("base have to be greater than zero"))
	}
	if base.GTE(sdk.OneDec().MulInt64(2)) {
		panic(fmt.Errorf("base have to be lesser than two"))
	}

	a := exp
	x, xneg := subSign(base, sdk.OneDec())
	term := sdk.OneDec()
	sum := sdk.OneDec()
	negative := false

	for i := 1; term.GTE(precision); i++ {
		bigK := sdk.OneDec().MulInt64(int64(i))
		c, cneg := subSign(a, bigK.Sub(sdk.OneDec()))
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
