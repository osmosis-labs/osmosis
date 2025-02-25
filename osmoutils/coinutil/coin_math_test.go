package coinutil_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
)

var (
	defaultCoins = sdk.NewCoins(
		sdk.NewCoin("foo", osmomath.NewInt(100)),
		sdk.NewCoin("bar", osmomath.NewInt(200)),
	)

	defaultMultiplier = osmomath.NewInt(2)

	defaultMulExpectedResult = sdk.NewCoins(
		sdk.NewCoin(defaultCoins[0].Denom, defaultCoins[0].Amount.Mul(defaultMultiplier)),
		sdk.NewCoin(defaultCoins[1].Denom, defaultCoins[1].Amount.Mul(defaultMultiplier)),
	)

	defaultQuoExpectedResult = sdk.NewCoins(
		sdk.NewCoin(defaultCoins[0].Denom, defaultCoins[0].Amount.Quo(defaultMultiplier)),
		sdk.NewCoin(defaultCoins[1].Denom, defaultCoins[1].Amount.Quo(defaultMultiplier)),
	)
)

// makes a deep copy to avoid accidentally mutating the input to a test.
func deepCopy(coins sdk.Coins) sdk.Coins {
	newCoins := make(sdk.Coins, len(coins))
	for i := range coins {
		newCoins[i].Amount = coins[i].Amount
		newCoins[i].Denom = coins[i].Denom
	}
	return newCoins
}

// Basic multiplication test.
func TestMul(t *testing.T) {
	t.Run("test mutative multiplication", func(t *testing.T) {
		t.Run("MulIntMut", func(t *testing.T) {
			defaultCoins := deepCopy(defaultCoins)
			coinutil.MulIntMut(defaultCoins, defaultMultiplier)
			require.Equal(t, defaultMulExpectedResult, defaultCoins)
		})

		t.Run("MulIntRawMut", func(t *testing.T) {
			defaultCoins := deepCopy(defaultCoins)
			coinutil.MulRawMut(defaultCoins, defaultMultiplier.Int64())
			require.Equal(t, defaultMulExpectedResult, defaultCoins)
		})

		t.Run("MulDecMut", func(t *testing.T) {
			defaultCoins := deepCopy(defaultCoins)
			coinutil.MulDecMut(defaultCoins, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultMulExpectedResult, defaultCoins)
		})
	})

	// Make a deep copy of the default coins for the input.
	// Validate that the copy input coins are not mutated.
	t.Run("test non-mutative multiplication", func(t *testing.T) {
		t.Run("MulInt", func(t *testing.T) {
			defaultCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.MulInt(defaultCoinsCopy, defaultMultiplier)
			require.Equal(t, defaultMulExpectedResult, result)
			require.Equal(t, defaultCoins, defaultCoinsCopy)
		})

		t.Run("MulIntRaw", func(t *testing.T) {
			defaultCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.MulRaw(defaultCoinsCopy, defaultMultiplier.Int64())
			require.Equal(t, defaultMulExpectedResult, result)
			require.Equal(t, defaultCoins, defaultCoinsCopy)
		})

		t.Run("MulDec", func(t *testing.T) {
			defaultCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.MulDec(defaultCoinsCopy, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultMulExpectedResult, result)
			require.Equal(t, defaultCoins, defaultCoinsCopy)
		})
	})
}

func TestQuo(t *testing.T) {
	t.Run("test mutative division", func(t *testing.T) {
		t.Run("QuoIntMut", func(t *testing.T) {
			defaultCoins := deepCopy(defaultCoins)
			coinutil.QuoIntMut(defaultCoins, defaultMultiplier)
			require.Equal(t, defaultQuoExpectedResult, defaultCoins)
		})

		t.Run("QuoIntRawMut", func(t *testing.T) {
			defaultCoins := deepCopy(defaultCoins)
			coinutil.QuoRawMut(defaultCoins, defaultMultiplier.Int64())
			require.Equal(t, defaultQuoExpectedResult, defaultCoins)
		})

		t.Run("QuoDecMut", func(t *testing.T) {
			defaultCoins := deepCopy(defaultCoins)
			coinutil.QuoDecMut(defaultCoins, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultQuoExpectedResult, defaultCoins)
		})
	})

	// Make a deep copy of the default coins for the input.
	// Validate that the copy input coins are not mutated.
	t.Run("test non-mutative division", func(t *testing.T) {
		t.Run("QuoInt", func(t *testing.T) {
			defaultCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.QuoInt(defaultCoinsCopy, defaultMultiplier)
			require.Equal(t, defaultQuoExpectedResult, result)
			require.Equal(t, defaultCoins, defaultCoinsCopy)
		})

		t.Run("QuoIntRaw", func(t *testing.T) {
			defaultCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.QuoRaw(defaultCoinsCopy, defaultMultiplier.Int64())
			require.Equal(t, defaultQuoExpectedResult, result)
			require.Equal(t, defaultCoins, defaultCoinsCopy)
		})

		t.Run("QuoDec", func(t *testing.T) {
			defaultCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.QuoDec(defaultCoinsCopy, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultQuoExpectedResult, result)
			require.Equal(t, defaultCoins, defaultCoinsCopy)
		})
	})
}
