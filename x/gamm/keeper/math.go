package keeper

import (
	"github.com/osmosis-labs/osmosis/v7/osmomath"

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

// solveConstantFunctionInvariant solves the constant function of an AMM
// that determines the relationship between the differences of two sides
// of assets inside the pool.
// For fixed balanceXBefore, balanceXAfter, weightX, balanceY, weightY,
// we could deduce the balanceYDelta, calculated by:
// balanceYDelta = balanceY * (1 - (balanceXBefore/balanceXAfter)^(weightX/weightY))
// balanceYDelta is positive when the balance liquidity decreases.
// balanceYDelta is negative when the balance liquidity increases.
func solveConstantFunctionInvariant(
	tokenBalanceFixedBefore,
	tokenBalanceFixedAfter,
	tokenWeightFixed,
	tokenBalanceUnknownBefore,
	tokenWeightUnknown sdk.Dec,
) sdk.Dec {
	// weightRatio = (weightX/weightY)
	weightRatio := tokenWeightFixed.Quo(tokenWeightUnknown)

	// y = balanceXBefore/balanceYAfter
	y := tokenBalanceFixedBefore.Quo(tokenBalanceFixedAfter)

	// amountY = balanceY * (1 - (y ^ weightRatio))
	foo := osmomath.Pow(y, weightRatio)
	multiplier := sdk.OneDec().Sub(foo)
	return tokenBalanceUnknownBefore.Mul(multiplier)
}

// calcOutGivenIn calculates token to be swapped out given
// the provided amount, fee deducted, using solveConstantFunctionInvariant
func calcOutGivenIn(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut,
	tokenAmountIn,
	swapFee sdk.Dec,
) sdk.Dec {
	// deduct swapfee on the in asset
	tokenAmountInAfterFee := tokenAmountIn.Mul(sdk.OneDec().Sub(swapFee))
	// delta balanceOut is positive(tokens inside the pool decreases)
	tokenAmountOut := solveConstantFunctionInvariant(tokenBalanceIn, tokenBalanceIn.Add(tokenAmountInAfterFee), tokenWeightIn, tokenBalanceOut, tokenWeightOut)
	return tokenAmountOut
}

// calcInGivenOut calculates token to be provided, fee added,
// given the swapped out amount, using solveConstantFunctionInvariant
func calcInGivenOut(
	tokenBalanceIn,
	tokenWeightIn,
	tokenBalanceOut,
	tokenWeightOut,
	tokenAmountOut,
	swapFee sdk.Dec,
) sdk.Dec {
	// delta balanceIn is negative(amount of tokens inside the pool increases)
	tokenAmountIn := solveConstantFunctionInvariant(tokenBalanceOut, tokenBalanceOut.Sub(tokenAmountOut), tokenWeightOut, tokenBalanceIn, tokenWeightIn).Neg()
	// We deduct a swap fee on the input asset. The swap happens by following the invariant curve on the input * (1 - swap fee)
	//  and then the swap fee is added to the pool.
	// Thus in order to give X amount out, we solve the invariant for the invariant input. However invariant input = (1 - swapfee) * trade input.
	// Therefore we divide by (1 - swapfee) here
	tokenAmountInBeforeFee := tokenAmountIn.Quo(sdk.OneDec().Sub(swapFee))
	return tokenAmountInBeforeFee

}

func feeRatio(
	normalizedWeight,
	swapFee sdk.Dec,
) sdk.Dec {
	zar := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	return sdk.OneDec().Sub(zar)
}

// calcSingleInGivenPoolOut calculates token to be provided, fee added,
// given the swapped out shares amount, using solveConstantFunctionInvariant
func calcSingleInGivenPoolOut(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	poolSupply,
	poolAmountOut,
	swapFee sdk.Dec,
) sdk.Dec {
	// delta balanceIn is negative(tokens inside the pool increases)
	// pool weight is always 1
	tokenAmountIn := solveConstantFunctionInvariant(poolSupply.Add(poolAmountOut), poolSupply, sdk.OneDec(), tokenBalanceIn, normalizedTokenWeightIn).Neg()
	// deduct swapfee on the in asset
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
	// deduct swapfee on the in asset
	tokenAmountInAfterFee := tokenAmountIn.Mul(feeRatio(normalizedTokenWeightIn, swapFee))
	// delta poolSupply is negative(total pool shares increases)
	// pool weight is always 1
	poolAmountOut := solveConstantFunctionInvariant(tokenBalanceIn.Add(tokenAmountInAfterFee), tokenBalanceIn, normalizedTokenWeightIn, poolSupply, sdk.OneDec()).Neg()
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

	// delta balanceOut is positive(tokens inside the pool decreases)
	// pool weight is always 1
	tokenAmountOut := solveConstantFunctionInvariant(poolSupply.Sub(poolAmountInAfterExitFee), poolSupply, sdk.OneDec(), tokenBalanceOut, normalizedTokenWeightOut)
	// deduct
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

	// delta poolSupply is positive(total pool shares decreases)
	// pool weight is always 1
	poolAmountIn := solveConstantFunctionInvariant(tokenBalanceOut.Sub(tokenAmountOutBeforeFee), tokenBalanceOut, normalizedTokenWeightOut, poolSupply, sdk.OneDec())

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	poolAmountInBeforeFee := poolAmountIn.Quo(sdk.OneDec().Sub(exitFee))
	return poolAmountInBeforeFee
}
