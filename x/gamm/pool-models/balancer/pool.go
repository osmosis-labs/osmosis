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
	errMsgFormatSharesAmountNotPositive       = "shares amount must be positive, was %d"
	errMsgFormatTokenAmountNotPositive        = "token amount must be positive, was %d"
	errMsgFormatTokensLargerThanMax           = "%d resulted tokens is larger than the max amount of %d"
	errMsgFormatSharesLargerThanMax           = "%d resulted shares is larger than the max amount of %d"
	errMsgFormatFailedInterimLiquidityUpdate  = "failed to update interim liquidity - pool asset %s does not exist"
	errMsgFormatRepeatingPoolAssetsNotAllowed = "repeating pool assets not allowed, found %s"
	v10Fork                                   = 4713065
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

// calcPoolOutGivenSingleIn - balance pAo.
func (p *Pool) calcSingleAssetJoin(tokenIn sdk.Coin, swapFee sdk.Dec, tokenInPoolAsset PoolAsset, totalShares sdk.Int) (numShares sdk.Int, err error) {
	_, err = p.GetPoolAsset(tokenIn.Denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

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

// JoinPool calculates the number of shares needed given tokensIn with swapFee applied.
// It updates the liquidity if the pool is joined successfully. If not, returns error.
// and updates pool accordingly.
func (p *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	numShares, newLiquidity, err := p.CalcJoinPoolShares(ctx, tokensIn, swapFee)
	if err != nil {
		return sdk.Int{}, err
	}

	// update pool with the calculated share and liquidity needed to join pool
	p.IncreaseLiquidity(numShares, newLiquidity)
	return numShares, nil
}

func (p *Pool) calcJoinPoolSharesBroken(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
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

	// Add all exact coins we can (no swap). ctx arg doesn't matter for Balancer.
	numShares, remCoins, err := cfmm_common.MaximalExactRatioJoinBroken(p, sdk.Context{}, tokensIn)
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

	// If there are coins that couldn't be perfectly joined, do single asset joins
	// for each of them.
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

// CalcJoinPoolShares calculates the number of shares created to join pool with the provided amount of `tokenIn`.
// The input tokens must either be:
// - a single token
// - contain exactly the same tokens as the pool contains
//
// It returns the number of shares created, the amount of coins actually joined into the pool
// (in case of not being able to fully join), or an error.
func (p *Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, tokensJoined sdk.Coins, err error) {
	if ctx.BlockHeight() < v10Fork {
		return p.calcJoinPoolSharesBroken(ctx, tokensIn, swapFee)
	}
	// 1) Get pool current liquidity + and token weights
	// 2) If single token provided, do single asset join and exit.
	// 3) If multi-asset join, first do as much of a join as we can with no swaps.
	// 4) Update pool shares / liquidity / remaining tokens to join accordingly
	// 5) For every remaining token to LP, do a single asset join, and update pool shares / liquidity.
	//
	// Note that all single asset joins do incur swap fee.
	//
	// Since CalcJoinPoolShares is non-mutative, the steps for updating pool shares / liquidity are
	// more complex / don't just alter the state.
	// We should simplify this logic further in the future, using balancer multi-join equations.

	// 1) get all 'pool assets' (aka current pool liquidity + balancer weight)
	poolAssetsByDenom, err := getPoolAssetsByDenom(p.GetAllPoolAssets())
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	totalShares := p.GetTotalShares()
	if tokensIn.Len() == 1 {
		// 2) Single token provided, so do single asset join and exit.
		numShares, err = p.calcSingleAssetJoin(tokensIn[0], swapFee, poolAssetsByDenom[tokensIn[0].Denom], totalShares)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}
		// we join all the tokens.
		tokensJoined = tokensIn
		return numShares, tokensJoined, nil
	} else if tokensIn.Len() != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("balancer pool only supports LP'ing with one asset or all assets in pool")
	}

	// 3) JoinPoolNoSwap with as many tokens as we can. (What is in perfect ratio)
	// * numShares is how many shares are perfectly matched.
	// * remainingTokensIn is how many coins we have left to join, that have not already been used.
	// if remaining coins is empty, logic is done (we joined all tokensIn)
	numShares, remainingTokensIn, err := cfmm_common.MaximalExactRatioJoin(p, sdk.Context{}, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	if remainingTokensIn.Empty() {
		tokensJoined = tokensIn
		return numShares, tokensJoined, nil
	}

	// 4) Still more coins to join, so we update the effective pool state here to account for
	// join that just happened.
	// * We add the joined coins to our "current pool liquidity" object (poolAssetsByDenom)
	// * We increment a variable for our "newTotalShares" to add in the shares that've been added.
	tokensJoined = tokensIn.Sub(remainingTokensIn)
	if err := updateIntermediaryPoolAssetsLiquidity(tokensJoined, poolAssetsByDenom); err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	newTotalShares := totalShares.Add(numShares)

	// 5) Now single asset join each remaining coin.
	newNumSharesFromRemaining, newLiquidityFromRemaining, err := p.calcJoinSingleAssetTokensIn(remainingTokensIn, newTotalShares, poolAssetsByDenom, swapFee)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	// update total amount LP'd variable, and total new LP shares variable, run safety check, and return
	numShares = numShares.Add(newNumSharesFromRemaining)
	tokensJoined = tokensJoined.Add(newLiquidityFromRemaining...)

	if tokensJoined.IsAnyGT(tokensIn) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("An error has occurred, more coins joined than token In")
	}

	return numShares, tokensJoined, nil
}

// calcJoinSingleAssetTokensIn attempts to calculate single
// asset join for all tokensIn given totalShares in pool,
// poolAssetsByDenom and swapFee. totalShares is the number
// of shares in pool before beginnning to join any of the tokensIn.
//
// Returns totalNewShares and totalNewLiquidity from joining all tokensIn
// by mimicking individually single asset joining each.
// or error if fails to calculate join for any of the tokensIn.
func (p *Pool) calcJoinSingleAssetTokensIn(tokensIn sdk.Coins, totalShares sdk.Int, poolAssetsByDenom map[string]PoolAsset, swapFee sdk.Dec) (sdk.Int, sdk.Coins, error) {
	totalNewShares := sdk.ZeroInt()
	totalNewLiquidity := sdk.NewCoins()
	for _, coin := range tokensIn {
		newShares, err := p.calcSingleAssetJoin(coin, swapFee, poolAssetsByDenom[coin.Denom], totalShares.Add(totalNewShares))
		if err != nil {
			return sdk.ZeroInt(), sdk.Coins{}, err
		}

		totalNewLiquidity = totalNewLiquidity.Add(coin)
		totalNewShares = totalNewShares.Add(newShares)
	}
	return totalNewShares, totalNewLiquidity, nil
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
