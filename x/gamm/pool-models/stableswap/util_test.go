package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func createTestPool(t *testing.T, poolLiquidity sdk.Coins, swapFee, exitFee sdk.Dec, scalingFactors []uint64) types.PoolI {
	pool, err := NewStableswapPool(1, PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, poolLiquidity, applyScalingFactorMultiplier(scalingFactors), "", "")

	require.NoError(t, err)

	return &pool
}
