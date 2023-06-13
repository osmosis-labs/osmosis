package math

import sdk "github.com/cosmos/cosmos-sdk/types"

func PriceToTickExact(price sdk.Dec) (int64, error) {
	return priceToTickExact(price)
}
