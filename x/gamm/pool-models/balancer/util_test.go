package balancer_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
)

func createTestPool(t *testing.T, spreadFactor, exitFee sdk.Dec, poolAssets ...balancer.PoolAsset) *balancer.Pool {
	t.Helper()
	pool, err := balancer.NewBalancerPool(
		1,
		balancer.NewPoolParams(spreadFactor, exitFee, nil),
		poolAssets,
		"",
		time.Now(),
	)
	require.NoError(t, err)

	return &pool
}

func assertExpectedSharesErrRatio(t *testing.T, expectedShares, actualShares sdk.Int) {
	t.Helper()
	allowedErrRatioDec, err := sdk.NewDecFromStr(allowedErrRatio)
	require.NoError(t, err)

	errTolerance := osmomath.ErrTolerance{
		MultiplicativeTolerance: allowedErrRatioDec,
	}

	require.Equal(
		t,
		0,
		errTolerance.Compare(expectedShares, actualShares),
		fmt.Sprintf("expectedShares: %s, actualShares: %s", expectedShares.String(), actualShares.String()))
}

func assertExpectedLiquidity(t *testing.T, tokensJoined, liquidity sdk.Coins) {
	t.Helper()
	require.Equal(t, tokensJoined, liquidity)
}

// assertPoolStateNotModified asserts that sut (system under test) does not modify
// pool state.
func assertPoolStateNotModified(t *testing.T, pool *balancer.Pool, sut func()) {
	t.Helper()
	// We need to make sure that this method does not mutate state.
	oldPoolAssets := pool.GetAllPoolAssets()
	oldLiquidity := pool.GetTotalPoolLiquidity(sdk.Context{})
	oldShares := pool.GetTotalShares()

	sut()

	newPoolAssets := pool.GetAllPoolAssets()
	newLiquidity := pool.GetTotalPoolLiquidity(sdk.Context{})
	newShares := pool.GetTotalShares()

	require.Equal(t, oldPoolAssets, newPoolAssets)
	require.Equal(t, oldLiquidity, newLiquidity)
	require.Equal(t, oldShares, newShares)
}
