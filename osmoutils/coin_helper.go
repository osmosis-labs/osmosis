package osmoutils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Get this into the SDK https://github.com/cosmos/cosmos-sdk/issues/12538
func CoinsDenoms(coins sdk.Coins) []string {
	denoms := make([]string, len(coins))
	for i, coin := range coins {
		denoms[i] = coin.Denom
	}
	return denoms
}

// MinCoins returns the minimum of each denom between both coins.
// For now it assumes they have the same denoms.
// TODO: Replace with method in SDK once we update our version
func MinCoins(coinsA sdk.Coins, coinsB sdk.Coins) sdk.Coins {
	resCoins := sdk.Coins{}
	for i, coin := range coinsA {
		if coinsB[i].Amount.GT(coin.Amount) {
			resCoins = append(resCoins, coin)
		} else {
			resCoins = append(resCoins, coinsB[i])
		}
	}
	return resCoins
}

// SubDecCoinArrays subtracts the contents of the second param from the first (decCoinsArrayA - decCoinsArrayB)
// Note that this takes in two _arrays_ of DecCoins, meaning that each term itself is of type DecCoins (i.e. an array of DecCoin).
func SubDecCoinArrays(decCoinsArrayA []sdk.DecCoins, decCoinsArrayB []sdk.DecCoins) ([]sdk.DecCoins, error) {
	if len(decCoinsArrayA) != len(decCoinsArrayB) {
		return []sdk.DecCoins{}, fmt.Errorf("DecCoin arrays must be of equal length to be subtracted")
	}

	finalDecCoinArray := []sdk.DecCoins{}
	for i := range decCoinsArrayA {
		finalDecCoinArray = append(finalDecCoinArray, decCoinsArrayA[i].Sub(decCoinsArrayB[i]))
	}

	return finalDecCoinArray, nil
}
