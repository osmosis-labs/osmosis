package osmomath

import (
	_ "errors"
	_ "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MulIntByU64(i sdk.Int, u uint64) (result sdk.Int) {
	if u == 0 {
		return sdk.ZeroInt()
	} else {
		result = i.Mul(sdk.NewIntFromUint64(u))
	}
	return
}

func MulCoinAmtsByU64(coins []sdk.Coin, factor []uint64) []sdk.Int {
	result := make([]sdk.Int, len(coins))
	for i, coin := range coins {
		res := MulIntByU64(coin.Amount, factor[i])
		result[i] = res
	}
	return result
}
