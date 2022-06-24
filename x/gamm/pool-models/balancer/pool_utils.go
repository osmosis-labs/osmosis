package balancer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// balancer notation: pAo - pool shares amount out, given single asset in
// the second argument requires the tokenWeightIn / total token weight.
func calcPoolSharesOutGivenSingleAssetIn(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	poolShares,
	tokenAmountIn,
	swapFee sdk.Dec,
) sdk.Dec {
	// deduct swapfee on the in asset.
	// We don't charge swap fee on the token amount that we imagine as unswapped (the normalized weight).
	// So effective_swapfee = swapfee * (1 - normalized_token_weight)
	tokenAmountInAfterFee := tokenAmountIn.Mul(feeRatio(normalizedTokenWeightIn, swapFee))
	// To figure out the number of shares we add, first notice that in balancer we can treat
	// the number of shares as linearly related to the `k` value function. This is due to the normalization.
	// e.g.
	// if x^.5 y^.5 = k, then we `n` x the liquidity to `(nx)^.5 (ny)^.5 = nk = k'`
	// We generalize this linear relation to do the liquidity add for the not-all-asset case.
	// Suppose we increase the supply of x by x', so we want to solve for `k'/k`.
	// This is `(x + x')^{weight} * old_terms / (x^{weight} * old_terms) = (x + x')^{weight} / (x^{weight})`
	// The number of new shares we need to make is then `old_shares * ((k'/k) - 1)`
	// Whats very cool, is that this turns out to be the exact same `solveConstantFunctionInvariant` code
	// with the answer's sign reversed.
	poolAmountOut := solveConstantFunctionInvariant(
		tokenBalanceIn.Add(tokenAmountInAfterFee),
		tokenBalanceIn,
		normalizedTokenWeightIn,
		poolShares,
		sdk.OneDec()).Neg()
	return poolAmountOut
}

// getPoolAssetsByDenom return a mapping from pool asset
// denom to the pool asset itself. There must be no duplicates.
// Returns error, if any found.
func getPoolAssetsByDenom(poolAssets []PoolAsset) (map[string]PoolAsset, error) {
	poolAssetsByDenom := make(map[string]PoolAsset)
	for _, poolAsset := range poolAssets {
		_, ok := poolAssetsByDenom[poolAsset.Token.Denom]
		if ok {
			return nil, fmt.Errorf(errMsgFormatRepeatingPoolAssetsNotAllowed, poolAsset.Token.Denom)
		}

		poolAssetsByDenom[poolAsset.Token.Denom] = poolAsset
	}
	return poolAssetsByDenom, nil
}

// updateIntermediaryPoolAssetsLiquidity updates poolAssetsByDenom with liquidity.
//
// all liqidity coins must exist in poolAssetsByDenom. Returns error, if not.
//
// This is a helper function that is useful for updating the pool asset amounts
// as an intermediary step in a multi-join methods such as CalcJoinPoolShares.
// In CalcJoinPoolShares with multi-asset joins, we first attempt to do
// a MaximalExactRatioJoin that might leave out some tokens in.
// Then, for every remaining tokens in, we attempt to do a single asset join.
// Since the first step (MaximalExactRatioJoin) affects the pool liqudity due to slippage,
// we would like to account for that in the subsequent steps of single asset join.
func updateIntermediaryPoolAssetsLiquidity(liquidity sdk.Coins, poolAssetsByDenom map[string]PoolAsset) error {
	for _, coin := range liquidity {
		poolAsset, ok := poolAssetsByDenom[coin.Denom]
		if !ok {
			return fmt.Errorf(errMsgFormatFailedInterimLiquidityUpdate, coin.Denom)
		}

		poolAsset.Token.Amount = poolAssetsByDenom[coin.Denom].Token.Amount.Add(coin.Amount)
		poolAssetsByDenom[coin.Denom] = poolAsset
	}
	return nil
}

// feeRatio returns the fee ratio that is defined as follows:
// 1 - ((1 - normalizedTokenWeightOut) * swapFee)
func feeRatio(normalizedWeight, swapFee sdk.Dec) sdk.Dec {
	return sdk.OneDec().Sub((sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee))
}

// calcSingleAssetInGivenPoolSharesOut returns token amount in with fee included
// given the swapped out shares amount, using solveConstantFunctionInvariant
func calcSingleAssetInGivenPoolSharesOut(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	totalPoolSharesSupply,
	sharesAmountOut,
	swapFee sdk.Dec,
) sdk.Dec {
	// delta balanceIn is negative(tokens inside the pool increases)
	// pool weight is always 1
	tokenAmountIn := solveConstantFunctionInvariant(totalPoolSharesSupply.Add(sharesAmountOut), totalPoolSharesSupply, sdk.OneDec(), tokenBalanceIn, normalizedTokenWeightIn).Neg()
	// deduct swapfee on the in asset
	tokenAmountInFeeIncluded := tokenAmountIn.Quo(feeRatio(normalizedTokenWeightIn, swapFee))
	return tokenAmountInFeeIncluded
}

// calcPoolSharesInGivenSingleAssetOut returns pool shares amount in, given single asset out.
// the returned shares in have the fee included in them.
// the second argument requires the tokenWeightOut / total token weight.
func calcPoolSharesInGivenSingleAssetOut(
	tokenBalanceOut,
	normalizedTokenWeightOut,
	totalPoolSharesSupply,
	tokenAmountOut,
	swapFee,
	exitFee sdk.Dec,
) sdk.Dec {
	tokenAmountOutFeeIncluded := tokenAmountOut.Quo(feeRatio(normalizedTokenWeightOut, swapFee))

	// delta poolSupply is positive(total pool shares decreases)
	// pool weight is always 1
	sharesIn := solveConstantFunctionInvariant(tokenBalanceOut.Sub(tokenAmountOutFeeIncluded), tokenBalanceOut, normalizedTokenWeightOut, totalPoolSharesSupply, sdk.OneDec())

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	sharesInFeeIncluded := sharesIn.Quo(sdk.OneDec().Sub(exitFee))
	return sharesInFeeIncluded
}
