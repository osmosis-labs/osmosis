package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testTotalWeight(t *testing.T, expected sdk.Int, pool PoolAccountI) {
	require.Equal(t, expected.String(), pool.GetTotalWeight().String())
}

func TestPoolAccountPoolParams(t *testing.T) {
	defaultSwapFee, _ := sdk.NewDecFromStr("0.025")
	defaultExitFee, _ := sdk.NewDecFromStr("0.025")
	wantErr := true
	noErr := false

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
			Lock:    false,
			SwapFee: params.SwapFee,
			ExitFee: params.ExitFee,
		}
		err := poolParams.Validate()
		if params.shouldErr {
			require.Error(t, err, "unexpected error, tc %v", i)
			require.Panics(t, func() { NewPoolAccount(1, poolParams, "") })
		} else {
			require.NoError(t, err, "unexpected error, tc %v", i)
		}
	}
}

func TestPoolAccountSetPoolAsset(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, "")

	err := pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
	})
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("unknown")
	require.Error(t, err)
	_, err = pacc.GetPoolAsset("")
	require.Error(t, err)

	testTotalWeight(t, sdk.NewInt(300), pacc)

	err = pacc.SetPoolAsset("test1", PoolAsset{
		Weight: sdk.NewInt(-1),
		Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
	})
	require.Error(t, err)

	err = pacc.SetPoolAsset("test1", PoolAsset{
		Weight: sdk.NewInt(0),
		Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
	})
	require.Error(t, err)

	err = pacc.SetPoolAsset("test1", PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
	})
	require.Error(t, err)

	err = pacc.SetPoolAsset("test1", PoolAsset{
		Weight: sdk.NewInt(100),
		Token: sdk.Coin{
			Denom:  "test1",
			Amount: sdk.NewInt(-1),
		},
	})
	require.Error(t, err)

	err = pacc.SetPoolAsset("test1", PoolAsset{
		Weight: sdk.NewInt(200),
		Token: sdk.Coin{
			Denom:  "test1",
			Amount: sdk.NewInt(1),
		},
	})
	require.NoError(t, err)

	require.Equal(t, sdk.NewInt(400).String(), pacc.GetTotalWeight().String())

	PoolAsset, err := pacc.GetPoolAsset("test1")
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(1).String(), PoolAsset.Token.Amount.String())
}

func TestPoolAccountPoolAssetsWeightAndTokenBalance(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, "")

	err := pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(0),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
	})
	require.Error(t, err)

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(-1),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
	})
	require.Error(t, err)

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
		},
	})
	require.Error(t, err)

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token: sdk.Coin{
				Denom:  "test1",
				Amount: sdk.NewInt(-1),
			},
		},
	})
	require.Error(t, err)

	require.Equal(t, 0, pacc.NumAssets())
}

func TestPoolAccountPoolAssets(t *testing.T) {
	// TODO: Refactor this to be table driven
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, "").(*PoolAccount)

	_, err := pacc.GetPoolAsset("test1")
	require.Error(t, err)

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	})
	require.NoError(t, err)

	require.Equal(t, 2, len(pacc.PoolAssets))
	require.Equal(t, 2, pacc.NumAssets())
	// Check that PoolAssets are sorted.
	require.Equal(t, "test1", pacc.PoolAssets[0].Token.Denom)
	require.Equal(t, "test2", pacc.PoolAssets[1].Token.Denom)

	_, err = pacc.GetPoolAsset("test1")
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("test2")
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("test3")
	require.Error(t, err)

	PoolAssets, err := pacc.GetPoolAssets("test1", "test2")
	require.NoError(t, err)
	require.Equal(t, 2, len(PoolAssets))

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test3", sdk.NewInt(10000)),
		},
	})
	require.Error(t, err)

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test3", sdk.NewInt(10000)),
		},
	})
	require.Error(t, err)

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test4", sdk.NewInt(10000)),
		},
	})
	require.NoError(t, err)

	require.Equal(t, 4, len(pacc.PoolAssets))
	require.Equal(t, 4, pacc.NumAssets())
	// Check that PoolAssets are sorted.
	require.Equal(t, "test1", pacc.PoolAssets[0].Token.Denom)
	require.Equal(t, "test2", pacc.PoolAssets[1].Token.Denom)
	require.Equal(t, "test3", pacc.PoolAssets[2].Token.Denom)
	require.Equal(t, "test4", pacc.PoolAssets[3].Token.Denom)

	_, err = pacc.GetPoolAsset("test1")
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("test2")
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("test3")
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("test4")
	require.NoError(t, err)
	_, err = pacc.GetPoolAsset("test5")
	require.Error(t, err)

	PoolAssets, err = pacc.GetPoolAssets("test1", "test2", "test3", "test4")
	require.NoError(t, err)
	require.Equal(t, 4, len(PoolAssets))

	_, err = pacc.GetPoolAssets("test1", "test5")
	require.Error(t, err)
	_, err = pacc.GetPoolAssets("test5")
	require.Error(t, err)

	PoolAssets, err = pacc.GetPoolAssets()
	require.NoError(t, err)
	require.Equal(t, 0, len(PoolAssets))
}

func TestPoolAccountTotalWeight(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, "")

	err := pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	})
	require.NoError(t, err)

	require.Equal(t, sdk.NewInt(300).String(), pacc.GetTotalWeight().String())

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test2", sdk.NewInt(10000)),
		},
	})
	require.Error(t, err)

	require.Equal(t, sdk.NewInt(300).String(), pacc.GetTotalWeight().String())

	err = pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(1),
			Token:  sdk.NewCoin("test3", sdk.NewInt(50000)),
		},
	})
	require.NoError(t, err)

	require.Equal(t, sdk.NewInt(301).String(), pacc.GetTotalWeight().String())
}

func TestPoolAccountMarshalYAML(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, "")

	err := pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	})
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
    lock: false
    swap_fee: "0.025000000000000000"
    exit_fee: "0.025000000000000000"
  total_weight: "300"
  total_share:
    denom: osmosis/pool/10
    amount: "0"
  pool_assets:
  - token:
      denom: test1
      amount: "10000"
    weight: "100"
  - token:
      denom: test2
      amount: "50000"
    weight: "200"
`
	require.Equal(t, want, string(bs))
}

func TestPoolAccountJson(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}, "").(*PoolAccount)

	err := pacc.AddPoolAssets([]PoolAsset{
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("test2", sdk.NewInt(50000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(10000)),
		},
	})
	require.NoError(t, err)

	bz, err := json.Marshal(pacc)
	require.NoError(t, err)

	bz1, err := pacc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a PoolAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, pacc.String(), a.String())
}
