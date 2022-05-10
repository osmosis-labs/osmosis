package stableswap_test

import (
	"errors"
	fmt "fmt"
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

func TestStableswapValidateAndSetInitiailPoolAssets(t *testing.T) {
	testcase := map[string]stableSwapPoolTest{
		"0 assets - failure  - need at least 2": {
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtTooLittlePoolAssetsGiven, 0),
		},
		"1 asset - failure  - need at least 2": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtTooLittlePoolAssetsGiven, 1),
		},
		"2 assets - success": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
		},
		"3 assets - success": {
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
		},
		"3 assets - duplicate - failure": {
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
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDuplicateDenomFound, "usdc"),
		},
		"3 assets - 1 denom with 0 scaling factor - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(0),
				},
				{
					Token:         sdk.NewCoin("usdt", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveScalingFactor, "ust", 0),
		},
		"3 assets - 1 denom with 0 token amount - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(0)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin("usdt", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveTokenAmount, "ust", 0),
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// Setup
			pool := createTestPoolNoValidation(t, tc.id, tc.poolAssets, tc.swapFee, tc.exitFee)
			require.NotNil(t, pool)

			stableSwapPool, ok := pool.(*stableswap.Pool)
			require.True(t, ok)

			// Test
			actual := stableSwapPool.ValidateAndSortInitialPoolAssets()

			require.Equal(t, tc.expectedError, actual)
		})
	}
}

func TestStableswapCreatePool(t *testing.T) {
	testcase := map[string]stableSwapPoolTest{
		"1 0 assets - error, need at least 2": {
			id:      100,
			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtTooLittlePoolAssetsGiven, 0),
		},
		"1 asset - error, need at least 2": {
			id: 100,
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
			},
			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtTooLittlePoolAssetsGiven, 1),
		},
		"2 assets - success": {
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
			},
			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",
		},
		"3 assets - success": {
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
		"error - invalid pool assets - duplicate": {
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
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},

			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDuplicateDenomFound, "usdc"),
		},
		"error - invalid denom with 0 scaling factor - error": {
			id: 100,
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(0),
				},
				{
					Token:         sdk.NewCoin("usdt", sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},

			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveScalingFactor, "ust", 0),
		},
		"error - invalid denom with 0 token amount - error": {
			id: 100,
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(0)),
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

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveTokenAmount, "ust", 0),
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// Test
			pool, err := stableswap.NewStableswapPool(tc.id, stableswap.PoolParams{
				SwapFee: tc.swapFee,
				ExitFee: tc.exitFee,
			},
				tc.poolAssets,
				tc.futureGovernor)

			if tc.expectedError != nil {
				require.Error(t, tc.expectedError, err)
				require.Nil(t, pool)
				return
			}

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
			pool := createTestPool(t, tc.id, tc.poolAssets, tc.swapFee, tc.exitFee)

			expectedLiquidity := sdk.NewCoins()
			for _, asset := range tc.poolAssets {
				expectedLiquidity = expectedLiquidity.Add(asset.Token)
			}

			// Test
			actualLiquidity := pool.GetTotalPoolLiquidity(sdk.Context{})

			require.EqualValues(t, expectedLiquidity, actualLiquidity)
		})
	}
}

func TestStableswapGetScaledPoolAmt(t *testing.T) {
	const nonExistentDenom = "nonExistentDenom"

	testcase := map[string]stableSwapPoolTest{
		"2 assets - request each - sucess": {
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
		"5 assets - request each - sucess": {
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
		"5 asssets - non existent denom - error": {
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
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
		},
		"5 asssets - empty string denom - error": {
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
			expectedError: errors.New(stableswap.ErrMsgEmptyDenomGiven),
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// setup
			pool := createTestPool(t, tc.id, tc.poolAssets, tc.swapFee, tc.exitFee)

			stableSwapPool, ok := pool.(*stableswap.Pool)
			require.True(t, ok)

			// Add non-existent denom to request
			if tc.expectedError != nil {
				// Test
				var denomToRequest string
				if tc.expectedError.Error() != stableswap.ErrMsgEmptyDenomGiven {
					denomToRequest = nonExistentDenom
				}

				actualScaled, err := stableSwapPool.GetScaledPoolAmt(denomToRequest)

				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError.Error())
				require.Equal(t, sdk.Int{}, actualScaled)
				return
			}

			for _, asset := range tc.poolAssets {

				expectedScaled := asset.Token.Amount.Quo(asset.ScalingFactor)

				// Test
				actualScaled, err := stableSwapPool.GetScaledPoolAmt(asset.Token.GetDenom())

				require.NoError(t, err)
				require.Equal(t, expectedScaled, actualScaled)
			}
		})
	}
}

func TestStableswapGetDeScaledPoolAmt(t *testing.T) {
	const nonExistentDenom = "nonExistentDenom"

	testcase := map[string]stableSwapPoolTest{
		"2 assets - request each - sucess": {
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
		"5 assets - request each - sucess": {
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
		"5 asssets - non existent denom - error": {
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
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
		},
		"5 asssets - empty string denom - error": {
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
			expectedError: errors.New(stableswap.ErrMsgEmptyDenomGiven),
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// setup
			pool := createTestPool(t, tc.id, tc.poolAssets, tc.swapFee, tc.exitFee)

			stableSwapPool, ok := pool.(*stableswap.Pool)
			require.True(t, ok)

			// Always request descale first amount to ensure that it is done on the input
			toDescaleRequestDec := tc.poolAssets[0].Token.Amount.ToDec()

			// Add non-existent denom to request
			if tc.expectedError != nil {
				// Test
				var denomToRequest string
				if tc.expectedError.Error() != stableswap.ErrMsgEmptyDenomGiven {
					denomToRequest = nonExistentDenom
				}

				actualDescaled, err := stableSwapPool.GetDescaledPoolAmt(denomToRequest, toDescaleRequestDec)

				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError.Error())
				require.Equal(t, sdk.Dec{}, actualDescaled)
				return
			}

			for _, asset := range tc.poolAssets {

				expectedScaled := toDescaleRequestDec.MulInt(asset.ScalingFactor)

				// Test
				actualDescaled, err := stableSwapPool.GetDescaledPoolAmt(asset.Token.GetDenom(), toDescaleRequestDec)

				require.NoError(t, err)
				require.Equal(t, expectedScaled, actualDescaled)
			}
		})
	}
}

func getDecFromStr(t *testing.T, str string) sdk.Dec {
	dec, err := sdk.NewDecFromStr(str)
	require.NoError(t, err)
	return dec
}
