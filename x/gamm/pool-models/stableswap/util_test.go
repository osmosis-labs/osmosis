package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func createTestPool(t *testing.T, poolLiquidity sdk.Coins, swapFee, exitFee sdk.Dec) types.PoolI {
	pool, err := NewStableswapPool(1, PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, poolLiquidity, "")

	require.NoError(t, err)

	return &pool
}
