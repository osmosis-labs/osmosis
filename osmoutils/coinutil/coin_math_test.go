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
		sdk.NewCoin("foo", sdk.NewInt(100)),
		sdk.NewCoin("bar", sdk.NewInt(200)),
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
			defaulCoins := deepCopy(defaultCoins)
			coinutil.MulIntMut(defaulCoins, defaultMultiplier)
			require.Equal(t, defaultMulExpectedResult, defaulCoins)
		})

		t.Run("MulIntRawMut", func(t *testing.T) {
			defaulCoins := deepCopy(defaultCoins)
			coinutil.MulRawMut(defaulCoins, defaultMultiplier.Int64())
			require.Equal(t, defaultMulExpectedResult, defaulCoins)
		})

		t.Run("MulDecMut", func(t *testing.T) {
			defaulCoins := deepCopy(defaultCoins)
			coinutil.MulDecMut(defaulCoins, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultMulExpectedResult, defaulCoins)
		})
	})

	// Make a deep copy of the default coins for the input.
	// Validate that the copy input coins are not mutated.
	t.Run("test non-mutative multiplication", func(t *testing.T) {
		t.Run("MulInt", func(t *testing.T) {
			defaulCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.MulInt(defaulCoinsCopy, defaultMultiplier)
			require.Equal(t, defaultMulExpectedResult, result)
			require.Equal(t, defaultCoins, defaulCoinsCopy)
		})

		t.Run("MulIntRaw", func(t *testing.T) {
			defaulCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.MulRaw(defaulCoinsCopy, defaultMultiplier.Int64())
			require.Equal(t, defaultMulExpectedResult, result)
			require.Equal(t, defaultCoins, defaulCoinsCopy)
		})

		t.Run("MulDec", func(t *testing.T) {
			defaulCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.MulDec(defaulCoinsCopy, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultMulExpectedResult, result)
			require.Equal(t, defaultCoins, defaulCoinsCopy)
		})
	})
}

func TestQuo(t *testing.T) {
	t.Run("test mutative division", func(t *testing.T) {
		t.Run("QuoIntMut", func(t *testing.T) {
			defaulCoins := deepCopy(defaultCoins)
			coinutil.QuoIntMut(defaulCoins, defaultMultiplier)
			require.Equal(t, defaultQuoExpectedResult, defaulCoins)
		})

		t.Run("QuoIntRawMut", func(t *testing.T) {
			defaulCoins := deepCopy(defaultCoins)
			coinutil.QuoRawMut(defaulCoins, defaultMultiplier.Int64())
			require.Equal(t, defaultQuoExpectedResult, defaulCoins)
		})

		t.Run("QuoDecMut", func(t *testing.T) {
			defaulCoins := deepCopy(defaultCoins)
			coinutil.QuoDecMut(defaulCoins, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultQuoExpectedResult, defaulCoins)
		})
	})

	// Make a deep copy of the default coins for the input.
	// Validate that the copy input coins are not mutated.
	t.Run("test non-mutative division", func(t *testing.T) {
		t.Run("QuoInt", func(t *testing.T) {
			defaulCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.QuoInt(defaulCoinsCopy, defaultMultiplier)
			require.Equal(t, defaultQuoExpectedResult, result)
			require.Equal(t, defaultCoins, defaulCoinsCopy)
		})

		t.Run("QuoIntRaw", func(t *testing.T) {
			defaulCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.QuoRaw(defaulCoinsCopy, defaultMultiplier.Int64())
			require.Equal(t, defaultQuoExpectedResult, result)
			require.Equal(t, defaultCoins, defaulCoinsCopy)
		})

		t.Run("QuoDec", func(t *testing.T) {
			defaulCoinsCopy := deepCopy(defaultCoins)
			result := coinutil.QuoDec(defaulCoinsCopy, osmomath.NewDecFromInt(defaultMultiplier))
			require.Equal(t, defaultQuoExpectedResult, result)
			require.Equal(t, defaultCoins, defaulCoinsCopy)
		})
	})
}
