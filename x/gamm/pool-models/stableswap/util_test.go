package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v18/x/gamm/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

func createTestPool(t *testing.T, poolLiquidity sdk.Coins, spreadFactor, exitFee osmomath.Dec, scalingFactors []uint64) types.CFMMPoolI {
	t.Helper()
	scalingFactors, _ = applyScalingFactorMultiplier(scalingFactors)

	pool, err := NewStableswapPool(1, PoolParams{
		SwapFee: spreadFactor,
		ExitFee: exitFee,
	}, poolLiquidity, scalingFactors, "", "")

	require.NoError(t, err)

	return &pool
}
