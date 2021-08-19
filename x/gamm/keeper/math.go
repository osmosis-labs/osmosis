package keeper

import (
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Don't EVER change after initializing
// TODO: Analyze choice here
var powPrecision, _ = sdk.NewDecFromStr("0.00000001")

// Singletons
var zero sdk.Dec = sdk.ZeroDec()
var one_half sdk.Dec = sdk.MustNewDecFromStr("0.5")
var one sdk.Dec = sdk.OneDec()
var two sdk.Dec = sdk.MustNewDecFromStr("2")

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
	foo := osmomath.Pow(y, weightRatio)
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
	foo := osmomath.Pow(y, weightRatio)
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
	poolRatio := osmomath.Pow(tokenInRatio, normalizedWeight)
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
	tokenInRatio := osmomath.Pow(poolRatio, boo)
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
	exitFee sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeightOut.Quo(totalWeight)
	// charge exit fee on the pool token side
	// pAiAfterExitFee = pAi*(1-exitFee)
	poolAmountInAfterExitFee := poolAmountIn.Mul(sdk.OneDec().Sub(exitFee))
	newPoolSupply := poolSupply.Sub(poolAmountInAfterExitFee)
	poolRatio := newPoolSupply.Quo(poolSupply)

	// newBalTo = poolRatio^(1/weightTo) * balTo;

	tokenOutRatio := osmomath.Pow(poolRatio, sdk.OneDec().Quo(normalizedWeight))
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
	exitFee sdk.Dec,
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
	poolRatio := osmomath.Pow(tokenOutRatio, normalizedWeight)
	newPoolSupply := poolRatio.Mul(poolSupply)
	poolAmountInAfterExitFee := poolSupply.Sub(newPoolSupply)

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	return poolAmountInAfterExitFee.Quo(sdk.OneDec().Sub(exitFee))
}
