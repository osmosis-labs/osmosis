package balancer

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/osmomath"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	errMsgFormatSharesAmountNotPositive = "shares amount must be positive, was %d"
	errMsgFormatTokenAmountNotPositive  = "token amount must be positive, was %d"
	errMsgFormatTokensLargerThanMax     = "%d resulted tokens is larger than the max amount of %d"
	errMsgFormatSharesLargerThanMax     = "%d resulted shares is larger than the max amount of %d"
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
	// we want to round this to `SigFigs` of precision
	ratio := osmomath.SigFigRound(fullRatio, types.SigFigs)
	return ratio, nil
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

// JoinPool calculates the number of shares needed and the updated liquidity for joining pool
// and updates pool accordingly.
func (p *Pool) JoinPool(_ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	numShares, newLiquidity, err := p.CalcJoinPoolShares(_ctx, tokensIn, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	// update pool with the calculated share and liquidity needed to join pool
	p.IncreaseLiquidity(numShares, newLiquidity)
	return numShares, nil
}

// CalcJoinPoolShares calculate the number of shares created to join pool with the provieded amount of `tokenIn`.
// When a single token is provided as an argument, we simply perform single asset join to the token.
// If not tokenIn provided as an argument isn't a sinlge token, it must contain all the tokens in the pool.
// For the case of multi-asset join for a pool, we first calculate the maximum amount we can join a pool without swap, then
// perform single asset join for the remaining coins.
// CalcJoinPoolShares does not directly alter the state of the pool, but only does the calculation for shares for joining the pool.
func (p *Pool) CalcJoinPoolShares(_ sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	poolAssets := p.GetAllPoolAssets()
	poolAssetsByDenom := make(map[string]PoolAsset)
	for _, poolAsset := range poolAssets {
		poolAssetsByDenom[poolAsset.Token.Denom] = poolAsset
	}

	totalShares := p.GetTotalShares()

	if tokensIn.Len() == 1 {
		numShares, err = p.calcSingleAssetJoin(tokensIn[0], swapFee, poolAssetsByDenom[tokensIn[0].Denom], totalShares)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}

		newLiquidity = tokensIn

		return numShares, newLiquidity, nil
	} else if tokensIn.Len() != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("balancer pool only supports LP'ing with one asset or all assets in pool")
	}

	// Add all exact coins we can join pool with without swap. ctx arg doesn't matter for Balancer.
	// calculate the number of shares we can join pool with without swap, and the remaining tokens
	// that has to be joined via single asset join
	// ctx arg doesn't matter for balancer
	numShares, remCoins, err := cfmm_common.MaximalExactRatioJoin(p, sdk.Context{}, tokensIn)
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

	newTotalShares := totalShares.Add(numShares)

	// If there are coins that couldn't be perfectly joined, do single asset joins
	// for each of them.
	if !remCoins.Empty() {
		for _, coin := range remCoins {
			newShares, err := p.calcSingleAssetJoin(coin, swapFee, poolAssetsByDenom[coin.Denom], newTotalShares)
			if err != nil {
				return sdk.ZeroInt(), sdk.NewCoins(), err
			}

			newLiquidity = newLiquidity.Add(coin)
			newTotalShares = newTotalShares.Add(newShares)
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

func (p *Pool) CalcTokenInShareAmountOut(
	ctx sdk.Context,
	tokenInDenom string,
	shareOutAmount sdk.Int,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	_, poolAssetIn, err := p.getPoolAssetAndIndex(tokenInDenom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := poolAssetIn.Weight.ToDec().Quo(p.GetTotalWeight().ToDec())

	// We round up tokenInAmount, as this is whats charged for the swap, for the precise amount out.
	// Otherwise, the pool would under-charge by this rounding error.
	tokenInAmount = calcSingleAssetInGivenPoolSharesOut(
		poolAssetIn.Token.Amount.ToDec(),
		normalizedWeight,
		p.GetTotalShares().ToDec(),
		shareOutAmount.ToDec(),
		swapFee,
	).Ceil().TruncateInt()

	if !tokenInAmount.IsPositive() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, errMsgFormatTokenAmountNotPositive, tokenInAmount.Int64())
	}

	return tokenInAmount, nil
}

func (p *Pool) JoinPoolTokenInMaxShareAmountOut(
	ctx sdk.Context,
	tokenInDenom string,
	shareOutAmount sdk.Int,
) (tokenInAmount sdk.Int, err error) {
	_, poolAssetIn, err := p.getPoolAssetAndIndex(tokenInDenom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := poolAssetIn.Weight.ToDec().Quo(p.GetTotalWeight().ToDec())

	tokenInAmount = calcSingleAssetInGivenPoolSharesOut(
		poolAssetIn.Token.Amount.ToDec(),
		normalizedWeight,
		p.GetTotalShares().ToDec(),
		shareOutAmount.ToDec(),
		p.GetSwapFee(ctx),
	).TruncateInt()

	if !tokenInAmount.IsPositive() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, errMsgFormatTokenAmountNotPositive, tokenInAmount.Int64())
	}

	poolAssetIn.Token.Amount = poolAssetIn.Token.Amount.Add(tokenInAmount)
	err = p.UpdatePoolAssetBalance(poolAssetIn.Token)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenInAmount, nil
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

func (p *Pool) ExitSwapExactAmountOut(
	ctx sdk.Context,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	_, poolAssetOut, err := p.getPoolAssetAndIndex(tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, err
	}

	sharesIn := calcPoolSharesInGivenSingleAssetOut(
		poolAssetOut.Token.Amount.ToDec(),
		poolAssetOut.Weight.ToDec().Quo(p.TotalWeight.ToDec()),
		p.GetTotalShares().ToDec(),
		tokenOut.Amount.ToDec(),
		p.GetSwapFee(ctx),
		p.GetExitFee(ctx),
	).TruncateInt()

	if !sharesIn.IsPositive() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, errMsgFormatSharesAmountNotPositive, sharesIn.Int64())
	}

	if sharesIn.GT(shareInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, errMsgFormatSharesLargerThanMax, sharesIn.Int64(), shareInMaxAmount.Uint64())
	}

	if err := p.exitPool(ctx, sdk.NewCoins(tokenOut), sharesIn); err != nil {
		return sdk.Int{}, err
	}

	return sharesIn, nil
}
