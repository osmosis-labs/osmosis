package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func (k Keeper) CreateNewConcentratedLiquidityPool(ctx sdk.Context, poolId uint64, denom0, denom1 string, currSqrtPrice, currTick sdk.Int) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)
	pool := Pool{
		Address:          poolAddr.String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
	}

	k.setPoolById(ctx, poolId, pool)
	err := pool.orderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return Pool{}, err
	}

	return pool, nil
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
		return sdk.Coin{}, fmt.Errorf("tokenIn (%s) does not match any asset in pool", tokenIn.Denom)
	}
	if tokenOutDenom != asset0 && tokenOutDenom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenOutDenom (%s) does not match any asset in pool", tokenOutDenom)
	}
	if tokenIn.Denom == tokenOutDenom {
		return sdk.Coin{}, fmt.Errorf("tokenIn (%s) cannot be the same as tokenOut (%s)", tokenIn.Denom, tokenOutDenom)
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
		amountOut := calcAmount0Delta(liq, priceNext, sqrtPCurTick)
		return sdk.NewCoin(tokenOutDenom, amountOut.TruncateInt()), nil
	} else if tokenIn.Denom == asset0 {
		priceNextTop := liq.Mul(sqrtPCurTick)
		priceNextBot := liq.Add(tokenAmountInAfterFee.Mul(sqrtPCurTick))
		priceNext := priceNextTop.Quo(priceNextBot)
		// new amount in, will be needed later
		//amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountOut := calcAmount1Delta(liq, priceNext, sqrtPCurTick)
		return sdk.NewCoin(tokenOutDenom, amountOut.TruncateInt()), nil
	}
	return sdk.Coin{}, fmt.Errorf("tokenIn (%s) does not match any asset in pool", tokenIn.Denom)
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	tokenOutAmt := tokenOut.Amount.ToDec()
	asset0 := p.Token0
	asset1 := p.Token1
	if tokenOut.Denom != asset0 && tokenOut.Denom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenOut denom (%s) does not match any asset in pool", tokenOut.Denom)
	}
	if tokenInDenom != asset0 && tokenInDenom != asset1 {
		return sdk.Coin{}, fmt.Errorf("tokenInDenom (%s) does not match any asset in pool", tokenInDenom)
	}
	if tokenOut.Denom == tokenInDenom {
		return sdk.Coin{}, fmt.Errorf("tokenOut (%s) cannot be the same as tokenIn (%s)", tokenOut.Denom, tokenInDenom)
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
		// amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn := calcAmount0Delta(liq, priceNext, sqrtPCurTick)
		// fee logic
		amountIn = amountIn.Quo(sdk.OneDec().Sub(swapFee))
		return sdk.NewCoin(tokenInDenom, amountIn.RoundInt()), nil
	} else if tokenOut.Denom == asset0 {
		priceNextTop := liq.Mul(sqrtPCurTick)
		priceNextBot := liq.Add(tokenOutAmt.Mul(sqrtPCurTick))
		priceNext := priceNextTop.Quo(priceNextBot)

		// new amount in, will be needed later
		// amountIn = calcAmount1(liq, priceNext, sqrtPCurTick)
		amountIn := calcAmount1Delta(liq, priceNext, sqrtPCurTick)
		// fee logic
		amountIn = amountIn.Quo(sdk.OneDec().Sub(swapFee))
		return sdk.NewCoin(tokenInDenom, amountIn.RoundInt()), nil
	}
	return sdk.Coin{}, fmt.Errorf("tokenOut (%s) does not match any asset in pool", tokenOut.Denom)
}

func (p *Pool) orderInitialPoolDenoms(denom0, denom1 string) error {
	if denom0 == denom1 {
		return fmt.Errorf("cannot have the same asset in a single pool")
	}
	if denom0 > denom1 {
		denom1, denom0 = denom0, denom1
	}

	p.Token0 = denom0
	p.Token1 = denom1

	return nil
}
