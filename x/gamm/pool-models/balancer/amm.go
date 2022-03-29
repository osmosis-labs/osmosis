package balancer

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/osmomath"
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

	// y = balanceXBefore/balanceYAfter
	y := tokenBalanceFixedBefore.Quo(tokenBalanceFixedAfter)

	// amountY = balanceY * (1 - (y ^ weightRatio))
	yToWeightRatio := osmomath.Pow(y, weightRatio)
	paranthetical := sdk.OneDec().Sub(yToWeightRatio)
	amountY := tokenBalanceUnknownBefore.Mul(paranthetical)
	return amountY
}

// CalcOutAmtGivenIn calculates token to be swapped out given
// the provided amount, fee deducted, using solveConstantFunctionInvariant.
func (p Pool) CalcOutAmtGivenIn(
	ctx sdk.Context, tokensIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (
	tokenOut sdk.DecCoin, err error,
) {
	tokenIn, poolAssetIn, poolAssetOut, err := p.parsePoolAssets(tokensIn, tokenOutDenom)
	if err != nil {
		return sdk.DecCoin{}, err
	}

	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	poolTokenInBalance := poolAssetIn.Token.Amount.ToDec()
	poolPostSwapInBalance := poolTokenInBalance.Add(tokenAmountInAfterFee)

	// deduct swapfee on the in asset
	// delta balanceOut is positive(tokens inside the pool decreases)
	tokenAmountOut := solveConstantFunctionInvariant(
		poolTokenInBalance, poolPostSwapInBalance, poolAssetIn.Weight.ToDec(),
		poolAssetOut.Token.Amount.ToDec(), poolAssetOut.Weight.ToDec())
	return sdk.NewDecCoinFromDec(tokenOutDenom, tokenAmountOut), nil
}

// calcInAmtGivenOut calculates token to be provided, fee added,
// given the swapped out amount, using solveConstantFunctionInvariant.
func (p Pool) CalcInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (
	tokenIn sdk.DecCoin, err error,
) {
	tokenOut, poolAssetOut, poolAssetIn, err := p.parsePoolAssets(tokensOut, tokenInDenom)
	if err != nil {
		return sdk.DecCoin{}, err
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
	return sdk.NewDecCoinFromDec(tokenInDenom, tokenAmountInBeforeFee), nil
}

// ApplySwap.
func (p *Pool) ApplySwap(ctx sdk.Context, tokensIn sdk.Coins, tokensOut sdk.Coins) error {
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
// so spot_price = (Base_supply / Weight_base) / (Quote_supply / Weight_quote)
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

	numerator := base.Token.Amount.ToDec().Quo(base.Weight.ToDec())
	denom := quote.Token.Amount.ToDec().Quo(quote.Weight.ToDec())
	ratio := numerator.Quo(denom)

	return ratio, nil
}

// balancer notation: pAo - poolshares amount out, given single asset in
// the second argument requires the tokenWeightIn / total token weight.
func calcPoolOutGivenSingleIn(
	tokenBalanceIn,
	normalizedTokenWeightIn,
	poolShares,
	tokenAmountIn,
	swapFee sdk.Dec,
) sdk.Dec {
	// deduct swapfee on the in asset.
	// We don't charge swap fee on the token amount that we imagine as unswapped (the normalized weight).
	// So effective_swapfee = swapfee * (1 - normalized_token_weight)
	effectiveSwapFee := (sdk.OneDec().Sub(normalizedTokenWeightIn)).Mul(swapFee)
	// Apply swap fee, by multiplying tokenAmountIn by (1 - effective_swap_fee)
	tokenAmountInAfterFee := tokenAmountIn.Mul(sdk.OneDec().Sub(effectiveSwapFee))
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

// calcPoolOutGivenSingleIn - balance pAo.
func (p *Pool) singleAssetJoin(tokenIn sdk.Coin, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	tokenInPoolAsset, err := p.GetPoolAsset(tokenIn.Denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	totalWeight := p.GetTotalWeight()
	if totalWeight.IsZero() {
		return sdk.ZeroInt(), errors.New("pool misconfigured, total weight = 0")
	}
	normalizedWeight := tokenInPoolAsset.Weight.ToDec().Quo(totalWeight.ToDec())
	return calcPoolOutGivenSingleIn(
		tokenInPoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		p.GetTotalShares().ToDec(),
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
	if tokensIn.Len() == 1 {
		numShares, err = p.singleAssetJoin(tokensIn[0], swapFee)
		p.updateLiquidity(numShares, tokensIn)
		return numShares, err
	} else if tokensIn.Len() != p.NumAssets() {
		return sdk.ZeroInt(), errors.New(
			"balancer pool only supports LP'ing with one asset, or all assets in pool")
	}
	// Add all exact coins we can (no swap)
	numShares, remCoins, err := p.maximalExactRatioJoin(tokensIn)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	p.updateLiquidity(numShares, tokensIn.Sub(remCoins))
	// if there are coins that couldn't be perfectly joined, do single asset joins for each of them.
	if !remCoins.Empty() {
		for _, coin := range remCoins {
			newShares, err := p.singleAssetJoin(coin, swapFee)
			if err != nil {
				return sdk.ZeroInt(), err
			}
			p.updateLiquidity(newShares, sdk.Coins{coin})
			numShares = numShares.Add(newShares)
		}
	}
	return numShares, nil
}

func (p *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	totalShares := p.GetTotalShares()
	if exitingShares.GTE(totalShares) {
		return sdk.Coins{}, errors.New(("too many shares out"))
	}

	refundedShares := exitingShares
	if !exitFee.IsZero() {
		// exitingShares * (1 - exit fee)
		// Todo: make a -1 constant
		oneSubExitFee := sdk.OneDec().Sub(exitFee)
		refundedShares = oneSubExitFee.MulInt(exitingShares).TruncateInt()
	}

	shareOutRatio := refundedShares.ToDec().QuoInt(totalShares)
	// Make it shareOutRatio * pool LP balances
	exitedCoins = sdk.Coins{}
	balances := p.GetTotalPoolLiquidity(ctx)
	for _, asset := range balances {
		exitAmt := shareOutRatio.MulInt(asset.Amount).TruncateInt()
		if exitAmt.LTE(sdk.ZeroInt()) {
			continue
		}
		exitedCoins = exitedCoins.Add(sdk.NewCoin(asset.Denom, exitAmt))
		// update pool assets for this exit amount.
		newAmt := asset.Amount.Sub(exitAmt)
		err = p.UpdatePoolAssetBalance(sdk.NewCoin(asset.Denom, newAmt))
		if err != nil {
			return sdk.Coins{}, err
		}
	}

	p.TotalShares = sdk.NewCoin(p.TotalShares.Denom, totalShares.Sub(exitingShares))
	return exitedCoins, nil
}
