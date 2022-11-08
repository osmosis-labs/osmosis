package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return cltypes.OrderInitialPoolDenoms(denom0, denom1)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, liquidityCreated sdk.Dec, err error) {
	return k.createPosition(ctx, poolId, owner, amount0Desired, amount1Desired, amount0Min, amount1Min, lowerTick, upperTick)
}

func GetLiquidityFromAmounts(sqrtPrice, sqrtPriceA, sqrtPriceB sdk.Dec, amount0, amount1 sdk.Int) (liquidity sdk.Dec) {
	return getLiquidityFromAmounts(sqrtPrice, sqrtPriceA, sqrtPriceB, amount0, amount1)
}

func PriceToTick(price sdk.Dec) sdk.Int {
	return priceToTick(price)
}

func TickToSqrtPrice(tickIndex sdk.Int) (sdk.Dec, error) {
	return tickToSqrtPrice(tickIndex)
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo TickInfo) {
	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
}

func GetNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return getNextSqrtPriceFromAmount0RoundingUp(sqrtPriceCurrent, liquidity, amountRemaining)
}

func GetNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining sdk.Dec) (sqrtPriceNext sdk.Dec) {
	return getNextSqrtPriceFromAmount1RoundingDown(sqrtPriceCurrent, liquidity, amountRemaining)
}

func CalcAmount0Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec, roundUp bool) sdk.Dec {
	return calcAmount0Delta(liq, sqrtPriceA, sqrtPriceB, roundUp)
}

func CalcAmount1Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec, roundUp bool) sdk.Dec {
	return calcAmount1Delta(liq, sqrtPriceA, sqrtPriceB, roundUp)
}

func ComputeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec, zeroForOne bool) (sqrtPriceNext, amountIn, amountOut sdk.Dec) {
	return computeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining, zeroForOne)
}

func Liquidity0(amount sdk.Int, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	return liquidity0(amount, sqrtPriceA, sqrtPriceB)
}

func Liquidity1(amount sdk.Int, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	return liquidity1(amount, sqrtPriceA, sqrtPriceB)
}

func (k Keeper) GetPoolbyId(ctx sdk.Context, poolId uint64) Pool {
	return k.getPoolbyId(ctx, poolId)
}
