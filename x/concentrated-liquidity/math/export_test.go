package math

import sdk "github.com/cosmos/cosmos-sdk/types"

func PriceToTick(price sdk.Dec) (int64, error) {
	return priceToTick(price)
}
