package coins

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// Mutative helpers that mutate the input coins

func MulIntMut(coins sdk.Coins, num osmomath.Int) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.Mul(num)
	}
}

func MulRawMut(coins sdk.Coins, num int64) sdk.Coins {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.MulRaw(num)
	}
	return coins
}

func MulDecMut(coins sdk.Coins, num osmomath.Dec) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.ToLegacyDec().Mul(num).TruncateInt()
	}
}

func QuoIntMut(coins sdk.Coins, num osmomath.Int) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.Quo(num)
	}
}

func QuoRawMut(coins sdk.Coins, num int64) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.QuoRaw(num)
	}
}

func QuoDecMut(coins sdk.Coins, num osmomath.Dec) {
	for i := range coins {
		coins[i].Amount = coins[i].Amount.ToLegacyDec().Quo(num).TruncateInt()
	}
}

// Non-mutative coin helpers that reallocate and return new coins

func MulInt(coins sdk.Coins, num osmomath.Int) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.Mul(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

func MulRaw(coins sdk.Coins, num int64) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.MulRaw(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

func MulDec(coins sdk.Coins, num osmomath.Dec) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.ToLegacyDec().Mul(num).TruncateInt()
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

func QuoInt(coins sdk.Coins, num osmomath.Int) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.Quo(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

func QuoRaw(coins sdk.Coins, num int64) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.QuoRaw(num)
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}

func QuoDec(coins sdk.Coins, num osmomath.Dec) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))

	for i := range coins {
		newCoins[i].Amount = coins[i].Amount.ToLegacyDec().Quo(num).TruncateInt()
		newCoins[i].Denom = coins[i].Denom
	}

	return newCoins
}
