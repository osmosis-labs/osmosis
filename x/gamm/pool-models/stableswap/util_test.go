package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func createTestPool(t *testing.T, poolLiquidity sdk.Coins, swapFee, exitFee sdk.Dec, scalingFactors []uint64) swaproutertypes.PoolI {
	pool, err := NewStableswapPool(1, PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, poolLiquidity, applyScalingFactorMultiplier(scalingFactors), "", "")

	require.NoError(t, err)

	return &pool
}
