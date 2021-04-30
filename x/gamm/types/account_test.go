package types

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultSwapFee    = sdk.MustNewDecFromStr("0.025")
	defaultExitFee    = sdk.MustNewDecFromStr("0.025")
	defaultPoolParams = PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultFutureGovernor = ""
	//
	dummyPoolAssets = []PoolAsset{}
	wantErr         = true
	noErr           = false
)

// Expected is un-scaled
func testTotalWeight(t *testing.T, expected sdk.Int, pool PoolAccountI) {
	scaledExpected := expected.MulRaw(GuaranteedWeightPrecision)
	require.Equal(t,
		scaledExpected.String(),
		pool.GetTotalWeight().String())
}

func TestPoolAccountShareDenom(t *testing.T) {
	var poolId uint64 = 10

	pacc, err := NewPoolAccount(poolId, defaultPoolParams, dummyPoolAssets, defaultFutureGovernor)
	require.NoError(t, err)

	require.Equal(t, "gamm/pool/10", pacc.GetTotalShare().Denom)
}

func TestPoolAccountPoolParams(t *testing.T) {
	// Tests that creating a pool with the given pair of swapfee and exit fee
	// errors or succeeds as intended. Furthermore, it checks that
	// NewPoolAccount panics in the error case.
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
		poolParams := PoolParams{
			SwapFee: params.SwapFee,
			ExitFee: params.ExitFee,
		}
		err := poolParams.Validate(dummyPoolAssets)
		if params.shouldErr {
			require.Error(t, err, "unexpected lack of error, tc %v", i)
			// Check that these are also caught if passed to the underlying pool creation func
			_, err = NewPoolAccount(1, poolParams, dummyPoolAssets, defaultFutureGovernor)
			require.Error(t, err)
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
		}
	}
}

// TODO: Refactor this into multiple tests
func TestPoolAccountUpdatePoolAssetBalance(t *testing.T) {
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

	pacc, err := NewPoolAccount(poolId, defaultPoolParams, initialAssets, defaultFutureGovernor)
	require.NoError(t, err)

	_, err = pacc.GetPoolAsset("unknown")
	require.Error(t, err)
	_, err = pacc.GetPoolAsset("")
	require.Error(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc)

	// Break abstractions and start reasoning about the underlying internal representation's APIs.
	// TODO: This test actually just needs to be refactored to not be doing this, and just
	// create a different pool each time.
	pacc_internal := pacc.(*PoolAccount)

	err = pacc_internal.setInitialPoolAssets([]PoolAsset{PoolAsset{
		Weight: sdk.NewInt(-1),
		Token:  sdk.NewCoin("negativeWeight", sdk.NewInt(50000)),
	}})

	require.Error(t, err)

	err = pacc_internal.setInitialPoolAssets([]PoolAsset{PoolAsset{
		Weight: sdk.NewInt(0),
		Token:  sdk.NewCoin("zeroWeight", sdk.NewInt(50000)),
	}})
	require.Error(t, err)

	err = pacc_internal.UpdatePoolAssetBalance(
		sdk.NewCoin("test1", sdk.NewInt(0)))
	require.Error(t, err)

	err = pacc_internal.UpdatePoolAssetBalance(
		sdk.Coin{Denom: "test1", Amount: sdk.NewInt(-1)},
	)
	require.Error(t, err)

	err = pacc_internal.UpdatePoolAssetBalance(
		sdk.NewCoin("test1", sdk.NewInt(1)))
	require.NoError(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc_internal)

	PoolAsset, err := pacc_internal.GetPoolAsset("test1")
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(1).String(), PoolAsset.Token.Amount.String())
}

func TestPoolAccountPoolAssetsWeightAndTokenBalance(t *testing.T) {
	// TODO: Add more cases
	tests := []struct {
		assets      []PoolAsset
		totalWeight int64
		shouldErr   bool
	}{
		// weight 0
		{
			[]PoolAsset{
				{
					Weight: sdk.NewInt(0),
					Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
				},
			},
			0,
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
			-1,
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
			100,
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
			100,
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
			300,
			noErr,
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
			7300,
			noErr,
		},
	}

	var poolId uint64 = 10

	for i, tc := range tests {
		pacc, err := NewPoolAccount(poolId, defaultPoolParams, tc.assets, defaultFutureGovernor)
		if tc.shouldErr {
			require.Error(t, err, "unexpected lack of error, tc %v", i)
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
			testTotalWeight(t, sdk.NewInt(tc.totalWeight), pacc)
		}
	}
}

// TODO: Figure out what parts of this test, if any, make sense.
// func TestPoolAccountPoolAssets(t *testing.T) {
// 	// Adds []PoolAssets, one after another
// 	// if the addition doesn't error, adds the weight of the pool assets to a running total,
// 	// and ensures the pool's total weight is equal to the expected.
// 	// This also ensures that the pool assets remain sorted within the pool account.
// 	// Furthermore, it ensures that GetPoolAsset succeeds for everything in the pool,
// 	// and fails for things not in it.
// 	denomNotInPool := "xyzCoin"

// 	tests := []struct {
// 		assets         []PoolAsset
// 		newAssetsAdded int
// 		shouldErr      bool
// 	}{
// 		{
// 			[]PoolAsset{
// 				{
// 					Weight: sdk.NewInt(200),
// 					Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
// 				},
// 				{
// 					Weight: sdk.NewInt(100),
// 					Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
// 				},
// 			},
// 			2,
// 			noErr,
// 		},
// 		{
// 			[]PoolAsset{
// 				{
// 					Weight: sdk.NewInt(200),
// 					Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
// 				},
// 				{
// 					Weight: sdk.NewInt(100),
// 					Token:  sdk.NewCoin("test3", sdk.NewInt(10000)),
// 				},
// 			},
// 			0,
// 			wantErr,
// 		},
// 		{
// 			[]PoolAsset{
// 				{
// 					Weight: sdk.NewInt(200),
// 					Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
// 				},
// 				{
// 					Weight: sdk.NewInt(100),
// 					Token:  sdk.NewCoin("test3", sdk.NewInt(10000)),
// 				},
// 			},
// 			0,
// 			wantErr,
// 		},
// 		{
// 			[]PoolAsset{
// 				{
// 					Weight: sdk.NewInt(200),
// 					Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
// 				},
// 				{
// 					Weight: sdk.NewInt(100),
// 					Token:  sdk.NewCoin("test4", sdk.NewInt(10000)),
// 				},
// 			},
// 			2,
// 			noErr,
// 		},
// 	}

// 	expectedTotalWeight := sdk.ZeroInt()
// 	expectedNumAssets := 0
// 	var poolId uint64 = 10
// 	pacc, err := NewPoolAccount(poolId, defaultPoolParams, nil, defaultFutureGovernor)
// 	pacc_internal := pacc.(*PoolAccount)

// 	// Just check that theres no asset called test1 at the start.
// 	_, err := pacc.GetPoolAsset("test1")
// 	require.Error(t, err)

// 	for i, tc := range tests {
// 		err = pacc.AddPoolAssets(tc.assets)
// 		if tc.shouldErr {
// 			require.Error(t, err, "unexpected lack of error, tc %v", i)
// 		} else {
// 			require.NoError(t, err, "unexpected error, tc %v", i)
// 			// Check that the number of assets in the pool is correct
// 			expectedNumAssets += len(tc.assets)
// 			require.Equal(t, expectedNumAssets, pacc.NumAssets())
// 			// Check that the total weight is correct
// 			for _, asset := range tc.assets {
// 				expectedTotalWeight = expectedTotalWeight.Add(asset.Weight)
// 			}
// 			testTotalWeight(t, expectedTotalWeight, pacc)
// 			// Check that the assets in the pool are sorted by denomination
// 			// TODO: The following is just left as a stub
// 			require.Equal(t, "test1", pacc.PoolAssets[0].Token.Denom)
// 			require.Equal(t, "test2", pacc.PoolAssets[1].Token.Denom)
// 		}
// 		// Check that GetPoolAsset works for every denom in pool
// 		for _, asset := range pacc.PoolAssets {
// 			_, err = pacc.GetPoolAsset(asset.Token.Denom)
// 			require.NoError(t, err)
// 		}
// 		// Check that GetPoolAsset fails for a denom not in the pool
// 		_, err = pacc.GetPoolAsset(denomNotInPool)
// 		require.Error(t, err)
// 	}

// 	// Hardcoded GetPoolAssets tests.
// 	// TODO: Find ways to generalize these.
// 	assets, err := pacc.GetPoolAssets("test1", "test2")
// 	require.NoError(t, err)
// 	require.Equal(t, 2, len(assets))

// 	assets, err = pacc.GetPoolAssets("test1", "test2", "test3", "test4")
// 	require.NoError(t, err)
// 	require.Equal(t, 4, len(assets))

// 	_, err = pacc.GetPoolAssets("test1", "test5")
// 	require.Error(t, err)
// 	_, err = pacc.GetPoolAssets("test5")
// 	require.Error(t, err)

// 	assets, err = pacc.GetPoolAssets()
// 	require.NoError(t, err)
// 	require.Equal(t, 0, len(assets))
// }

func TestPoolAccountPokeTokenWeights(t *testing.T) {
	// Set default date
	defaultStartTime := time.Unix(1618703511, 0)
	defaultStartTimeUnix := defaultStartTime.Unix()
	defaultDuration := 100 * time.Second
	floatGuaranteedPrecison := float64(GuaranteedWeightPrecision)

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
						Weight: sdk.NewInt(1 * GuaranteedWeightPrecision),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(1 * GuaranteedWeightPrecision),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
				TargetPoolWeights: []PoolAsset{
					{
						Weight: sdk.NewInt(1 * GuaranteedWeightPrecision),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(2 * GuaranteedWeightPrecision),
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
						sdk.NewInt(int64(1.25 * floatGuaranteedPrecison)),
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
						Weight: sdk.NewInt(2 * GuaranteedWeightPrecision),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(2 * GuaranteedWeightPrecision),
						Token:  sdk.NewCoin("asset2", sdk.NewInt(0)),
					},
				},
				TargetPoolWeights: []PoolAsset{
					{
						Weight: sdk.NewInt(4 * GuaranteedWeightPrecision),
						Token:  sdk.NewCoin("asset1", sdk.NewInt(0)),
					},
					{
						Weight: sdk.NewInt(1 * GuaranteedWeightPrecision),
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
						sdk.NewInt(int64(2.5 * floatGuaranteedPrecison)),
						// Quarter way between 2 & 1 = 1.75
						sdk.NewInt(int64(1.75 * floatGuaranteedPrecison)),
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
			initialWeights[i] = v.Weight
		}
		for i, v := range params.TargetPoolWeights {
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
				Weight: asset.Weight.QuoRaw(GuaranteedWeightPrecision),
				Token:  sdk.NewInt64Coin(asset.Token.Denom, 10000),
			}
			initialPoolAssets[i] = assetCopy
			fmt.Printf("%v\n", initialPoolAssets[i])
		}
		// Initialize the pool
		pacc, err := NewPoolAccount(uint64(poolNum), PoolParams{
			SwapFee:                  defaultSwapFee,
			ExitFee:                  defaultExitFee,
			SmoothWeightChangeParams: &tc.params,
		}, initialPoolAssets, defaultFutureGovernor)
		require.NoError(t, err, "poolNumber %v", poolNum)
		// Consistency check that SmoothWeightChangeParams params are set
		require.NotNil(t, pacc.GetPoolParams().SmoothWeightChangeParams)

		testCases := addDefaultCases(paramsCopy, tc.cases)
		for caseNum, testCase := range testCases {
			pacc.PokeTokenWeights(testCase.blockTime)
			for assetNum, asset := range pacc.GetAllPoolAssets() {
				require.Equal(t, testCase.expectedWeights[assetNum], asset.Weight,
					"Didn't get the expected weights, poolNumber %v, caseNumber %v, assetNumber %v",
					poolNum, caseNum, assetNum)
			}
		}
		// Should have been deleted by the last test case of after PokeTokenWeights pokes past end time.
		// TODO: This doesn't work due to PokeTokenWeights having a non-pointer receiver =/
		// require.Nil(t, pacc.GetPoolParams().SmoothWeightChangeParams)
	}

}

func TestPoolAccountMarshalYAML(t *testing.T) {
	var poolId uint64 = 10

	ymlAssetTest := []PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	}
	pacc, err := NewPoolAccount(poolId, PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}, ymlAssetTest, defaultFutureGovernor)
	require.NoError(t, err)

	bs, err := yaml.Marshal(pacc)
	require.NoError(t, err)

	want := `|
  address: cosmos1m48tfmd0e6yqgfhraxl9ddt7lygpsnsrhtwpas
  public_key: ""
  account_number: 0
  sequence: 0
  id: 10
  pool_params:
    swap_fee: "0.025000000000000000"
    exit_fee: "0.025000000000000000"
    smooth_weight_change_params: null
  future_pool_governor: ""
  total_weight: "322122547200"
  total_share:
    denom: gamm/pool/10
    amount: "0"
  pool_assets:
  - token:
      denom: test1
      amount: "10000"
    weight: "107374182400"
  - token:
      denom: test2
      amount: "50000"
    weight: "214748364800"
`
	require.Equal(t, want, string(bs))
}

func TestPoolAccountJson(t *testing.T) {
	var poolId uint64 = 10

	jsonAssetTest := []PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	}
	pacc, err := NewPoolAccount(poolId, PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}, jsonAssetTest, defaultFutureGovernor)
	require.NoError(t, err)

	paccInternal := pacc.(*PoolAccount)

	bz, err := json.Marshal(pacc)
	require.NoError(t, err)

	bz1, err := paccInternal.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a PoolAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, pacc.String(), a.String())
}
