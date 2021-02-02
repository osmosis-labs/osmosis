package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPoolAccountPoolParams(t *testing.T) {
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	require.Error(t, PoolParams{
		Lock: false,
		// Can't set the swap fee as negative
		SwapFee: sdk.NewDecWithPrec(-1, 2),
		ExitFee: exitFee,
	}.Validate())

	require.Error(t, PoolParams{
		Lock: false,
		// Can't set the swap fee as 1
		SwapFee: sdk.NewDec(1),
		ExitFee: exitFee,
	}.Validate())

	require.Error(t, PoolParams{
		Lock: false,
		// Can't set the swap fee above 1
		SwapFee: sdk.NewDecWithPrec(15, 1),
		ExitFee: exitFee,
	}.Validate())

	require.Error(t, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		// Can't set the exit fee as negative
		ExitFee: sdk.NewDecWithPrec(-1, 2),
	}.Validate())

	require.Error(t, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		// Can't set the exit fee as 1
		ExitFee: sdk.NewDec(1),
	}.Validate())

	require.Error(t, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		// Can't set the exit fee above 1
		ExitFee: sdk.NewDecWithPrec(15, 1),
	}.Validate())

	require.Panics(t, func() {
		// Can't create with negative swap fee.
		NewPoolAccount(1, PoolParams{
			Lock:    false,
			SwapFee: sdk.NewDecWithPrec(-1, 2),
			ExitFee: exitFee,
		})
	})

	require.Panics(t, func() {
		// Can't create with negative exit fee.
		NewPoolAccount(1, PoolParams{
			Lock:    false,
			SwapFee: swapFee,
			ExitFee: sdk.NewDecWithPrec(-1, 2),
		})
	})
}

func TestPoolAccountSetRecord(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	})

	err := pacc.AddRecords([]Record{
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
	_, err = pacc.GetRecord("unknown")
	require.Error(t, err)
	_, err = pacc.GetRecord("")
	require.Error(t, err)

	require.Equal(t, sdk.NewInt(300).String(), pacc.GetTotalWeight().String())

	err = pacc.SetRecord("test1", Record{
		Weight: sdk.NewInt(-1),
		Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
	})
	require.Error(t, err)

	err = pacc.SetRecord("test1", Record{
		Weight: sdk.NewInt(0),
		Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
	})
	require.Error(t, err)

	err = pacc.SetRecord("test1", Record{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
	})
	require.Error(t, err)

	err = pacc.SetRecord("test1", Record{
		Weight: sdk.NewInt(100),
		Token: sdk.Coin{
			Denom:  "test1",
			Amount: sdk.NewInt(-1),
		},
	})
	require.Error(t, err)

	err = pacc.SetRecord("test1", Record{
		Weight: sdk.NewInt(200),
		Token: sdk.Coin{
			Denom:  "test1",
			Amount: sdk.NewInt(1),
		},
	})
	require.NoError(t, err)

	require.Equal(t, sdk.NewInt(400).String(), pacc.GetTotalWeight().String())

	record, err := pacc.GetRecord("test1")
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(1).String(), record.Token.Amount.String())
}

func TestPoolAccountRecordsWeightAndTokenBalance(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	})

	err := pacc.AddRecords([]Record{
		{
			Weight: sdk.NewInt(0),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
	})
	require.Error(t, err)

	err = pacc.AddRecords([]Record{
		{
			Weight: sdk.NewInt(-1),
			Token:  sdk.NewCoin("test1", sdk.NewInt(50000)),
		},
	})
	require.Error(t, err)

	err = pacc.AddRecords([]Record{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test1", sdk.NewInt(0)),
		},
	})
	require.Error(t, err)

	err = pacc.AddRecords([]Record{
		{
			Weight: sdk.NewInt(100),
			Token: sdk.Coin{
				Denom:  "test1",
				Amount: sdk.NewInt(-1),
			},
		},
	})
	require.Error(t, err)

	require.Equal(t, 0, pacc.LenRecords())
}

func TestPoolAccountRecords(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	}).(*PoolAccount)

	_, err := pacc.GetRecord("test1")
	require.Error(t, err)

	err = pacc.AddRecords([]Record{
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

	require.Equal(t, 2, len(pacc.Records))
	require.Equal(t, 2, pacc.LenRecords())
	// Check that records are sorted.
	require.Equal(t, "test1", pacc.Records[0].Token.Denom)
	require.Equal(t, "test2", pacc.Records[1].Token.Denom)

	_, err = pacc.GetRecord("test1")
	require.NoError(t, err)
	_, err = pacc.GetRecord("test2")
	require.NoError(t, err)
	_, err = pacc.GetRecord("test3")
	require.Error(t, err)

	records, err := pacc.GetRecords("test1", "test2")
	require.NoError(t, err)
	require.Equal(t, 2, len(records))

	err = pacc.AddRecords([]Record{
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

	err = pacc.AddRecords([]Record{
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

	err = pacc.AddRecords([]Record{
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

	require.Equal(t, 4, len(pacc.Records))
	require.Equal(t, 4, pacc.LenRecords())
	// Check that records are sorted.
	require.Equal(t, "test1", pacc.Records[0].Token.Denom)
	require.Equal(t, "test2", pacc.Records[1].Token.Denom)
	require.Equal(t, "test3", pacc.Records[2].Token.Denom)
	require.Equal(t, "test4", pacc.Records[3].Token.Denom)

	_, err = pacc.GetRecord("test1")
	require.NoError(t, err)
	_, err = pacc.GetRecord("test2")
	require.NoError(t, err)
	_, err = pacc.GetRecord("test3")
	require.NoError(t, err)
	_, err = pacc.GetRecord("test4")
	require.NoError(t, err)
	_, err = pacc.GetRecord("test5")
	require.Error(t, err)

	records, err = pacc.GetRecords("test1", "test2", "test3", "test4")
	require.NoError(t, err)
	require.Equal(t, 4, len(records))

	_, err = pacc.GetRecords("test1", "test5")
	require.Error(t, err)
	_, err = pacc.GetRecords("test5")
	require.Error(t, err)

	records, err = pacc.GetRecords()
	require.NoError(t, err)
	require.Equal(t, 0, len(records))
}

func TestPoolAccountTotalWeight(t *testing.T) {
	var poolId uint64 = 10
	swapFee, _ := sdk.NewDecFromStr("0.025")
	exitFee, _ := sdk.NewDecFromStr("0.025")

	pacc := NewPoolAccount(poolId, PoolParams{
		Lock:    false,
		SwapFee: swapFee,
		ExitFee: exitFee,
	})

	err := pacc.AddRecords([]Record{
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

	err = pacc.AddRecords([]Record{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("test2", sdk.NewInt(10000)),
		},
	})
	require.Error(t, err)

	require.Equal(t, sdk.NewInt(300).String(), pacc.GetTotalWeight().String())

	err = pacc.AddRecords([]Record{
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
	})

	err := pacc.AddRecords([]Record{
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
  address: cosmos13st38lsk0rudja4sv98r2p5608nry4cc9qj7ff
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
  records:
  - denormalized_weight: "100"
    token:
      denom: test1
      amount: "10000"
  - denormalized_weight: "200"
    token:
      denom: test2
      amount: "50000"
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
	}).(*PoolAccount)

	err := pacc.AddRecords([]Record{
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
