package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

<<<<<<< HEAD
func createTestPool(t *testing.T, poolLiquidity sdk.Coins, swapFee, exitFee sdk.Dec, scalingFactors []uint64) types.PoolI {
=======
func createTestPool(t *testing.T, poolLiquidity sdk.Coins, swapFee, exitFee sdk.Dec, scalingFactors []uint64) types.CFMMPoolI {
	scalingFactors, _ = applyScalingFactorMultiplier(scalingFactors)

>>>>>>> 2ac5d356 (Gamm stableswap improvements (#3839))
	pool, err := NewStableswapPool(1, PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, poolLiquidity, scalingFactors, "", "")

	require.NoError(t, err)

	return &pool
}
