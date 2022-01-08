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

func calcTokenGivenToken(
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
	tokenAmountOut := calcTokenGivenToken(tokenBalanceIn, tokenWeightIn, tokenBalanceOut, tokenWeightOut, tokenAmountIn)
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
	tokenAmountOut = tokenAmountOut.Neg()
	tokenAmountIn := calcTokenGivenToken(tokenBalanceOut, tokenWeightOut, tokenBalanceIn, tokenWeightIn, tokenAmountOut).Neg()
	// deduct swapfee on the in asset
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
	tokenWeight,
	totalWeight,
	swapFee sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeight.Quo(totalWeight)
	zar := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	return sdk.OneDec().Sub(zar)
}

func tokenDiffGivenPoolDiff(
	tokenBalance,
	tokenWeight,
	poolSupply,
	totalWeight,
	poolAmount sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeight.Quo(totalWeight)
	newPoolSupply := poolSupply.Add(poolAmount)

	poolSupplyDelta := weightDelta(
		sdk.OneDec().Quo(normalizedWeight), poolSupply, newPoolSupply)

	newTokenBalance := tokenBalance.Mul(poolSupplyDelta)

	// Do reverse order of fees charged in joinswap_ExternAmountIn, this way
	//     ``` pAo == joinswap_ExternAmountIn(Ti, joinswap_PoolAmountOut(pAo, Ti)) ```
	//uint tAi = tAiAfterFee / (1 - (1-weightTi) * swapFee) ;
	return newTokenBalance.Sub(tokenBalance)
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
	tokenAmountIn := tokenDiffGivenPoolDiff(tokenBalanceIn, tokenWeightIn, poolSupply, totalWeight, poolAmountOut)
	tokenAmountInBeforeFee := tokenAmountIn.Quo(feeRatio(tokenWeightIn, totalWeight, swapFee))
	return tokenAmountInBeforeFee
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
	// charge exit fee on the pool token side
	// pAiAfterExitFee = pAi*(1-exitFee)
	poolAmountInAfterExitFee := poolAmountIn.Mul(sdk.OneDec().Sub(exitFee))
	tokenAmountOut := tokenDiffGivenPoolDiff(tokenBalanceOut, tokenWeightOut, poolSupply, totalWeight, poolAmountInAfterExitFee.Neg()).Neg()
	tokenAmountOutAfterFee := tokenAmountOut.Mul(feeRatio(tokenWeightOut, totalWeight, swapFee))
	return tokenAmountOutAfterFee
}

func poolDiffGivenTokenDiff(
	tokenBalance,
	tokenWeight,
	poolSupply,
	totalWeight,
	tokenAmount sdk.Dec,
) sdk.Dec {
	normalizedWeight := tokenWeight.Quo(totalWeight)

	newTokenBalance := tokenBalance.Add(tokenAmount)

	tokenDelta := weightDelta(
		normalizedWeight, tokenBalance, newTokenBalance)
	newPoolSupply := poolSupply.Mul(tokenDelta)

	return newPoolSupply.Sub(poolSupply)
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
	tokenAmountInAfterFee := tokenAmountIn.Mul(feeRatio(tokenWeightIn, totalWeight, swapFee))
	poolAmountOut := poolDiffGivenTokenDiff(tokenBalanceIn, tokenWeightIn, poolSupply, totalWeight, tokenAmountInAfterFee)
	return poolAmountOut
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
	tokenAmountOutBeforeFee := tokenAmountOut.Quo(feeRatio(tokenWeightOut, totalWeight, swapFee))

	poolAmountIn := poolDiffGivenTokenDiff(tokenBalanceOut, tokenWeightOut, poolSupply, totalWeight, tokenAmountOutBeforeFee.Neg()).Neg()

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	poolAmountInBeforeFee := poolAmountIn.Quo(sdk.OneDec().Sub(exitFee))
	return poolAmountInBeforeFee
}
