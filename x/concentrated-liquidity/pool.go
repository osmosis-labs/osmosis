package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	_ types.PoolI = &Pool{}
)

func NewConcentratedLiquidityPool(poolId uint64) Pool {
	poolAddr := types.NewPoolAddress(poolId)

	// pool thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	pool := &Pool{
		Address: poolAddr.String(),
		Id:      poolId,
	}

	return *pool
}

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// pa is the smaller of sqrtpCur and the nextPrice
// pb is the larger of sqrtpCur and the nextPrice
func liquidity0(amount int64, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	product := pa.Mul(pb).Quo(sdk.NewDec(10).Power(6))
	diff := pb.Sub(pa)
	amt := sdk.NewDec(amount)
	return amt.Mul(product.Quo(diff))
}

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// pa is the smaller of sqrtpCur and the nextPrice
// pb is the larger of sqrtpCur and the nextPrice
func liquidity1(amount int64, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	amt := sdk.NewDec(amount)
	return amt.Quo(diff).Mul(sdk.NewDec(10).Power(6))
}

// calcAmount0 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// pa is the smaller of sqrtpCur and the nextPrice
// pb is the larger of sqrtpCur and the nextPrice
func calcAmount0(liq, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	mult := liq.Mul(sdk.NewDec(10).Power(6))
	return (mult.Mul(diff)).Quo(pb).Quo(pa)
}

// calcAmount1 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// pa is the smaller of sqrtpCur and the nextPrice
// pb is the larger of sqrtpCur and the nextPrice
func calcAmount1(liq, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	return liq.Mul(diff).Quo(sdk.NewDec(10).Power(6))
}

func (p Pool) GetAddress() sdk.AccAddress {
	return sdk.AccAddress{}
}

func (p Pool) String() string {
	return ""
}

func (p Pool) GetId() uint64 {
	return 0
}
func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}
func (p Pool) GetTotalShares() sdk.Int {
	return sdk.Int{}
}

func (p Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}

// this only works on a single directional trade, will implement bi directional trade in next milestone
func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokensIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	tokenIn := tokensIn[0]
	asset0 := "eth"
	asset1 := "usdc"
	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	// TODO: Replace with spot price
	priceLower := sdk.NewDec(4500000000)
	priceCur := sdk.NewDec(5000000000)
	priceUpper := sdk.NewDec(5500000000)

	sqrtPLowerTick, _ := priceLower.ApproxSqrt()
	sqrtPCurTick, _ := priceCur.ApproxSqrt()
	sqrtPUpperTick, _ := priceUpper.ApproxSqrt()

	// TODO: Roman change out with query to pool to get this info
	amountETH := int64(1000000)
	amountUSDC := int64(5000000000)

	liq0 := liquidity0(amountETH, sqrtPCurTick, sqrtPUpperTick)
	liq1 := liquidity1(amountUSDC, sqrtPCurTick, sqrtPLowerTick)

	liq := sdk.MinDec(liq0, liq1)
	if tokenIn.Denom == asset1 {
		priceDiff := tokenAmountInAfterFee.Quo(liq)
		priceNext := sqrtPCurTick.Add(priceDiff)
		// new amount in, will be needed later
		//amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountOut := calcAmount0(liq, priceNext, sqrtPCurTick)
		return sdk.NewCoin(tokenOutDenom, amountOut.TruncateInt()), nil
	} else if tokenIn.Denom == asset0 {
		priceNextTop := liq.Mul(sdk.NewDec(10).Power(6).Mul(sqrtPCurTick))
		priceNextBot := liq.Mul(sdk.NewDec(10).Power(6)).Add(tokenAmountInAfterFee.Mul(sqrtPCurTick))
		priceNext := priceNextTop.Quo(priceNextBot)
		// new amount in, will be needed later
		//amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountOut := calcAmount1(liq, priceNext, sqrtPCurTick)
		amountOut = amountOut.Mul(sdk.NewDec(10).Power(6))
		return sdk.NewCoin(tokenOutDenom, amountOut.TruncateInt()), nil
	}
	return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenOut := tokensOut[0]
	tokenOutAmt := tokenOut.Amount.ToDec()
	asset0 := "eth"
	asset1 := "usdc"

	// TODO: Replace with spot price
	priceLower := sdk.NewDec(4500000000)
	priceCur := sdk.NewDec(5000000000)
	priceUpper := sdk.NewDec(5500000000)

	sqrtPLowerTick, _ := priceLower.ApproxSqrt()
	sqrtPCurTick, _ := priceCur.ApproxSqrt()
	sqrtPUpperTick, _ := priceUpper.ApproxSqrt()

	// TODO: Roman change out with query to pool to get this info
	amountETH := int64(1000000)
	amountUSDC := int64(5000000000)

	liq0 := liquidity0(amountETH, sqrtPCurTick, sqrtPUpperTick)
	liq1 := liquidity1(amountUSDC, sqrtPCurTick, sqrtPLowerTick)

	liq := sdk.MinDec(liq0, liq1)

	if tokenOut.Denom == asset1 {
		priceDiff := tokenOutAmt.Quo(liq)
		priceNext := sqrtPCurTick.Add(priceDiff)

		// new amount in, will be needed later
		// amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn := calcAmount0(liq, priceNext, sqrtPCurTick)
		// fee logic
		amountIn = amountIn.Quo(sdk.OneDec().Sub(swapFee))
		return sdk.NewCoin(tokenInDenom, amountIn.TruncateInt()), nil
	} else if tokenOut.Denom == asset0 {
		priceNextTop := liq.Mul(sdk.NewDec(10).Power(6).Mul(sqrtPCurTick))
		priceNextBot := liq.Mul(sdk.NewDec(10).Power(6)).Add(tokenOutAmt.Mul(sqrtPCurTick))
		priceNext := priceNextTop.Quo(priceNextBot)

		// new amount in, will be needed later
		// amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn := calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn = amountIn.Mul(sdk.NewDec(10).Power(6))
		// fee logic
		amountIn = amountIn.Quo(sdk.OneDec().Sub(swapFee))
		return sdk.NewCoin(tokenInDenom, amountIn.TruncateInt()), nil
	}
	return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
}
