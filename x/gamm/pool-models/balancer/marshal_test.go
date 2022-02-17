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

func TestPoolProtoUnmarshal(t *testing.T) {

	// hex of serialzied poolI from v6.x
	decodedByteArray, err := hex.DecodeString("0a3f6f736d6f316b727033387a7a63337a7a356173396e64716b79736b686b7a76367839653330636b63713567346c637375357770776371793073613364656132100a1a260a113235303030303030303030303030303030121132353030303030303030303030303030302a110a0c67616d6d2f706f6f6c2f3130120130321e0a0e0a05746573743112053130303030120c313037333734313832343030321e0a0e0a05746573743212053530303030120c3231343734383336343830303a0c333232313232353437323030")
	require.NoError(t, err)

	pool2 := BalancerPool{}
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

func TestPoolJsonUnmarhsal(t *testing.T) {

	// hex of serialzied poolI from v6.x
	decodedByteArray, err := hex.DecodeString("7b2261646472657373223a226f736d6f316b727033387a7a63337a7a356173396e64716b79736b686b7a76367839653330636b63713567346c637375357770776371793073613364656132222c226964223a31302c22706f6f6c5f706172616d73223a7b2273776170466565223a22302e303235303030303030303030303030303030222c2265786974466565223a22302e303235303030303030303030303030303030227d2c226675747572655f706f6f6c5f676f7665726e6f72223a22222c22746f74616c5f776569676874223a223332323132323534373230302e303030303030303030303030303030303030222c22746f74616c5f736861726573223a7b2264656e6f6d223a2267616d6d2f706f6f6c2f3130222c22616d6f756e74223a2230227d2c22706f6f6c5f617373657473223a5b7b22746f6b656e223a7b2264656e6f6d223a227465737431222c22616d6f756e74223a223130303030227d2c22776569676874223a22313037333734313832343030227d2c7b22746f6b656e223a7b2264656e6f6d223a227465737432222c22616d6f756e74223a223530303030227d2c22776569676874223a22323134373438333634383030227d5d7d")
	require.NoError(t, err)

	pool2 := BalancerPool{}
	err = json.Unmarshal(decodedByteArray, &pool2)
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
