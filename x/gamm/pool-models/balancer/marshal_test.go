package balancer

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
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
	pacc, err := NewBalancerPool(poolId, PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}, jsonAssetTest, defaultFutureGovernor, defaultCurBlockTime)
	require.NoError(t, err)

	bz, err := json.Marshal(pacc)
	require.NoError(t, err)

	bz1, err := pacc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a Pool
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, pacc.String(), a.String())
}

func TestPoolProtoMarshal(t *testing.T) {

	// hex of serialzied poolI from v6.x
	decodedByteArray, err := hex.DecodeString("0a3f6f736d6f316b727033387a7a63337a7a356173396e64716b79736b686b7a76367839653330636b63713567346c637375357770776371793073613364656132100a1a260a113235303030303030303030303030303030121132353030303030303030303030303030302a110a0c67616d6d2f706f6f6c2f3130120130321e0a0e0a05746573743112053130303030120c313037333734313832343030321e0a0e0a05746573743212053530303030120c3231343734383336343830303a0c333232313232353437323030")
	require.NoError(t, err)

	pool2 := Pool{}
	err = proto.Unmarshal(decodedByteArray, &pool2)
	require.NoError(t, err)

	require.Equal(t, pool2.Id, uint64(10))
	require.Equal(t, pool2.PoolParams.SwapFee, defaultSwapFee)
	require.Equal(t, pool2.PoolParams.ExitFee, defaultExitFee)
	require.Equal(t, pool2.FuturePoolGovernor, "")
	require.Equal(t, pool2.TotalShares, sdk.Coin{Denom: "gamm/pool/10", Amount: sdk.ZeroInt()})
	require.Equal(t, pool2.PoolAssets, []types.PoolAsset{
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
