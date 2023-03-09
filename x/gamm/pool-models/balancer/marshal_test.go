package balancer_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ymlAssetTest = []balancer.PoolAsset{
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

	jsonAssetTest := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	}
	pacc, err := balancer.NewBalancerPool(poolId, balancer.PoolParams{
		SwapFee: defaultSwapFee,
	}, jsonAssetTest, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	bz, err := json.Marshal(pacc)
	require.NoError(t, err)

	bz1, err := pacc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a balancer.Pool
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, pacc.String(), a.String())
}

func TestPoolProtoMarshal(t *testing.T) {
	// since we decided to remove ExitFee, this hex string should be changed
	decodedByteArray, err := hex.DecodeString("100a1a130a1132353030303030303030303030303030302a110a0c67616d6d2f706f6f6c2f3130120130321e0a0e0a05746573743112053130303030120c313037333734313832343030321e0a0e0a05746573743212053530303030120c3231343734383336343830303a0130")
	require.NoError(t, err)

	pool2 := balancer.Pool{}
	err = proto.Unmarshal(decodedByteArray, &pool2)
	require.NoError(t, err)

	require.Equal(t, pool2.Id, uint64(10))
	require.Equal(t, pool2.PoolParams.SwapFee, defaultSwapFee)
	require.Equal(t, pool2.FuturePoolGovernor, "")
	require.Equal(t, pool2.TotalShares, sdk.Coin{Denom: "gamm/pool/10", Amount: sdk.ZeroInt()})
	require.Equal(t, pool2.PoolAssets, []balancer.PoolAsset{
		{
			Token: sdk.Coin{
				Denom:  "test1",
				Amount: sdk.NewInt(10000),
			},
			Weight: sdk.NewInt(107374182400),
		},
		{
			Token: sdk.Coin{
				Denom:  "test2",
				Amount: sdk.NewInt(50000),
			},
			Weight: sdk.NewInt(214748364800),
		},
	})
}
