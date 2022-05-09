package stableswap_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
)

func TestStableswapCreatePool(t *testing.T) {
	testcase := map[string]struct {
		id uint64

		poolAssets []stableswap.PoolAsset

		swapFee sdk.Dec
		exitFee sdk.Dec

		futureGovernor string

		expectedError error
	}{
		"success": {
			id: 100,
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin("usdt", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},

			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// setup
			pool, err := stableswap.NewStableswapPool(tc.id, stableswap.PoolParams{
				SwapFee: tc.swapFee,
				ExitFee: tc.exitFee,
			},
				tc.poolAssets,
				tc.futureGovernor)

			require.NoError(t, err)
			require.NotNil(t, pool)

			stableSwapPool, ok := pool.(*stableswap.Pool)
			require.True(t, ok)

			require.Equal(t, tc.poolAssets, stableSwapPool.PoolAssets)
		})
	}
}

func getDecFromStr(t *testing.T, str string) sdk.Dec {
	dec, err := sdk.NewDecFromStr(str)
	require.NoError(t, err)
	return dec
}
