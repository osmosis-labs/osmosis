package stableswap_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
)

type stableSwapPoolTest struct {
	id uint64

	poolAssets []stableswap.PoolAsset

	swapFee sdk.Dec
	exitFee sdk.Dec

	futureGovernor string

	expectedError error
}

func TestStableswapCreatePool(t *testing.T) {
	testcase := map[string]stableSwapPoolTest{
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

func TestStableswapGetTotalPoolLiquidity(t *testing.T) {
	testcase := map[string]stableSwapPoolTest{
		"2 assets": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(3454)),
					ScalingFactor: sdk.NewInt(123),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(211)),
					ScalingFactor: sdk.NewInt(3),
				},
			},
		},
		"5 assets": {
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
				{
					Token:         sdk.NewCoin("usda", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin("usdb", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// setup
			pool := createTestPool(t, tc.poolAssets, tc.swapFee, tc.exitFee)

			expectedLiquidity := sdk.NewCoins()
			for _, asset := range tc.poolAssets {
				expectedLiquidity = expectedLiquidity.Add(asset.Token)
			}

			actualLiquidity := pool.GetTotalPoolLiquidity(sdk.Context{})

			require.EqualValues(t, expectedLiquidity, actualLiquidity)
		})
	}
}

func getDecFromStr(t *testing.T, str string) sdk.Dec {
	dec, err := sdk.NewDecFromStr(str)
	require.NoError(t, err)
	return dec
}
