package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// var (
// 	_ types.PoolI = &Pool{}
// )

func NewConcentratedLiquidityPool(poolId uint64, denoms []string) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)

	// pool thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	pool := &Pool{
		Address: poolAddr.String(),
		Id:      poolId,
	}

	err := pool.SetInitialPoolAssets(denoms)
	if err != nil {
		return Pool{}, err
	}

	return *pool, nil
}

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// pa is the smaller of sqrtpCur and the nextPrice
// pb is the larger of sqrtpCur and the nextPrice
func liquidity0(amount int64, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	product := pa.Mul(pb)
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
	return amt.Quo(diff)
}

// calcAmount0 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// pa is the smaller of sqrtpCur and the nextPrice
// pb is the larger of sqrtpCur and the nextPrice
func calcAmount0(liq, pa, pb sdk.Dec) sdk.Dec {
	if pa.GT(pb) {
		pa, pb = pb, pa
	}
	diff := pb.Sub(pa)
	mult := liq
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
	return liq.Mul(diff)
}

// TODO: remove nolint
// nolint: unused
func priceToTick(price sdk.Dec) sdk.Int {
	logOfPrice := osmomath.BigDecFromSDKDec(price).ApproxLog2()
	logInt := osmomath.NewDecWithPrec(10001, 4)
	tick := logOfPrice.Quo(logInt.ApproxLog2())
	return tick.SDKDec().TruncateInt()
}

// TODO: remove nolint
// nolint: unused
func tickToPrice(tick sdk.Int) sdk.Dec {
	price := sdk.NewDecWithPrec(10001, 4).Power(tick.Uint64())
	return price
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
func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	asset0 := p.Token0
	asset1 := p.Token1
	if tokenIn.Denom != asset0 && tokenIn.Denom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
	}
	if tokenOutDenom != asset0 && tokenOutDenom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenOutDenom does not match any asset in pool")
	}
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Coin{}, fmt.Errorf("tokenIn cannot be the same as tokenOut")
	}
	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(swapFee))

	// TODO: Replace with spot price
	priceLower := sdk.NewDec(4500)
	priceCur := sdk.NewDec(5000)
	priceUpper := sdk.NewDec(5500)

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
		priceNextTop := liq.Mul(sqrtPCurTick)
		priceNextBot := liq.Add(tokenAmountInAfterFee.Mul(sqrtPCurTick))
		priceNext := priceNextTop.Quo(priceNextBot)
		// new amount in, will be needed later
		//amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountOut := calcAmount1(liq, priceNext, sqrtPCurTick)
		return sdk.NewCoin(tokenOutDenom, amountOut.TruncateInt()), nil
	}
	return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenOutAmt := tokenOut.Amount.ToDec()
	asset0 := p.Token0
	asset1 := p.Token1
	if tokenOut.Denom != asset0 && tokenOut.Denom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
	}
	if tokenInDenom != asset0 && tokenInDenom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenOutDenom does not match any asset in pool")
	}
	if tokenOut.Denom == tokenInDenom {
		return sdk.Coin{}, fmt.Errorf("tokenIn cannot be the same as tokenOut")
	}

	// TODO: Replace with spot price
	priceLower := sdk.NewDec(4500)
	priceCur := sdk.NewDec(5000)
	priceUpper := sdk.NewDec(5500)

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
		// amountOut = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn := calcAmount0(liq, priceNext, sqrtPCurTick)
		// fee logic
		amountIn = amountIn.Quo(sdk.OneDec().Sub(swapFee))
		return sdk.NewCoin(tokenInDenom, amountIn.TruncateInt()), nil
	} else if tokenOut.Denom == asset0 {
		priceNextTop := liq.Mul(sqrtPCurTick)
		priceNextBot := liq.Add(tokenOutAmt.Mul(sqrtPCurTick))
		priceNext := priceNextTop.Quo(priceNextBot)

		// new amount in, will be needed later
		// amountOut = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn := calcAmount1(liq, priceNext, sqrtPCurTick)
		// fee logic
		amountIn = amountIn.Quo(sdk.OneDec().Sub(swapFee))
		return sdk.NewCoin(tokenInDenom, amountIn.TruncateInt()), nil
	}
	return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
}

func (p *Pool) SetInitialPoolAssets(poolDenoms []string) error {
	if len(poolDenoms) != 2 {
		return fmt.Errorf("pool must contain exactly two assets")
	}
	if poolDenoms[0] == poolDenoms[1] {
		return fmt.Errorf("cannot have same asset in pool")
	}
	if poolDenoms[0] > poolDenoms[1] {
		poolDenoms[1] = poolDenoms[0]
	}

	p.Token0 = poolDenoms[0]
	p.Token1 = poolDenoms[1]

	return nil
}
