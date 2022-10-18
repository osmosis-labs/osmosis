package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	_ types.PoolI = &Pool{}
)

func NewConcentratedLiquidityPool(poolId uint64) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)

	// pool thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	pool := &Pool{
		Address: poolAddr.String(),
		Id:      poolId,
	}

	return *pool, nil
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
	mult := liq.Mul(sdk.NewDec(10).Power(6))
	return (mult.Mul(diff)).Quo(pb).Quo(pa)
}

func CalcAmount1(liq, pa, pb sdk.Dec) sdk.Dec {
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

// this only works on a single directional trade, will implement bi directional trade in next milestone
func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	tokenAmountInAfterFee := tokenIn[0].Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	// TODO: Replace with spot price
	priceLower := sdk.NewDec(4500)
	priceCur := sdk.NewDec(5000)
	priceUpper := sdk.NewDec(5500)

	sqrtpLow, _ := priceLower.ApproxSqrt()
	sqrtpLow = sqrtpLow.Mul(sdk.NewDec(10).Power(6))
	sqrtpCur, _ := priceCur.ApproxSqrt()
	sqrtpCur = sqrtpCur.Mul(sdk.NewDec(10).Power(6))
	sqrtpUpp, _ := priceUpper.ApproxSqrt()
	sqrtpUpp = sqrtpUpp.Mul(sdk.NewDec(10).Power(6))

	// TODO: Roman change out with query to pool to get this info
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

	// new amount in, will be needed later
	//amountIn = CalcAmount1(liq, priceNext, sqrtpCur)
	amountOut := CalcAmount0(liq, priceNext, sqrtpCur)
	coinOut := sdk.NewCoin(tokenOutDenom, amountOut.TruncateInt())
	return coinOut, nil
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenAmountInBeforeFee := tokenOut[0].Amount.ToDec().Quo(sdk.OneDec().Sub(swapFee))

	// TODO: Replace with spot price
	priceLower := sdk.NewDec(4500)
	priceCur := sdk.NewDec(5000)
	priceUpper := sdk.NewDec(5500)

	sqrtpLow, _ := priceLower.ApproxSqrt()
	sqrtpLow = sqrtpLow.Mul(sdk.NewDec(10).Power(6))
	sqrtpCur, _ := priceCur.ApproxSqrt()
	sqrtpCur = sqrtpCur.Mul(sdk.NewDec(10).Power(6))
	sqrtpUpp, _ := priceUpper.ApproxSqrt()
	sqrtpUpp = sqrtpUpp.Mul(sdk.NewDec(10).Power(6))

	// TODO: Roman change out with query to pool to get this info
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

	priceDiff := tokenAmountInBeforeFee.QuoTruncateMut(liq).Mul(sdk.NewDec(10).Power(6))
	priceNext := sqrtpCur.Add(priceDiff)

	// new amount in, will be needed later
	//amountIn := CalcAmount1(liq, priceNext, sqrtpCur)
	amountIn := CalcAmount0(liq, priceNext, sqrtpCur)
	coinIn := sdk.NewCoin(tokenInDenom, amountIn.TruncateInt())
	return coinIn, nil
}

func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}
