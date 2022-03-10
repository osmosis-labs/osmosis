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

// CalcOutAmtGivenIn calculates token to be swapped out given
// the provided amount, fee deducted, using solveConstantFunctionInvariant
func (p Pool) CalcOutAmtGivenIn(
	ctx sdk.Context, tokensIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (
	tokenOut sdk.DecCoin, err error) {
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
// given the swapped out amount, using solveConstantFunctionInvariant
func (p Pool) CalcInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (
	tokenIn sdk.DecCoin, err error) {
	tokenOut, poolAssetOut, poolAssetIn, err := p.parsePoolAssets(tokensOut, tokenInDenom)
	if err != nil {
		return sdk.DecCoin{}, err
	}

	// delta balanceOut is positive(tokens inside the pool decreases)
	poolTokenOutBalance := poolAssetOut.Token.Amount.ToDec()
	poolPreSwapOutBalance := poolTokenOutBalance.Sub(tokenOut.Amount.ToDec())
	tokenAmountIn := solveConstantFunctionInvariant(
		poolTokenOutBalance, poolPreSwapOutBalance, poolAssetOut.Weight.ToDec(),
		poolAssetIn.Token.Amount.ToDec(), poolAssetIn.Weight.ToDec())

	// We deduct a swap fee on the input asset. The swap happens by following the invariant curve on the input * (1 - swap fee)
	//  and then the swap fee is added to the pool.
	// Thus in order to give X amount out, we solve the invariant for the invariant input. However invariant input = (1 - swapfee) * trade input.
	// Therefore we divide by (1 - swapfee) here
	tokenAmountInBeforeFee := tokenAmountIn.Mul(sdk.OneDec().Sub(swapFee))
	return sdk.NewDecCoinFromDec(tokenInDenom, tokenAmountInBeforeFee), nil
}

// ApplySwap
func (p *Pool) ApplySwap(ctx sdk.Context, tokensIn sdk.Coins, tokensOut sdk.Coins) error {
	// Also ensures that len(tokensIn) = 0 = len(tokensOut)
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
func (p Pool) SpotPrice(ctx sdk.Context, quoteAsset string, baseAsset string) (sdk.Dec, error) {
	quote, base, err := p.parsePoolAssetsByDenoms(quoteAsset, baseAsset)
	if err != nil {
		return sdk.Dec{}, err
	}

	numerator := base.Token.Amount.ToDec().QuoInt(base.Weight)
	denom := quote.Token.Amount.ToDec().QuoInt(quote.Weight)
	ratio := numerator.Quo(denom)

	return ratio, nil
}

func feeRatio(
	normalizedWeight,
	swapFee sdk.Dec,
) sdk.Dec {
	zar := (sdk.OneDec().Sub(normalizedWeight)).Mul(swapFee)
	return sdk.OneDec().Sub(zar)
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

func (p *Pool) singleAssetJoin() {}
func (p *Pool) exactRatioJoin()  {}

func (p *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	return sdk.ZeroInt(), errors.New("TODO: Implement")
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

	shareOutRatio := totalShares.ToDec().QuoInt(refundedShares)
	// Make it shareOutRatio * pool LP balances
	exitedCoins = sdk.Coins{}
	balances := p.GetTotalLpBalances(ctx)
	for _, asset := range balances {
		exitAmt := shareOutRatio.MulInt(asset.Amount).TruncateInt()
		if exitAmt.LTE(sdk.ZeroInt()) {
			continue
		}
		exitedCoins = exitedCoins.Add(sdk.NewCoin(asset.Denom, exitAmt))
		// update pool assets for this exit amount.
		newAmt := asset.Amount.Sub(exitAmt)
		p.UpdatePoolAssetBalance(sdk.NewCoin(asset.Denom, newAmt))
	}

	p.TotalShares = sdk.NewCoin(p.TotalShares.Denom, totalShares.Sub(exitingShares))
	return exitedCoins, nil
}
