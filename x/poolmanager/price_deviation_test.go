package poolmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/x/poolmanager"
)

// TestCalculatePriceDeviationLargeAmount ensures that the price deviation used by the
// EstimateTradeBasedOnPriceImpact queries handles input amounts that do not fit into an
// int64. Token amounts are sdk.Int (up to 2^256), and amounts above 2^63-1 are common for
// assets with 18 decimals, so the helper must not route the amount through Int64().
func TestCalculatePriceDeviationLargeAmount(t *testing.T) {
	// 1e19 is greater than math.MaxInt64 (~9.22e18), so Amount.Int64() would panic.
	fromAmount, ok := osmomath.NewIntFromString("10000000000000000000")
	require.True(t, ok)
	outAmount, ok := osmomath.NewIntFromString("4000000000000000000")
	require.True(t, ok)

	fromCoin := sdk.NewCoin("uosmo", fromAmount)
	tokenOut := sdk.NewCoin("uion", outAmount)
	spotPrice := osmomath.NewDec(2)

	// currTradePrice = 1e19 / 4e18 = 2.5
	// priceDeviation = |(2.5 - 2) / 2| = 0.25
	got := poolmanager.CalculatePriceDeviation(fromCoin, tokenOut, spotPrice)
	require.Equal(t, osmomath.MustNewDecFromStr("0.25"), got)
}
