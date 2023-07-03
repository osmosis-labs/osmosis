package balancer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

// subPoolAssetWeights subtracts the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can (and probably will have some) be negative.
func subPoolAssetWeights(base []PoolAsset, other []PoolAsset) []PoolAsset {
	weightDifference := make([]PoolAsset, len(base))
	// TODO: Consider deleting these panics for performance
	if len(base) != len(other) {
		panic("subPoolAssetWeights called with invalid input, len(base) != len(other)")
	}
	for i, asset := range base {
		if asset.Token.Denom != other[i].Token.Denom {
			panic(fmt.Sprintf("subPoolAssetWeights called with invalid input, "+
				"expected other's %vth asset to be %v, got %v",
				i, asset.Token.Denom, other[i].Token.Denom))
		}
		curWeightDiff := asset.Weight.Sub(other[i].Weight)
		weightDifference[i] = PoolAsset{Token: asset.Token, Weight: curWeightDiff}
	}
	return weightDifference
}

// addPoolAssetWeights adds the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can be negative.
func addPoolAssetWeights(base []PoolAsset, other []PoolAsset) []PoolAsset {
	weightSum := make([]PoolAsset, len(base))
	// TODO: Consider deleting these panics for performance
	if len(base) != len(other) {
		panic("addPoolAssetWeights called with invalid input, len(base) != len(other)")
	}
	for i, asset := range base {
		if asset.Token.Denom != other[i].Token.Denom {
			panic(fmt.Sprintf("addPoolAssetWeights called with invalid input, "+
				"expected other's %vth asset to be %v, got %v",
				i, asset.Token.Denom, other[i].Token.Denom))
		}
		curWeightSum := asset.Weight.Add(other[i].Weight)
		weightSum[i] = PoolAsset{Token: asset.Token, Weight: curWeightSum}
	}
	return weightSum
}

// assumes 0 < d < 1
func poolAssetsMulDec(base []PoolAsset, d sdk.Dec) []PoolAsset {
	newWeights := make([]PoolAsset, len(base))
	for i, asset := range base {
		// TODO: This can adversarially panic at the moment! (as can Pool.TotalWeight)
		// Ensure this won't be able to panic in the future PR where we bound
		// each assets weight, and add precision
		newWeight := d.MulInt(asset.Weight).RoundInt()
		newWeights[i] = PoolAsset{Token: asset.Token, Weight: newWeight}
	}
	return newWeights
}

// ValidateUserSpecifiedWeight ensures that a weight that is provided from user-input anywhere
// for creating a pool obeys the expected guarantees.
// Namely, that the weight is in the range [1, MaxUserSpecifiedWeight)
func ValidateUserSpecifiedWeight(weight sdk.Int) error {
	if !weight.IsPositive() {
		return errorsmod.Wrap(types.ErrNotPositiveWeight, weight.String())
	}

	if weight.GTE(MaxUserSpecifiedWeight) {
		return errorsmod.Wrap(types.ErrWeightTooLarge, weight.String())
	}
	return nil
}

// solveConstantFunctionInvariant solves the constant function of an AMM
// that determines the relationship between the differences of two sides
// of assets inside the pool.
// For fixed balanceXBefore, balanceXAfter, weightX, balanceY, weightY,
// we could deduce the balanceYDelta, calculated by:
// balanceYDelta = balanceY * (1 - (balanceXBefore/balanceXAfter)^(weightX/weightY))
// balanceYDelta is positive when the balance liquidity decreases.
// balanceYDelta is negative when the balance liquidity increases.
//
// panics if tokenWeightUnknown is 0.
func solveConstantFunctionInvariant(
	tokenBalanceFixedBefore,
	tokenBalanceFixedAfter,
	tokenWeightFixed,
	tokenBalanceUnknownBefore,
	tokenWeightUnknown sdk.Dec,
) sdk.Dec {
	// weightRatio = (weightX/weightY)
	weightRatio := tokenWeightFixed.Quo(tokenWeightUnknown)

	// y = balanceXBefore/balanceXAfter
	y := tokenBalanceFixedBefore.Quo(tokenBalanceFixedAfter)

	// amountY = balanceY * (1 - (y ^ weightRatio))
	yToWeightRatio := osmomath.Pow(y, weightRatio)
	paranthetical := sdk.OneDec().Sub(yToWeightRatio)
	amountY := tokenBalanceUnknownBefore.Mul(paranthetical)
	return amountY
}

// balancer notation: pAo - pool shares amount out, given single asset in
// the second argument requires the tokenWeightIn / total token weight.
func calcPoolSharesOutGivenSingleAssetIn(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	poolShares,
	tokenAmountIn,
	spreadFactor sdk.Dec,
) sdk.Dec {
	// deduct spread factor on the in asset.
	// We don't charge spread factor on the token amount that we imagine as unswapped (the normalized weight).
	// So effective_swapfee = spread factor * (1 - normalized_token_weight)
	tokenAmountInAfterFee := tokenAmountIn.Mul(feeRatio(normalizedTokenWeightIn, spreadFactor))
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
			return nil, fmt.Errorf(formatRepeatingPoolAssetsNotAllowedErrFormat, poolAsset.Token.Denom)
		}

		poolAssetsByDenom[poolAsset.Token.Denom] = poolAsset
	}
	return poolAssetsByDenom, nil
}

// updateIntermediaryPoolAssetsLiquidity updates poolAssetsByDenom with liquidity.
//
// all liquidity coins must exist in poolAssetsByDenom. Returns error, if not.
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
			return fmt.Errorf(failedInterimLiquidityUpdateErrFormat, coin.Denom)
		}

		poolAsset.Token.Amount = poolAssetsByDenom[coin.Denom].Token.Amount.Add(coin.Amount)
		poolAssetsByDenom[coin.Denom] = poolAsset
	}
	return nil
}

// feeRatio returns the fee ratio that is defined as follows:
// 1 - ((1 - normalizedTokenWeightOut) * spreadFactor)
func feeRatio(normalizedWeight, spreadFactor sdk.Dec) sdk.Dec {
	return sdk.OneDec().Sub((sdk.OneDec().Sub(normalizedWeight)).Mul(spreadFactor))
}

// calcSingleAssetInGivenPoolSharesOut returns token amount in with fee included
// given the swapped out shares amount, using solveConstantFunctionInvariant
func calcSingleAssetInGivenPoolSharesOut(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	totalPoolSharesSupply,
	sharesAmountOut,
	spreadFactor sdk.Dec,
) sdk.Dec {
	// delta balanceIn is negative(tokens inside the pool increases)
	// pool weight is always 1
	tokenAmountIn := solveConstantFunctionInvariant(totalPoolSharesSupply.Add(sharesAmountOut), totalPoolSharesSupply, sdk.OneDec(), tokenBalanceIn, normalizedTokenWeightIn).Neg()
	// deduct spread factor on the in asset
	tokenAmountInFeeIncluded := tokenAmountIn.Quo(feeRatio(normalizedTokenWeightIn, spreadFactor))
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
	spreadFactor,
	exitFee sdk.Dec,
) sdk.Dec {
	tokenAmountOutFeeIncluded := tokenAmountOut.Quo(feeRatio(normalizedTokenWeightOut, spreadFactor))

	// delta poolSupply is positive(total pool shares decreases)
	// pool weight is always 1
	sharesIn := solveConstantFunctionInvariant(tokenBalanceOut.Sub(tokenAmountOutFeeIncluded), tokenBalanceOut, normalizedTokenWeightOut, totalPoolSharesSupply, sdk.OneDec())

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	sharesInFeeIncluded := sharesIn.Quo(sdk.OneDec().Sub(exitFee))
	return sharesInFeeIncluded
}

// ensureDenomInPool check to make sure the input denoms exist in the provided pool asset map
func ensureDenomInPool(poolAssetsByDenom map[string]PoolAsset, tokensIn sdk.Coins) error {
	for _, coin := range tokensIn {
		_, ok := poolAssetsByDenom[coin.Denom]
		if !ok {
			return errorsmod.Wrapf(types.ErrDenomNotFoundInPool, invalidInputDenomsErrFormat, coin.Denom)
		}
	}

	return nil
}
