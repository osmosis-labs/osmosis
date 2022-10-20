package concentrated_liquidity

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	_ types.PoolI = &Pool{}
)

func (k Keeper) CreateNewConcentratedLiquidityPool(ctx sdk.Context, poolId uint64, denom0, denom1 string, currSqrtPrice, currTick sdk.Int) Pool {
	fmt.Println("===0")
	poolAddr := types.NewPoolAddress(poolId)
	fmt.Println("===1")
	pool := Pool{
		Address:          poolAddr.String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
	}
	fmt.Println("===2")

	k.setPoolById(ctx, poolId, pool)
	fmt.Println("===3")

	return pool
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
	return sdk.Coin{}, fmt.Errorf("tokenIn does not match any asset in pool")
}
