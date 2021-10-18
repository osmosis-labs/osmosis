package balancer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

var ymlAssetTest = []types.PoolAsset{
	{
		Weight: sdk.NewInt(200),
		Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
	},
	{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
	},
}

func TestPoolJson(t *testing.T) {
	var poolId uint64 = 10

	jsonAssetTest := []types.PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	}
	pacc, err := NewBalancerPool(poolId, BalancerPoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}, jsonAssetTest, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	bz, err := json.Marshal(pacc)
	require.NoError(t, err)

	bz1, err := pacc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a BalancerPool
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, pacc.String(), a.String())
}
