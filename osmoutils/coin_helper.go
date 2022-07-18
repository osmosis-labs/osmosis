package osmoutils

import sdk "github.com/cosmos/cosmos-sdk/types"

// TODO: Get this into the SDK https://github.com/cosmos/cosmos-sdk/issues/12538
func CoinsDenoms(coins sdk.Coins) []string {
	denoms := make([]string, len(coins))
	for i, coin := range coins {
		denoms[i] = coin.Denom
	}
	return denoms
}
