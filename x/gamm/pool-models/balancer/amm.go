package balancer

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/osmomath"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

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

// CalcOutAmtGivenIn calculates tokens to be swapped out given the provided
// amount and fee deducted, using solveConstantFunctionInvariant.
func (p Pool) CalcOutAmtGivenIn(
	ctx sdk.Context,
	tokensIn sdk.Coins,
	tokenOutDenom string,
	swapFee sdk.Dec,
) (sdk.Coin, error) {
	tokenIn, poolAssetIn, poolAssetOut, err := p.parsePoolAssets(tokensIn, tokenOutDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))
	poolTokenInBalance := poolAssetIn.Token.Amount.ToDec()
	poolPostSwapInBalance := poolTokenInBalance.Add(tokenAmountInAfterFee)

	// deduct swapfee on the tokensIn
	// delta balanceOut is positive(tokens inside the pool decreases)
	tokenAmountOut := solveConstantFunctionInvariant(
		poolTokenInBalance,
		poolPostSwapInBalance,
		poolAssetIn.Weight.ToDec(),
		poolAssetOut.Token.Amount.ToDec(),
		poolAssetOut.Weight.ToDec(),
	)

	// We ignore the decimal component, as we round down the token amount out.
	tokenAmountOutInt := tokenAmountOut.TruncateInt()
	if !tokenAmountOutInt.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}

	return sdk.NewCoin(tokenOutDenom, tokenAmountOutInt), nil
}

// SwapOutAmtGivenIn is a mutative method for CalcOutAmtGivenIn, which includes the actual swap.
func (p *Pool) SwapOutAmtGivenIn(
	ctx sdk.Context,
	tokensIn sdk.Coins,
	tokenOutDenom string,
	swapFee sdk.Dec,
) (
	tokenOut sdk.Coin, err error,
) {
	tokenOutCoin, err := p.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	err = p.applySwap(ctx, tokensIn, sdk.Coins{tokenOutCoin})
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenOutCoin, nil
}

// CalcInAmtGivenOut calculates token to be provided, fee added,
// given the swapped out amount, using solveConstantFunctionInvariant.
func (p Pool) CalcInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (
	tokenIn sdk.Coin, err error,
) {
	tokenOut, poolAssetOut, poolAssetIn, err := p.parsePoolAssets(tokensOut, tokenInDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// delta balanceOut is positive(tokens inside the pool decreases)
	poolTokenOutBalance := poolAssetOut.Token.Amount.ToDec()
	poolPostSwapOutBalance := poolTokenOutBalance.Sub(tokenOut.Amount.ToDec())
	// (x_0)(y_0) = (x_0 + in)(y_0 - out)
	tokenAmountIn := solveConstantFunctionInvariant(
		poolTokenOutBalance, poolPostSwapOutBalance, poolAssetOut.Weight.ToDec(),
		poolAssetIn.Token.Amount.ToDec(), poolAssetIn.Weight.ToDec()).Neg()

	// We deduct a swap fee on the input asset. The swap happens by following the invariant curve on the input * (1 - swap fee)
	// and then the swap fee is added to the pool.
	// Thus in order to give X amount out, we solve the invariant for the invariant input. However invariant input = (1 - swapfee) * trade input.
	// Therefore we divide by (1 - swapfee) here
	tokenAmountInBeforeFee := tokenAmountIn.Quo(sdk.OneDec().Sub(swapFee))

	// We round up tokenInAmt, as this is whats charged for the swap, for the precise amount out.
	// Otherwise, the pool would under-charge by this rounding error.
	tokenInAmt := tokenAmountInBeforeFee.Ceil().TruncateInt()

	if !tokenInAmt.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}
	return sdk.NewCoin(tokenInDenom, tokenInAmt), nil
}

// SwapInAmtGivenOut is a mutative method for CalcOutAmtGivenIn, which includes the actual swap.
func (p *Pool) SwapInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (
	tokenIn sdk.Coin, err error,
) {
	tokenInCoin, err := p.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, swapFee)
	if err != nil {
		return sdk.Coin{}, err
	}

	err = p.applySwap(ctx, sdk.Coins{tokenInCoin}, tokensOut)
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenInCoin, nil
}

// ApplySwap.
func (p *Pool) applySwap(ctx sdk.Context, tokensIn sdk.Coins, tokensOut sdk.Coins) error {
	// Also ensures that len(tokensIn) = 1 = len(tokensOut)
	inPoolAsset, outPoolAsset, err := p.parsePoolAssetsCoins(tokensIn, tokensOut)
	if err != nil {
		return err
	}
	inPoolAsset.Token.Amount = inPoolAsset.Token.Amount.Add(tokensIn[0].Amount)
	outPoolAsset.Token.Amount = outPoolAsset.Token.Amount.Sub(tokensOut[0].Amount)

	return p.UpdatePoolAssetBalances(sdk.NewCoins(
		inPoolAsset.Token,
		outPoolAsset.Token,
	))
}

// SpotPrice returns the spot price of the pool
// This is the weight-adjusted balance of the tokens in the pool.
// In order reduce the propagated effect of incorrect trailing digits,
// we take the ratio of weights and divide this by ratio of supplies
// this is equivalent to spot_price = (Base_supply / Weight_base) / (Quote_supply / Weight_quote)
// but cancels out the common term in weight.
//
// panics if pool is misconfigured and has any weight as 0.
func (p Pool) SpotPrice(ctx sdk.Context, baseAsset, quoteAsset string) (sdk.Dec, error) {
	quote, base, err := p.parsePoolAssetsByDenoms(quoteAsset, baseAsset)
	if err != nil {
		return sdk.Dec{}, err
	}
	if base.Weight.IsZero() || quote.Weight.IsZero() {
		return sdk.Dec{}, errors.New("pool is misconfigured, got 0 weight")
	}

	// spot_price = (Base_supply / Weight_base) / (Quote_supply / Weight_quote)
	// spot_price = (weight_quote / weight_base) * (base_supply / quote_supply)
	invWeightRatio := quote.Weight.ToDec().Quo(base.Weight.ToDec())
	supplyRatio := base.Token.Amount.ToDec().Quo(quote.Token.Amount.ToDec())
	fullRatio := supplyRatio.Mul(invWeightRatio)
	ratio := (fullRatio.Mul(types.SigFigs).RoundInt()).ToDec().Quo(types.SigFigs)
	return ratio, nil
}

// feeRatio returns the fee ratio that is defined as follows:
// 1 - ((1 - normalizedTokenWeightOut) * swapFee)
func feeRatio(normalizedWeight, swapFee sdk.Dec) sdk.Dec {
	return sdk.OneDec().Sub((sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee))
}

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

// calcPoolOutGivenSingleIn - balance pAo.
func (p *Pool) calcSingleAssetJoin(tokenIn sdk.Coin, swapFee sdk.Dec, tokenInPoolAsset PoolAsset, totalShares sdk.Int) (numShares sdk.Int, err error) {
	totalWeight := p.GetTotalWeight()
	if totalWeight.IsZero() {
		return sdk.ZeroInt(), errors.New("pool misconfigured, total weight = 0")
	}
	normalizedWeight := tokenInPoolAsset.Weight.ToDec().Quo(totalWeight.ToDec())
	return calcPoolSharesOutGivenSingleAssetIn(
		tokenInPoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		totalShares.ToDec(),
		tokenIn.Amount.ToDec(),
		swapFee,
	).TruncateInt(), nil
}

func (p *Pool) maximalExactRatioJoin(tokensIn sdk.Coins) (numShares sdk.Int, remCoins sdk.Coins, err error) {
	coinShareRatios := make([]sdk.Dec, len(tokensIn), len(tokensIn))
	minShareRatio := sdk.MaxSortableDec
	maxShareRatio := sdk.ZeroDec()

	poolLiquidity := p.GetTotalPoolLiquidity(sdk.Context{})

	for i, coin := range tokensIn {
		shareRatio := coin.Amount.ToDec().QuoInt(poolLiquidity.AmountOfNoDenomValidation(coin.Denom))
		if shareRatio.LT(minShareRatio) {
			minShareRatio = shareRatio
		}
		if shareRatio.GT(maxShareRatio) {
			maxShareRatio = shareRatio
		}
		coinShareRatios[i] = shareRatio
	}

	remCoins = sdk.Coins{}
	if minShareRatio.Equal(sdk.MaxSortableDec) {
		return numShares, remCoins, errors.New("unexpected error in balancer maximalExactRatioJoin")
	}
	numShares = minShareRatio.MulInt(p.TotalShares.Amount).TruncateInt()

	// if we have multiple shares, calculate remCoins
	if !minShareRatio.Equal(maxShareRatio) {
		// we have to calculate remCoins
		for i, coin := range tokensIn {
			if !coinShareRatios[i].Equal(minShareRatio) {
				usedAmount := minShareRatio.MulInt(coin.Amount).Ceil().TruncateInt()
				newAmt := coin.Amount.Sub(usedAmount)
				// add to RemCoins
				if !newAmt.IsZero() {
					remCoins = remCoins.Add(sdk.Coin{Denom: coin.Denom, Amount: newAmt})
				}
			}
		}
	}

	return numShares, remCoins, nil
}

func (p *Pool) JoinPool(_ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	numShares, newLiquidity, err := p.CalcJoinPoolShares(_ctx, tokensIn, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}
	p.updateLiquidity(numShares, newLiquidity)
	return numShares, nil
}

func (p *Pool) CalcJoinPoolShares(_ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	poolAssets := p.GetAllPoolAssets()
	poolAssetsByDenom := make(map[string]PoolAsset)
	for _, poolAsset := range poolAssets {
		poolAssetsByDenom[poolAsset.Token.Denom] = poolAsset
	}
	totalShares := p.GetTotalShares()

	if tokensIn.Len() == 1 {
		numShares, err = p.calcSingleAssetJoin(tokensIn[0], swapFee, poolAssetsByDenom[tokensIn[0].Denom], totalShares)
		newLiquidity = tokensIn
		return numShares, newLiquidity, err
	} else if tokensIn.Len() != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New(
			"balancer pool only supports LP'ing with one asset, or all assets in pool")
	}
	// Add all exact coins we can (no swap)
	numShares, remCoins, err := p.maximalExactRatioJoin(tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	// update liquidity for accurate calcSingleAssetJoin calculation
	newLiquidity = tokensIn.Sub(remCoins)
	for _, coin := range newLiquidity {
		poolAsset := poolAssetsByDenom[coin.Denom]
		poolAsset.Token.Amount = poolAssetsByDenom[coin.Denom].Token.Amount.Add(coin.Amount)
		poolAssetsByDenom[coin.Denom] = poolAsset
	}
	totalShares = totalShares.Add(numShares)

	// if there are coins that couldn't be perfectly joined, do single asset joins for each of them.
	if !remCoins.Empty() {
		for _, coin := range remCoins {
			newShares, err := p.calcSingleAssetJoin(coin, swapFee, poolAssetsByDenom[coin.Denom], totalShares)
			if err != nil {
				return sdk.ZeroInt(), sdk.NewCoins(), err
			}
			newLiquidity = newLiquidity.Add(coin)
			numShares = numShares.Add(newShares)
		}
	}
	return numShares, newLiquidity, nil
}

func (p *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	exitingCoins, err = p.CalcExitPoolShares(ctx, exitingShares, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}

	if err := p.exitPool(ctx, exitingCoins, exitingShares); err != nil {
		return sdk.Coins{}, err
	}

	return exitingCoins, nil
}

// exitPool exits the pool given exitingCoins and exitingShares.
// updates the pool's liquidity and totalShares.
func (p *Pool) exitPool(ctx sdk.Context, exitingCoins sdk.Coins, exitingShares sdk.Int) error {
	balances := p.GetTotalPoolLiquidity(ctx).Sub(exitingCoins)
	if err := p.UpdatePoolAssetBalances(balances); err != nil {
		return err
	}

	totalShares := p.GetTotalShares()
	p.TotalShares = sdk.NewCoin(p.TotalShares.Denom, totalShares.Sub(exitingShares))

	return nil
}

func (p *Pool) CalcExitPoolShares(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return cfmm_common.CalcExitPool(ctx, p, exitingShares, exitFee)
}

// balancer notation: pAi - pool shares amount in, given single asset out.
// the returned shares in have the fee included in them.
// the second argument requires the tokenWeightOut / total token weight.
func calcPoolSharesInGivenSingleAssetOut(
	tokenBalanceOut,
	normalizedTokenWeightOut,
	poolSupply,
	tokenAmountOut,
	swapFee,
	exitFee sdk.Dec,
) sdk.Dec {
	// feeRatio is defined as follows:
	// 1 - ((1 - normalizedTokenWeightOut) * swapFee)
	feeRatio := sdk.OneDec().Sub((sdk.OneDec().Sub(normalizedTokenWeightOut)).Mul(swapFee))

	tokenAmountOutBeforeFee := tokenAmountOut.Quo(feeRatio)

	// delta poolSupply is positive(total pool shares decreases)
	// pool weight is always 1
	sharesIn := solveConstantFunctionInvariant(tokenBalanceOut.Sub(tokenAmountOutBeforeFee), tokenBalanceOut, normalizedTokenWeightOut, poolSupply, sdk.OneDec())

	// charge exit fee on the pool token side
	// pAi = pAiAfterExitFee/(1-exitFee)
	sharesInFeeIncluded := sharesIn.Quo(sdk.OneDec().Sub(exitFee))
	return sharesInFeeIncluded
}

func (p *Pool) ExitSwapExactAmountOut(
	ctx sdk.Context,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	_, pAsset, err := p.getPoolAssetAndIndex(tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, err
	}

	sharesIn := calcPoolSharesInGivenSingleAssetOut(
		pAsset.Token.Amount.ToDec(),
		pAsset.Weight.ToDec().Quo(p.TotalWeight.ToDec()),
		p.GetTotalShares().ToDec(),
		tokenOut.Amount.ToDec(),
		p.GetSwapFee(ctx),
		p.GetExitFee(ctx),
	).TruncateInt()

	if sharesIn.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if sharesIn.GT(shareInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", pAsset.Token.Denom)
	}

	if err := p.exitPool(ctx, sdk.NewCoins(tokenOut), sharesIn); err != nil {
		return sdk.Int{}, err
	}

	return sharesIn, nil
}
