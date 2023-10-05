package coinutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// Mutative helpers that mutate the input coins

// MulIntMut multiplies the coins by the given integer
// Mutates the input coins
func MulIntMut(coins sdk.Coins, num osmomath.Int) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.Mul(num)
	}
}

// MulRawMut multiplies the coins by the given integer
// Mutates the input coins
func MulRawMut(coins sdk.Coins, num int64) sdk.Coins {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.MulRaw(num)
	}
	return coins
}

// MulDecMut multiplies the coins by the given decimal
// Mutates the input coins
func MulDecMut(coins sdk.Coins, num osmomath.Dec) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.ToLegacyDec().Mul(num).TruncateInt()
	}
}

// QuoIntMut divides the coins by the given integer
// Mutates the input coins
func QuoIntMut(coins sdk.Coins, num osmomath.Int) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.Quo(num)
	}
}

// QuoRawMut divides the coins by the given integer
// Mutates the input coins
func QuoRawMut(coins sdk.Coins, num int64) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.QuoRaw(num)
	}
}

// QuoIntMut divides the coins by the given decimal
// Mutates the input coins
func QuoDecMut(coins sdk.Coins, num osmomath.Dec) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.ToLegacyDec().Quo(num).TruncateInt()
	}
}

// Non-mutative coin helpers that reallocate and return new coins

// MulInt multiplies the coins by the given integer
// Does not mutate the input coins and returns new coins.
func MulInt(coins sdk.Coins, num osmomath.Int) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.Mul(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

// MulRaw multiplies the coins by the given integer
// Does not mutate the input coins and returns new coins.
func MulRaw(coins sdk.Coins, num int64) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.MulRaw(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

// MulDec multiplies the coins by the given decimal
// Does not mutate the input coins and returns new coins.
func MulDec(coins sdk.Coins, num osmomath.Dec) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.ToLegacyDec().Mul(num).TruncateInt()
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

// QuoInt divides the coins by the given integer
// Does not mutate the input coins and returns new coins.
func QuoInt(coins sdk.Coins, num osmomath.Int) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.Quo(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

// QuoRaw divides the coins by the given integer
// Does not mutate the input coins and returns new coins.
func QuoRaw(coins sdk.Coins, num int64) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.QuoRaw(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

// QuoDec divides the coins by the given integer
// Does not mutate the input coins and returns new coins.
func QuoDec(coins sdk.Coins, num osmomath.Dec) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.ToLegacyDec().Quo(num).TruncateInt()
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}
