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

const nonExistentDenom = "nonExistentDenom"

func TestStableswapValidateAndSetInitiailPoolAssets(t *testing.T) {
	testcase := map[string]stableSwapPoolTest{
		"0 assets - failure  - need at least 2": {
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtTooLittlePoolAssetsGiven, 0),
		},
		"1 asset - failure  - need at least 2": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtTooLittlePoolAssetsGiven, 1),
		},
		"2 assets - success": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
		},
		"3 assets - success": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
		},
		"3 assets - duplicate - failure": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDuplicateDenomFound, usdcDenom),
		},
		"3 assets - 1 denom with 0 scaling factor - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(0),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveScalingFactor, ustDenom, 0),
		},
		"3 assets - 1 denom with 0 token amount - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(0)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveTokenAmount, ustDenom, 0),
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
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
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
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
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
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
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
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},

			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDuplicateDenomFound, usdcDenom),
		},
		"error - invalid denom with 0 scaling factor - error": {
			id: 100,
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(0),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},

			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveScalingFactor, ustDenom, 0),
		},
		"error - invalid denom with 0 token amount - error": {
			id: 100,
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(0)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},

			swapFee: getDecFromStr(t, "0.01"),
			exitFee: getDecFromStr(t, "0.03"),

			futureGovernor: "testGovernor",

			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveTokenAmount, ustDenom, 0),
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
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(3454)),
					ScalingFactor: sdk.NewInt(123),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(211)),
					ScalingFactor: sdk.NewInt(3),
				},
			},
		},
		"5 assets": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
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

func TestStableswapGetPoolAssetAndIndex(t *testing.T) {
	// Pool assets in test cases must be sorted for the test to function correctly.
	testcase := map[string]stableSwapPoolTest{
		"2 assets - request each - sucess": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(3454)),
					ScalingFactor: sdk.NewInt(123),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(211)),
					ScalingFactor: sdk.NewInt(3),
				},
			},
		},
		"5 assets - request each - sucess": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
		},
		"5 asssets - non existent denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
		},
		"5 asssets - empty string denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100000)),
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

				idx, poolAsset, err := stableSwapPool.GetPoolAssetAndIndex(denomToRequest)

				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError.Error())
				require.Equal(t, stableswap.PoolAsset{}, poolAsset)
				require.Equal(t, -1, idx)
				return
			}

			for i, asset := range tc.poolAssets {

				// Test
				idx, poolAsset, err := stableSwapPool.GetPoolAssetAndIndex(asset.Token.Denom)

				require.NoError(t, err)
				require.Equal(t, asset, poolAsset)
				require.Equal(t, i, idx)
			}
		})
	}
}

func TestStableswapGetScaledPoolAmt(t *testing.T) {
	testcase := map[string]stableSwapPoolTest{
		"2 assets - request each - sucess": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(3454)),
					ScalingFactor: sdk.NewInt(123),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(211)),
					ScalingFactor: sdk.NewInt(3),
				},
			},
		},
		"5 assets - request each - sucess": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
		},
		"5 asssets - non existent denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
		},
		"5 asssets - empty string denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
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
	testcase := map[string]stableSwapPoolTest{
		"2 assets - request each - sucess": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(3454)),
					ScalingFactor: sdk.NewInt(123),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(211)),
					ScalingFactor: sdk.NewInt(3),
				},
			},
		},
		"5 assets - request each - sucess": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
		},
		"5 asssets - non existent denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
		},
		"5 asssets - empty string denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
					ScalingFactor: sdk.NewInt(10),
				},
				{
					Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
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

func TestStableswapUpdatePoolLiquidityForSwap(t *testing.T) {
	// Pool assets in test cases must be sorted for the test to function correctly.
	testcase := map[string]struct {
		base        stableSwapPoolTest
		tokensIn    sdk.Coins
		tokensOut   sdk.Coins
		shouldPanic bool
	}{
		"2 assets - single in and out - valid": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
				},
			},
			tokensIn:  sdk.Coins{sdk.NewCoin(usdcDenom, sdk.NewInt(100000000))},
			tokensOut: sdk.Coins{sdk.NewCoin(ustDenom, sdk.NewInt(50))},
		},
		"2 assets - single in and out - drained pool to 0 - error": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
				},
				expectedError: fmt.Errorf(stableswap.ErrMsgFmrDrainedPool, ustDenom, 0),
			},
			tokensIn:  sdk.Coins{sdk.NewCoin(usdcDenom, sdk.NewInt(100000000))},
			tokensOut: sdk.Coins{sdk.NewCoin(ustDenom, sdk.NewInt(100))},
		},
		"2 assets - single in and out - subtracted to negative - panic": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
				},
			},
			tokensIn:    sdk.Coins{sdk.NewCoin(usdcDenom, sdk.NewInt(100000000))},
			tokensOut:   sdk.Coins{sdk.NewCoin(ustDenom, sdk.NewInt(101))},
			shouldPanic: true,
		},
		"5 assets - single in and out - valid": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
					{
						Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
				},
			},
			tokensIn:  sdk.Coins{sdk.NewCoin(usdcDenom, sdk.NewInt(100000000))},
			tokensOut: sdk.Coins{sdk.NewCoin(ustDenom, sdk.NewInt(50))},
		},
		"5 assets - multiple in and out - valid": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
					{
						Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
				},
			},
			tokensIn:  sdk.Coins{sdk.NewCoin(usdbDenom, sdk.NewInt(200000)), sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)), sdk.NewCoin(ustDenom, sdk.NewInt(50))},
			tokensOut: sdk.Coins{sdk.NewCoin(usdaDenom, sdk.NewInt(50000)), sdk.NewCoin(ustDenom, sdk.NewInt(99))},
		},
		"5 assets - multiple in and out, in not sorted - panic": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
					{
						Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
				},
			},
			tokensIn:    sdk.Coins{sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)), sdk.NewCoin(ustDenom, sdk.NewInt(50)), sdk.NewCoin(usdbDenom, sdk.NewInt(200000))},
			tokensOut:   sdk.Coins{sdk.NewCoin(usdaDenom, sdk.NewInt(50000)), sdk.NewCoin(ustDenom, sdk.NewInt(99))},
			shouldPanic: true,
		},
		"5 assets - multiple in and out, out not sorted - panic": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
					{
						Token:         sdk.NewCoin(usdtDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdaDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
					{
						Token:         sdk.NewCoin(usdbDenom, sdk.NewInt(100000)),
						ScalingFactor: sdk.NewInt(10),
					},
				},
			},
			tokensIn:    sdk.Coins{sdk.NewCoin(usdbDenom, sdk.NewInt(200000)), sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)), sdk.NewCoin(ustDenom, sdk.NewInt(50))},
			tokensOut:   sdk.Coins{sdk.NewCoin(ustDenom, sdk.NewInt(99)), sdk.NewCoin(usdaDenom, sdk.NewInt(50000))},
			shouldPanic: true,
		},
		"2 assets - tokenIn denom does not exist - valid": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
				},
				expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
			},
			tokensIn:  sdk.Coins{sdk.NewCoin(nonExistentDenom, sdk.NewInt(100000000))},
			tokensOut: sdk.Coins{sdk.NewCoin(ustDenom, sdk.NewInt(50))},
		},
		"2 assets - tokenOut denom does not exist - valid": {
			base: stableSwapPoolTest{
				poolAssets: []stableswap.PoolAsset{
					{
						Token:         sdk.NewCoin(usdcDenom, sdk.NewInt(100000000)),
						ScalingFactor: sdk.NewInt(100000),
					},
					{
						Token:         sdk.NewCoin(ustDenom, sdk.NewInt(100)),
						ScalingFactor: sdk.NewInt(1),
					},
				},
				expectedError: fmt.Errorf(stableswap.ErrMsgFmtDenomDoesNotExist, nonExistentDenom),
			},
			tokensIn:  sdk.Coins{sdk.NewCoin(usdcDenom, sdk.NewInt(100000000))},
			tokensOut: sdk.Coins{sdk.NewCoin(nonExistentDenom, sdk.NewInt(50))},
		},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			// setup
			pool := createTestPool(t, tc.base.id, tc.base.poolAssets, tc.base.swapFee, tc.base.exitFee)

			stableSwapPool, ok := pool.(*stableswap.Pool)
			require.True(t, ok)

			defer func() {
				r := recover()
				if r == nil {
					return
				}
				if tc.shouldPanic {
					return
				}
				t.Error("Panicked when should not have")
			}()

			expectedLiqudity := stableSwapPool.GetTotalPoolLiquidity(sdk.Context{})

			// Test
			err := stableSwapPool.UpdatePoolLiquidityForSwap(tc.tokensIn, tc.tokensOut)

			// Add non-existent denom to request
			if tc.base.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.base.expectedError.Error())
				return
			}

			require.NoError(t, err)

			// Prepare expectedLiqudity
			expectedLiqudity = expectedLiqudity.Add(tc.tokensIn...)
			expectedLiqudity = expectedLiqudity.Sub(tc.tokensOut)

			require.EqualValues(t, expectedLiqudity, stableSwapPool.GetTotalPoolLiquidity(sdk.Context{}))
		})
	}
}

func getDecFromStr(t *testing.T, str string) sdk.Dec {
	dec, err := sdk.NewDecFromStr(str)
	require.NoError(t, err)
	return dec
}
