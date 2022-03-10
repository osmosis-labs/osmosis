package balancer

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/osmomath"
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

func (p Pool) parsePoolAssetsByDenoms(tokenADenom, tokenBDenom string) (
	Aasset types.PoolAsset, Basset types.PoolAsset, err error) {
	Aasset, found1 := types.GetPoolAssetByDenom(p.PoolAssets, tokenADenom)
	Basset, found2 := types.GetPoolAssetByDenom(p.PoolAssets, tokenBDenom)
	if !(found1 && found2) {
		return Aasset, Basset, errors.New("TODO: fill message here")
	}
	return Aasset, Basset, nil
}

func (p Pool) parsePoolAssets(tokensA sdk.Coins, tokenBDenom string) (
	tokenA sdk.Coin, Aasset types.PoolAsset, Basset types.PoolAsset, err error) {
	if len(tokensA) != 1 {
		return tokenA, Aasset, Basset, errors.New("TODO: Fill message here")
	}
	Aasset, Basset, err = p.parsePoolAssetsByDenoms(tokensA[0].Denom, tokenBDenom)
	return tokensA[0], Aasset, Basset, err
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

// TODO: Copy paste correct code here
func (p Pool) SpotPrice(ctx sdk.Context, quoteAsset string, baseAsset string) (sdk.Dec, error) {
	return sdk.ZeroDec(), nil
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

func (p *Pool) ExitPool(ctx sdk.Context, numShares sdk.Int) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, errors.New("TODO: Implement")
}
