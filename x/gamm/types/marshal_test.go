package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appParams "github.com/osmosis-labs/osmosis/app/params"
)

var ymlAssetTest = []PoolAsset{
	{
		Weight: sdk.NewInt(200),
		Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
	},
	{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
	},
}

func TestPoolMarshalYAML(t *testing.T) {
	appParams.SetAddressPrefixes()
	pacc, err := NewPool(defaultPoolId, PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}, ymlAssetTest, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	bs, err := yaml.Marshal(pacc)
	require.NoError(t, err)

	want := `|
  address: osmo1krp38zzc3zz5as9ndqkyskhkzv6x9e30ckcq5g4lcsu5wpwcqy0sa3dea2
  id: 10
  pool_params:
    swap_fee: "0.025000000000000000"
    exit_fee: "0.025000000000000000"
    smooth_weight_change_params: null
  future_pool_governor: ""
  total_weight: "300.000000000000000000"
  total_shares:
    denom: gamm/pool/10
    amount: "0"
  pool_assets:
  - |
    token:
      denom: test1
      amount: "10000"
    weight: "100.000000000000000000"
  - |
    token:
      denom: test2
      amount: "50000"
    weight: "200.000000000000000000"
`
	require.Equal(t, want, string(bs))
}

func TestLBPPoolMarshalYAML(t *testing.T) {
	appParams.SetAddressPrefixes()
	lbpParams := SmoothWeightChangeParams{
		Duration: time.Hour,
		TargetPoolWeights: []PoolAsset{
			{
				Weight: sdk.NewInt(300),
				Token:  sdk.NewCoin("test2", sdk.NewInt(0)),
			},
			{
				Weight: sdk.NewInt(700),
				Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
			},
		},
	}
	pacc, err := NewPool(defaultPoolId, PoolParams{
		SwapFee:                  defaultSwapFee,
		ExitFee:                  defaultExitFee,
		SmoothWeightChangeParams: &lbpParams,
	}, ymlAssetTest, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	bs, err := yaml.Marshal(pacc)
	require.NoError(t, err)

	expectedStartTimeBz, err := yaml.Marshal(defaultCurBlockTime)
	expectedStartTimeString := strings.Trim(string(expectedStartTimeBz), "\n")
	require.NoError(t, err)

	want := fmt.Sprintf(`|
  address: osmo1krp38zzc3zz5as9ndqkyskhkzv6x9e30ckcq5g4lcsu5wpwcqy0sa3dea2
  id: 10
  pool_params:
    swap_fee: "0.025000000000000000"
    exit_fee: "0.025000000000000000"
    smooth_weight_change_params:
      start_time: %s
      duration: 1h0m0s
      initial_pool_weights:
      - |
        token:
          denom: test1
          amount: "0"
        weight: "100.000000000000000000"
      - |
        token:
          denom: test2
          amount: "0"
        weight: "200.000000000000000000"
      target_pool_weights:
      - |
        token:
          denom: test1
          amount: "0"
        weight: "700.000000000000000000"
      - |
        token:
          denom: test2
          amount: "0"
        weight: "300.000000000000000000"
  future_pool_governor: ""
  total_weight: "300.000000000000000000"
  total_shares:
    denom: gamm/pool/10
    amount: "0"
  pool_assets:
  - |
    token:
      denom: test1
      amount: "10000"
    weight: "100.000000000000000000"
  - |
    token:
      denom: test2
      amount: "50000"
    weight: "200.000000000000000000"
`, expectedStartTimeString)
	require.Equal(t, want, string(bs))
}

func TestPoolJson(t *testing.T) {
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
	pacc, err := NewPool(poolId, PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}, jsonAssetTest, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	paccInternal := pacc.(*Pool)

	bz, err := json.Marshal(pacc)
	require.NoError(t, err)

	bz1, err := paccInternal.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a Pool
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, pacc.String(), a.String())
}
