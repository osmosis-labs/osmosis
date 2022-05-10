package stableswap_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
	types "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func createTestPool(t *testing.T, poolAssets []stableswap.PoolAsset, swapFee, exitFee sdk.Dec) types.PoolI {
	pool, err := stableswap.NewStableswapPool(1, stableswap.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	},
		poolAssets,
		"")

	require.NoError(t, err)
	require.NotNil(t, pool)

	return pool
}
