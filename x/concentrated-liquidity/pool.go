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
	price_cur := sdk.NewDec(5000)
	price_upp := sdk.NewDec(5500)

	sqrtp_low, _ := priceLower.ApproxSqrt()
	sqrtp_low = sqrtp_low.Mul(sdk.NewDec(10).Power(6))
	sqrtp_cur, _ := price_cur.ApproxSqrt()
	sqrtp_cur = sqrtp_cur.Mul(sdk.NewDec(10).Power(6))
	sqrtp_upp, _ := price_upp.ApproxSqrt()
	sqrtp_upp = sqrtp_upp.Mul(sdk.NewDec(10).Power(6))

	amountETH := int64(1000000)
	amountUSDC := int64(5000000000)

	liq0 := Liquidity0(amountETH, sqrtp_cur, sqrtp_upp)
	fmt.Printf("liq0 %v \n", liq0)
	liq1 := Liquidity1(amountUSDC, sqrtp_cur, sqrtp_low)
	fmt.Printf("liq1 %v \n", liq1)
	var liq sdk.Dec
	if liq0.LT(liq1) {
		liq = liq0
	} else {
		liq = liq1
	}

	price_diff := tokenAmountInAfterFee.QuoTruncateMut(liq).Mul(sdk.NewDec(10).Power(6))
	fmt.Printf("price_diff %v \n", price_diff)

	price_next := sqrtp_cur.Add(price_diff)

	amountOut := CalcAmount0(liq, price_next, sqrtp_cur)
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
