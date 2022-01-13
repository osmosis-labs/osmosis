package balancer

import (
	"github.com/osmosis-labs/osmosis/osmomath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func weightDelta(
	normalizedWeight,
	beforeWeight,
	afterWeight sdk.Dec,
) (unknownNewBalance sdk.Dec) {
	deltaRatio := afterWeight.Quo(beforeWeight)
	// newBalTo = poolRatio^(1/weightTo) * balTo;
	return osmomath.Pow(deltaRatio, normalizedWeight)
}

func solveTokenFromShare(
	tokenBalance,
	tokenWeight,
	poolSupply,
	poolAmount sdk.Dec,
) sdk.Dec {
	newPoolSupply := poolSupply.Add(poolAmount)

	poolSupplyDelta := weightDelta(
		sdk.OneDec().Quo(tokenWeight), poolSupply, newPoolSupply)

	newTokenBalance := tokenBalance.Mul(poolSupplyDelta)

	// Do reverse order of fees charged in joinswap_ExternAmountIn, this way
	//     ``` pAo == joinswap_ExternAmountIn(Ti, joinswap_PoolAmountOut(pAo, Ti)) ```
	//uint tAi = tAiAfterFee / (1 - (1-weightTi) * swapFee) ;
	return newTokenBalance.Sub(tokenBalance)
}

func solveShareFromToken(
	tokenBalance,
	tokenWeight,
	poolSupply,
	tokenAmount sdk.Dec,
) sdk.Dec {
	newTokenBalance := tokenBalance.Add(tokenAmount)

	tokenDelta := weightDelta(
		tokenWeight, tokenBalance, newTokenBalance)
	newPoolSupply := poolSupply.Mul(tokenDelta)

	return newPoolSupply.Sub(poolSupply)
}
