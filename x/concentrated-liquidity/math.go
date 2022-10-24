package concentrated_liquidity

import sdk "github.com/cosmos/cosmos-sdk/types"

// liquidity0 takes an amount of asset0 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
func liquidity0(amount int64, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	product := sqrtPriceA.Mul(sqrtPriceB)
	diff := sqrtPriceB.Sub(sqrtPriceA)
	amt := sdk.NewDec(amount)
	return amt.Mul(product.Quo(diff))
}

// liquidity1 takes an amount of asset1 in the pool as well as the sqrtpCur and the nextPrice
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// liquidity1 = amount / (sqrtPriceB - sqrtPriceA)
func liquidity1(amount int64, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	amt := sdk.NewDec(amount)
	return amt.Quo(diff)
}

// calcAmount0 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 0
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount0Delta = (liquidity * (sqrtPriceA - sqrtPriceB)) / (sqrtPriceA * sqrtPriceB)
func calcAmount0Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	mult := liq
	return (mult.Mul(diff)).Quo(sqrtPriceB).Quo(sqrtPriceA)
}

// calcAmount1 takes the asset with the smaller liqudity in the pool as well as the sqrtpCur and the nextPrice and calculates the amount of asset 1
// sqrtPriceA is the smaller of sqrtpCur and the nextPrice
// sqrtPriceB is the larger of sqrtpCur and the nextPrice
// calcAmount1Delta = liq * (sqrtPriceA - sqrtPriceB)
func calcAmount1Delta(liq, sqrtPriceA, sqrtPriceB sdk.Dec) sdk.Dec {
	if sqrtPriceA.GT(sqrtPriceB) {
		sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
	}
	diff := sqrtPriceB.Sub(sqrtPriceA)
	return liq.Mul(diff)
}
