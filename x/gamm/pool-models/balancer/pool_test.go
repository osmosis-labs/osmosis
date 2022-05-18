package balancer

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	defaultSwapFee            = sdk.MustNewDecFromStr("0.025")
	defaultExitFee            = sdk.MustNewDecFromStr("0.025")
	defaultPoolId             = uint64(10)
	defaultBalancerPoolParams = PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultFutureGovernor = ""
	defaultCurBlockTime   = time.Unix(1618700000, 0)

	dummyPoolAssets = []PoolAsset{}
	wantErr         = true
	noErr           = false
)

// Expected is un-scaled
func testTotalWeight(t *testing.T, expected sdk.Int, pool Pool) {
	scaledExpected := expected.MulRaw(GuaranteedWeightPrecision)
	require.Equal(t,
		scaledExpected.String(),
		pool.GetTotalWeight().String())
}

func TestBalancerPoolParams(t *testing.T) {
	// Tests that creating a pool with the given pair of swapfee and exit fee
	// errors or succeeds as intended. Furthermore, it checks that
	// NewPool panics in the error case.
	tests := []struct {
		SwapFee   sdk.Dec
		ExitFee   sdk.Dec
		shouldErr bool
	}{
		// Should work
		{defaultSwapFee, defaultExitFee, noErr},
		// Can't set the swap fee as negative
		{sdk.NewDecWithPrec(-1, 2), defaultExitFee, wantErr},
		// Can't set the swap fee as 1
		{sdk.NewDec(1), defaultExitFee, wantErr},
		// Can't set the swap fee above 1
		{sdk.NewDecWithPrec(15, 1), defaultExitFee, wantErr},
		// Can't set the exit fee as negative
		{defaultSwapFee, sdk.NewDecWithPrec(-1, 2), wantErr},
		// Can't set the exit fee as 1
		{defaultSwapFee, sdk.NewDec(1), wantErr},
		// Can't set the exit fee above 1
		{defaultSwapFee, sdk.NewDecWithPrec(15, 1), wantErr},
	}

	for i, params := range tests {
		PoolParams := PoolParams{
			SwapFee: params.SwapFee,
			ExitFee: params.ExitFee,
		}
		err := PoolParams.Validate(dummyPoolAssets)
		if params.shouldErr {
			require.Error(t, err, "unexpected lack of error, tc %v", i)
			// Check that these are also caught if passed to the underlying pool creation func
			_, err = NewBalancerPool(1, PoolParams, dummyPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
			require.Error(t, err)
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
		}
	}
}

// TODO: Refactor this into multiple tests
func TestBalancerPoolUpdatePoolAssetBalance(t *testing.T) {
	var poolId uint64 = 10

	initialAssets := []PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
	}

	pacc, err := NewBalancerPool(poolId, defaultBalancerPoolParams, initialAssets, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	_, err = pacc.GetPoolAsset("unknown")
	require.Error(t, err)
	_, err = pacc.GetPoolAsset("")
	require.Error(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc)

	// Break abstractions and start reasoning about the underlying internal representation's APIs.
	// TODO: This test actually just needs to be refactored to not be doing this, and just
	// create a different pool each time.

	err = pacc.setInitialPoolAssets([]PoolAsset{{
		Weight: sdk.NewInt(-1),
		Token:  sdk.NewCoin("negativeWeight", sdk.NewInt(50000)),
	}})

	require.Error(t, err)

	err = pacc.setInitialPoolAssets([]PoolAsset{{
		Weight: sdk.NewInt(0),
		Token:  sdk.NewCoin("zeroWeight", sdk.NewInt(50000)),
	}})
	require.Error(t, err)

	err = pacc.UpdatePoolAssetBalance(
		sdk.NewCoin("test1", sdk.NewInt(0)))
	require.Error(t, err)

	err = pacc.UpdatePoolAssetBalance(
		sdk.Coin{Denom: "test1", Amount: sdk.NewInt(-1)},
	)
	require.Error(t, err)

	err = pacc.UpdatePoolAssetBalance(
		sdk.NewCoin("test1", sdk.NewInt(1)))
	require.NoError(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc)

	PoolAsset, err := pacc.GetPoolAsset("test1")
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(1).String(), PoolAsset.Token.Amount.String())
}

func TestBalancerPoolAssetsWeightAndTokenBalance(t *testing.T) {
	// TODO: Add more cases
	// asset names should be i ascending order, starting from test1
	tests := []struct {
		assets    []PoolAsset
		shouldErr bool
	}{
		// weight 0
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(0),
					Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
				},
			},
			wantErr,
		},
		// negative weight
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(-1),
					Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
				},
			},
			wantErr,
		},
		// 0 token amount
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
				},
			},
			wantErr,
		},
		// negative token amount
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token: sdk.Coin{
						Denom:  "test1",
						Amount: sdk.NewInt(-1),
					},
				},
			},
			wantErr,
		},
		// total weight 300
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
				},
			},
			noErr,
		},
		// two of the same token
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
				},
				{
					Weight: sdk.NewInt(300),
					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test2", sdk.NewInt(10000)),
				},
			},
			wantErr,
		},
		// total weight 7300
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(200),
					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
				},
				{
					Weight: sdk.NewInt(7000),
					Token:  sdk.NewCoin("test3", sdk.NewInt(10000)),
				},
			},
			noErr,
		},
	}

	var poolId uint64 = 10

	for i, tc := range tests {
		pacc, err := NewBalancerPool(poolId, defaultBalancerPoolParams, tc.assets, defaultFutureGovernor, defaultCurBlockTime)
		if tc.shouldErr {
			require.Error(t, err, "unexpected lack of error, tc %v", i)
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
			expectedTotalWeight := sdk.ZeroInt()
			for i, asset := range tc.assets {
				expectedTotalWeight = expectedTotalWeight.Add(asset.Weight)

				// Ensure pool assets are sorted
				require.Equal(t, "test"+fmt.Sprint(i+1), pacc.PoolAssets[i].Token.Denom)
			}
			testTotalWeight(t, expectedTotalWeight, pacc)
		}
	}
}

// TODO: Figure out what parts of this test, if any, make sense.
func TestGetBalancerPoolAssets(t *testing.T) {
	// Adds []PoolAssets, one after another
	// if the addition doesn't error, adds the weight of the pool assets to a running total,
	// and ensures the pool's total weight is equal to the expected.
	// This also ensures that the pool assets remain sorted within the pool.
	// Furthermore, it ensures that GetPoolAsset succeeds for everything in the pool,
	// and fails for things not in it.
	denomNotInPool := "xyzCoin"

	assets := []PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test4", sdk.NewInt(10000)),
		},
	}

	// TODO: We need way more robust test cases here, and should table drive these cases
	pacc, err := NewBalancerPool(defaultPoolId, defaultBalancerPoolParams, assets, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	// Hardcoded GetPoolAssets tests.
	assets, err = pacc.GetPoolAssets("test1", "test2")
	require.NoError(t, err)
	require.Equal(t, 2, len(assets))

	assets, err = pacc.GetPoolAssets("test1", "test2", "test3", "test4")
	require.NoError(t, err)
	require.Equal(t, 4, len(assets))

	_, err = pacc.GetPoolAssets("test1", "test5")
	require.Error(t, err)
	_, err = pacc.GetPoolAssets(denomNotInPool)
	require.Error(t, err)

	assets, err = pacc.GetPoolAssets()
	require.NoError(t, err)
	require.Equal(t, 0, len(assets))
}

func TestLBPParamsEmptyStartTime(t *testing.T) {
	// Test that when the start time is empty, the pool
	// sets its start time to be the first start time it is called on
	defaultDuration := 100 * time.Second

	initialPoolAssets := []PoolAsset{
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewCoin("asset1", sdk.NewInt(1000)),
		},
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewCoin("asset2", sdk.NewInt(1000)),
		},
	}

	params := SmoothWeightChangeParams{
		Duration: defaultDuration,
		TargetPoolWeights: []PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
			},
			{
				Weight: sdk.NewInt(2),
				Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
			},
		},
	}

	pacc, err := NewBalancerPool(defaultPoolId, PoolParams{
		SmoothWeightChangeParams: &params,
		SwapFee:                  defaultSwapFee,
		ExitFee:                  defaultExitFee,
	}, initialPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	// Consistency check that SmoothWeightChangeParams params are set
	require.NotNil(t, pacc.PoolParams.SmoothWeightChangeParams)
	// Ensure that the start time got set
	require.Equal(t, pacc.PoolParams.SmoothWeightChangeParams.StartTime, defaultCurBlockTime)
}

func TestBalancerPoolPokeTokenWeights(t *testing.T) {
	// Set default date
	defaultStartTime := time.Unix(1618703511, 0)
	defaultStartTimeUnix := defaultStartTime.Unix()
	defaultDuration := 100 * time.Second
	floatGuaranteedPrecision := float64(GuaranteedWeightPrecision)

	// testCases don't need to be ordered by time. but the blockTime should be
	// less than the end time of the SmoothWeightChange. Testing past the end time
	// is already handled.
	type testCase struct {
		blockTime       time.Time
		expectedWeights []sdk.Int
	}

	// Tests how the pool weights get updated via PokeTokenWeights at different block times.
	// The framework underneath will automatically add tests for times before the start time,
	// at the start time, at the end time, and after the end time. It is up to the test writer to
	// test the behavior at times in-between.
	tests := []struct {
		// We take the initial weights from here
		params SmoothWeightChangeParams
		cases  []testCase
	}{
		{
			// 1:1 pool, between asset1 and asset2
			// transitioning to a 1:2 pool
			params: SmoothWeightChangeParams{
				StartTime: defaultStartTime,
				Duration:  defaultDuration,
				InitialPoolWeights: []PoolAsset{
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
				TargetPoolWeights: []PoolAsset{
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(2),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
			},
			cases: []testCase{
				{
					// Halfway through at 50 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+50, 0),
					expectedWeights: []sdk.Int{
						sdk.NewInt(1 * GuaranteedWeightPrecision),
						// Halfway between 1 & 2
						sdk.NewInt(3 * GuaranteedWeightPrecision / 2),
					},
				},
				{
					// Quarter way through at 25 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+25, 0),
					expectedWeights: []sdk.Int{
						sdk.NewInt(1 * GuaranteedWeightPrecision),
						// Quarter way between 1 & 2 = 1.25
						sdk.NewInt(int64(1.25 * floatGuaranteedPrecision)),
					},
				},
			},
		},
		{
			// 2:2 pool, between asset1 and asset2
			// transitioning to a 4:1 pool
			params: SmoothWeightChangeParams{
				StartTime: defaultStartTime,
				Duration:  defaultDuration,
				InitialPoolWeights: []PoolAsset{
					{
						Weight: sdk.NewInt(2),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(2),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
				TargetPoolWeights: []PoolAsset{
					{
						Weight: sdk.NewInt(4),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(1),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
			},
			cases: []testCase{
				{
					// Halfway through at 50 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+50, 0),
					expectedWeights: []sdk.Int{
						// Halfway between 2 & 4
						sdk.NewInt(6 * GuaranteedWeightPrecision / 2),
						// Halfway between 1 & 2
						sdk.NewInt(3 * GuaranteedWeightPrecision / 2),
					},
				},
				{
					// Quarter way through at 25 seconds elapsed
					blockTime: time.Unix(defaultStartTimeUnix+25, 0),
					expectedWeights: []sdk.Int{
						// Quarter way between 2 & 4 = 2.5
						sdk.NewInt(int64(2.5 * floatGuaranteedPrecision)),
						// Quarter way between 2 & 1 = 1.75
						sdk.NewInt(int64(1.75 * floatGuaranteedPrecision)),
					},
				},
			},
		},
	}

	// Add test cases at a time before the start, the start, the end, and a time after the end.
	addDefaultCases := func(params SmoothWeightChangeParams, cases []testCase) []testCase {
		// Set times one second before the start, and one second after the end
		timeBeforeWeightChangeStart := time.Unix(params.StartTime.Unix()-1, 0)
		timeAtWeightChangeEnd := params.StartTime.Add(params.Duration)
		timeAfterWeightChangeEnd := time.Unix(timeAtWeightChangeEnd.Unix()+1, 0)
		initialWeights := make([]sdk.Int, len(params.InitialPoolWeights))
		finalWeights := make([]sdk.Int, len(params.TargetPoolWeights))
		for i, v := range params.InitialPoolWeights {
			initialWeights[i] = v.Weight.MulRaw(GuaranteedWeightPrecision)
		}
		for i, v := range params.TargetPoolWeights {
			// Doesn't need to be scaled, due to this being done already in param initialization,
			// and because params is only shallow copied
			finalWeights[i] = v.Weight
		}
		// Set the test cases for times before the start, and the start
		updatedCases := []testCase{
			{
				blockTime:       timeBeforeWeightChangeStart,
				expectedWeights: initialWeights,
			},
			{
				blockTime:       params.StartTime,
				expectedWeights: initialWeights,
			},
		}
		// Append the provided cases
		updatedCases = append(updatedCases, cases...)
		finalCases := []testCase{
			{
				blockTime:       timeAtWeightChangeEnd,
				expectedWeights: finalWeights,
			},
			{
				blockTime:       timeAfterWeightChangeEnd,
				expectedWeights: finalWeights,
			},
		}
		// Append the final cases
		updatedCases = append(updatedCases, finalCases...)
		return updatedCases
	}

	for poolNum, tc := range tests {
		paramsCopy := tc.params
		// First we create the initial pool assets we will use
		initialPoolAssets := make([]PoolAsset, len(paramsCopy.InitialPoolWeights))
		for i, asset := range paramsCopy.InitialPoolWeights {
			assetCopy := PoolAsset{
				Weight: asset.Weight,
				Token:  sdk.NewInt64Coin(asset.Token.Denom, 10000),
			}
			initialPoolAssets[i] = assetCopy
		}
		// Initialize the pool
		pacc, err := NewBalancerPool(uint64(poolNum), PoolParams{
			SwapFee:                  defaultSwapFee,
			ExitFee:                  defaultExitFee,
			SmoothWeightChangeParams: &tc.params,
		}, initialPoolAssets, defaultFutureGovernor, defaultCurBlockTime)
		require.NoError(t, err, "poolNumber %v", poolNum)

		// Consistency check that SmoothWeightChangeParams params are set
		require.NotNil(t, pacc.PoolParams.SmoothWeightChangeParams)

		testCases := addDefaultCases(paramsCopy, tc.cases)
		for caseNum, testCase := range testCases {
			pacc.PokePool(testCase.blockTime)

			totalWeight := sdk.ZeroInt()

			for assetNum, asset := range pacc.GetAllPoolAssets() {
				require.Equal(t, testCase.expectedWeights[assetNum], asset.Weight,
					"Didn't get the expected weights, poolNumber %v, caseNumber %v, assetNumber %v",
					poolNum, caseNum, assetNum)

				totalWeight = totalWeight.Add(asset.Weight)
			}

			require.Equal(t, totalWeight, pacc.GetTotalWeight())
		}
		// Should have been deleted by the last test case of after PokeTokenWeights pokes past end time.
		require.Nil(t, pacc.PoolParams.SmoothWeightChangeParams)
	}
}

// TestCalculateAmountOutAndIn_InverseRelationship tests that the same amount of token is guaranteed upon
// sequential operation of CalcInAmtGivenOut and CalcOutAmtGivenIn.
func TestCalculateAmountOutAndIn_InverseRelationship(t *testing.T) {
	type testcase struct {
		denomOut         string
		initialPoolOut   int64
		initialWeightOut int64
		initialCalcOut   int64

		denomIn         string
		initialPoolIn   int64
		initialWeightIn int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000_000_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 100,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 100,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000,
			initialWeightOut: 200,
			initialCalcOut:   100,

			denomIn:         "ion",
			initialPoolIn:   1_000_000,
			initialWeightIn: 50,
		},
		{
			denomOut:         "uosmo",
			initialPoolOut:   1_000_000,
			initialWeightOut: 200,
			initialCalcOut:   100000,

			denomIn:         "ion",
			initialPoolIn:   1_000_000_000,
			initialWeightIn: 50,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("tokenOutInitial: %d, tokenInInitial: %d, initialOut: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.initialPoolIn,
			tc.initialCalcOut,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				ctx := createTestContext(t)

				poolAssetOut := PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomOut, tc.initialPoolOut),
					Weight: sdk.NewInt(tc.initialWeightOut),
				}

				poolAssetIn := PoolAsset{
					Token:  sdk.NewInt64Coin(tc.denomIn, tc.initialPoolIn),
					Weight: sdk.NewInt(tc.initialWeightIn),
				}

				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				exitFeeDec, err := sdk.NewDecFromStr("0")
				require.NoError(t, err)

				pool := createTestPool(t, []PoolAsset{
					poolAssetOut,
					poolAssetIn,
				},
					swapFeeDec,
					exitFeeDec,
				)
				require.NotNil(t, pool)

				initialOut := sdk.NewInt64Coin(poolAssetOut.Token.Denom, tc.initialCalcOut)
				initialOutCoins := sdk.NewCoins(initialOut)

				actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, poolAssetIn.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), poolAssetOut.Token.Denom, swapFeeDec)
				require.NoError(t, err)

				require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

				expected := initialOut.Amount.ToDec()
				actual := inverseTokenOut.Amount.ToDec()

				// allow a rounding error of up to 1 for this relation
				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, expected, actual, tol))
			})
		}
	}
}

func TestCalcSingleAssetInAndOut_InverseRelationship(t *testing.T) {
	type testcase struct {
		initialPoolOut   int64
		initialPoolIn    int64
		initialWeightOut int64
		tokenOut         int64
		initialWeightIn  int64
	}

	// For every test case in testcases, apply a swap fee in swapFeeCases.
	testcases := []testcase{
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 50,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         50,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000_000_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  50,
		},
		{
			initialPoolOut:   1_000_000,
			tokenOut:         100,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   2_351_333,
			tokenOut:         7,
			initialWeightOut: 148,
			initialWeightIn:  57,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         25,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
		{
			initialPoolOut:   1_000,
			tokenOut:         26,
			initialWeightOut: 100,
			initialWeightIn:  100,
		},
	}

	swapFeeCases := []string{"0", "0.001", "0.1", "0.5", "0.99"}

	getTestCaseName := func(tc testcase, swapFeeCase string) string {
		return fmt.Sprintf("initialPoolOut: %d, initialCalcOut: %d, initialWeightOut: %d, initialWeightIn: %d, swapFee: %s",
			tc.initialPoolOut,
			tc.tokenOut,
			tc.initialWeightOut,
			tc.initialWeightIn,
			swapFeeCase,
		)
	}

	for _, tc := range testcases {
		for _, swapFee := range swapFeeCases {
			t.Run(getTestCaseName(tc, swapFee), func(t *testing.T) {
				swapFeeDec, err := sdk.NewDecFromStr(swapFee)
				require.NoError(t, err)

				initialPoolBalanceOut := sdk.NewInt(tc.initialPoolOut)

				initialWeightOut := sdk.NewInt(tc.initialWeightOut)
				initialWeightIn := sdk.NewInt(tc.initialWeightIn)

				initialTotalShares := types.InitPoolSharesSupply.ToDec()
				initialCalcTokenOut := sdk.NewInt(tc.tokenOut)

				actualSharesOut := CalcPoolSharesOutGivenSingleAssetIn(
					initialPoolBalanceOut.ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares,
					initialCalcTokenOut.ToDec(),
					swapFeeDec,
				)

				inverseCalcTokenOut := CalcSingleAssetInGivenPoolSharesOut(
					initialPoolBalanceOut.Add(initialCalcTokenOut).ToDec(),
					initialWeightOut.ToDec().Quo(initialWeightOut.Add(initialWeightIn).ToDec()),
					initialTotalShares.Add(actualSharesOut),
					actualSharesOut,
					swapFeeDec,
				)

				tol := sdk.NewDec(1)
				require.True(osmoutils.DecApproxEq(t, initialCalcTokenOut.ToDec(), inverseCalcTokenOut, tol))
			})
		}
	}
}
