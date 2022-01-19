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

func solveConstantFunctionInvariant(
	tokenBalanceFixed,
	tokenWeightFixed,
	tokenBalanceUnknown,
	tokenWeightUnknown,
	tokenAmountFixed sdk.Dec,
) sdk.Dec {
	weightRatio := tokenWeightFixed.Quo(tokenWeightUnknown)
	y := tokenBalanceFixed.Quo(tokenBalanceFixed.Add(tokenAmountFixed))
	foo := osmomath.Pow(y, weightRatio)

	multiplier := sdk.OneDec().Sub(foo)
	return tokenBalanceUnknown.Mul(multiplier)
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
	// deduct swapfee on the in asset
	tokenAmountIn = tokenAmountIn.Mul(sdk.OneDec().Sub(swapFee))
	tokenAmountOut := solveConstantFunctionInvariant(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut, tokenAmountIn)
	return tokenAmountOut
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
	// provide negative tokenOutAmount as it decreases pool liquidity
	tokenAmountIn := solveConstantFunctionInvariant(tokenBalanceOut, tokenWeightOut, tokenBalanceIn, tokenWeightIn, tokenAmountOut.Neg()).Neg()
	// We deduct a swap fee on the input asset. The swap happens by following the invariant curve on the input * (1 - swap fee)
	//  and then the swap fee is added to the pool.
	// Thus in order to give X amount out, we solve the invariant for the invariant input. However invariant input = (1 - swapfee) * trade input.
	// Therefore we divide by (1 - swapfee) here
	tokenAmountInBeforeFee := tokenAmountIn.Quo(sdk.OneDec().Sub(swapFee))
	return tokenAmountInBeforeFee

}

func weightDelta(
	normalizedWeight,
	beforeWeight,
	afterWeight sdk.Dec,
) (unknownNewBalance sdk.Dec) {
	deltaRatio := afterWeight.Quo(beforeWeight)
	// newBalTo = poolRatio^(1/weightTo) * balTo;
	return osmomath.Pow(deltaRatio, normalizedWeight)
}

func feeRatio(
	normalizedWeight,
	swapFee sdk.Dec,
) sdk.Dec {
	zar := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	return sdk.OneDec().Sub(zar)
}

func tokenDiffGivenPoolDiff(
	normalizedWeight,
	tokenBalance,
	poolSupply,
	poolAmount sdk.Dec,
) sdk.Dec {
	newPoolSupply := poolSupply.Add(poolAmount)

	poolSupplyDelta := weightDelta(
		sdk.OneDec().Quo(normalizedWeight), poolSupply, newPoolSupply)

	newTokenBalance := tokenBalance.Mul(poolSupplyDelta)

	// Do reverse order of fees charged in joinswap_ExternAmountIn, this way
	//     ``` pAo == joinswap_ExternAmountIn(Ti, joinswap_PoolAmountOut(pAo, Ti)) ```
	//uint tAi = tAiAfterFee / (1 - (1-weightTi) * swapFee) ;
	return newTokenBalance.Sub(tokenBalance)
}

func poolDiffGivenTokenDiff(
	normalizedWeight,
	tokenBalance,
	poolSupply,
	tokenAmount sdk.Dec,
) sdk.Dec {
	newTokenBalance := tokenBalance.Add(tokenAmount)

	tokenDelta := weightDelta(
		normalizedWeight, tokenBalance, newTokenBalance)
	newPoolSupply := poolSupply.Mul(tokenDelta)

	return newPoolSupply.Sub(poolSupply)
}

//tAi
func calcSingleInGivenPoolOut(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	poolSupply,
	poolAmountOut,
	swapFee sdk.Dec,
) sdk.Dec {
	tokenAmountIn := tokenDiffGivenPoolDiff(normalizedTokenWeightIn, tokenBalanceIn, poolSupply, poolAmountOut)
	tokenAmountInBeforeFee := tokenAmountIn.Quo(feeRatio(normalizedTokenWeightIn, swapFee))
	return tokenAmountInBeforeFee
}

// pAo
func calcPoolOutGivenSingleIn(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	poolSupply,
	tokenAmountIn,
	swapFee sdk.Dec,
) sdk.Dec {
	tokenAmountInAfterFee := tokenAmountIn.Mul(feeRatio(normalizedTokenWeightIn, swapFee))
	poolAmountOut := poolDiffGivenTokenDiff(normalizedTokenWeightIn, tokenBalanceIn, poolSupply, tokenAmountInAfterFee)
	return poolAmountOut
}

// tAo
func calcSingleOutGivenPoolIn(
	tokenBalanceOut,
	normalizedTokenWeightOut,
	poolSupply,
	poolAmountIn,
	swapFee sdk.Dec,
	exitFee sdk.Dec,
) sdk.Dec {
	// charge exit fee on the pool token side
	// pAiAfterExitFee = pAi*(1-exitFee)
	poolAmountInAfterExitFee := poolAmountIn.Mul(sdk.OneDec().Sub(exitFee))
	tokenAmountOut := tokenDiffGivenPoolDiff(normalizedTokenWeightOut, tokenBalanceOut, poolSupply, poolAmountInAfterExitFee.Neg()).Neg()
	tokenAmountOutAfterFee := tokenAmountOut.Mul(feeRatio(normalizedTokenWeightOut, swapFee))
	return tokenAmountOutAfterFee
}

// pAi
func calcPoolInGivenSingleOut(
	tokenBalanceOut,
	normalizedTokenWeightOut,
	poolSupply,
	tokenAmountOut,
	swapFee sdk.Dec,
	exitFee sdk.Dec,
) sdk.Dec {
	tokenAmountOutBeforeFee := tokenAmountOut.Quo(feeRatio(normalizedTokenWeightOut, swapFee))

	poolAmountIn := poolDiffGivenTokenDiff(normalizedTokenWeightOut, tokenBalanceOut, poolSupply, tokenAmountOutBeforeFee.Neg()).Neg()

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	poolAmountInBeforeFee := poolAmountIn.Quo(sdk.OneDec().Sub(exitFee))
	return poolAmountInBeforeFee
}
