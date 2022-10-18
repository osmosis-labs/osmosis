package concentrated_liquidity

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func NewConcentratedPool(poolId uint64, firstDenom string, secondDenom string) (types.Pool, error) {
	denom0 := firstDenom
	denom1 := secondDenom

	// we store token in lexiographical order
	if denom0 < denom1 {
		denom0, denom1 = secondDenom, firstDenom
	}
	pool := types.Pool{
		// TODO: implement pool address method
		// Address: types.NewPoolAddress(poolId).String(),
		Id: poolId,
	}
	return pool, nil
}

func Liquidity0(amount int64, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	product := pa.Mul(pb).Quo(sdk.NewDec(10).Power(6))
	diff := pb.Sub(pa)
	amt := sdk.NewDec(amount)
	return amt.Mul(product.Quo(diff))
}

func Liquidity1(amount int64, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	amt := sdk.NewDec(amount)
	return amt.Quo(diff).Mul(sdk.NewDec(10).Power(6))
}

func CalcAmount0(liq, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	quotient := pb.Quo(pa)
	mult := liq.Mul(sdk.NewDec(10).Power(6))
	return (mult.Mul(diff)).Quo(quotient)
}

func CalcAmount1(liq, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	return liq.Mul(diff).Quo(sdk.NewDec(10).Power(6))
}

func (k Keeper) JoinPool(ctx sdk.Context, tokenIn sdk.Coin, lowerTick sdk.Int, upperTick sdk.Int) (numShares sdk.Int, err error) {
	// first check and validate arguments
	if lowerTick.GTE(types.MaxTick) || lowerTick.LT(types.MinTick) || upperTick.GT(types.MaxTick) {
		// TODO: come back to errors
		return sdk.Int{}, fmt.Errorf("validation fail")
	}

	if tokenIn.Amount.IsZero() {
		return sdk.Int{}, fmt.Errorf("token in amount is zero")
	}

	// update tick with new liquidity
	// k.

	return sdk.Int{}, nil
}

func (k Keeper) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coin, priceUpper, priceLower sdk.Dec, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	//amount_in := sdk.NewDec(42000000)
	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	//price_low := sdk.NewDec(4545)
	priceCur := sdk.NewDec(5000)
	//price_upp := sdk.NewDec(5500)

	sqrtpLow, _ := priceLower.ApproxSqrt()
	sqrtpLow = sqrtpLow.Mul(sdk.NewDec(10).Power(6))
	sqrtpCur, _ := priceCur.ApproxSqrt()
	sqrtpCur = sqrtpCur.Mul(sdk.NewDec(10).Power(6))
	sqrtpUpp, _ := priceUpper.ApproxSqrt()
	sqrtpUpp = sqrtpUpp.Mul(sdk.NewDec(10).Power(6))

	amountETH := int64(1000000)
	amountUSDC := int64(5000000000)

	liq0 := Liquidity0(amountETH, sqrtpCur, sqrtpUpp)
	liq1 := Liquidity1(amountUSDC, sqrtpCur, sqrtpLow)

	var liq sdk.Dec
	if liq0.LT(liq1) {
		liq = liq0
	} else {
		liq = liq1
	}

	priceDiff := tokenAmountInAfterFee.QuoTruncateMut(liq).Mul(sdk.NewDec(10).Power(6))
	priceNext := sqrtpCur.Add(priceDiff)

	amountOut := CalcAmount0(liq, priceNext, sqrtpCur)
	fmt.Printf("amount_out %v \n", amountOut.Quo(sdk.NewDec(10).Power(6)))
	return sdk.Coin{}, nil
}
func (k Keeper) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}
func (k Keeper) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}

func (k Keeper) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, err error) {
	return sdk.Int{}, nil
}
func (k Keeper) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.Int{}, sdk.Coins{}, nil
}
func (k Keeper) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.Int{}, sdk.Coins{}, nil
}
func (k Keeper) ExitPool(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, nil
}

func (k Keeper) CalcExitPoolCoinsFromShares(ctx sdk.Context, numShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, nil
}
