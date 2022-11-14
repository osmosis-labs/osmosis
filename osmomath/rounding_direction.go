package osmomath

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RoundingDirection int

const (
	RoundUnconstrained RoundingDirection = 0
	RoundUp            RoundingDirection = 1
	RoundDown          RoundingDirection = 2
	RoundBankers       RoundingDirection = 3
)

func DivIntByU64ToBigDec(i sdk.Int, u uint64, round RoundingDirection) (BigDec, error) {
	if u == 0 {
		return BigDec{}, errors.New("div by zero")
	}
	d := BigDecFromSDKDec(i.ToDec())
	if round == RoundUp {
		return d.QuoRoundUp(NewBigDec(int64(u))), nil
	} else if round == RoundDown {
		return d.QuoInt64(int64(u)), nil
	} else if round == RoundBankers {
		return d.Quo(NewBigDec(int64(u))), nil
	}
	return BigDec{}, fmt.Errorf("invalid rounding mode %d", int(round))
}

func DivCoinAmtsByU64ToBigDec(coins []sdk.Coin, scales []uint64, round RoundingDirection) ([]BigDec, error) {
	result := make([]BigDec, len(coins))
	for i, coin := range coins {
		res, err := DivIntByU64ToBigDec(coin.Amount, scales[i], round)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}
	return result, nil
}
