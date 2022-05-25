package balancer_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func createTestPool(t *testing.T, poolAssets []balancer.PoolAsset, swapFee, exitFee sdk.Dec) types.PoolI {
	pool, err := balancer.NewBalancerPool(1, balancer.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, poolAssets, "", time.Now())

	require.NoError(t, err)

	return &pool
}
