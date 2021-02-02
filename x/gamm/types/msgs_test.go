package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgCreatePool(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1, err := sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	require.NoError(t, err)

	createMsg := func(after func(msg MsgCreatePool) MsgCreatePool) MsgCreatePool {
		properMsg := MsgCreatePool{
			Sender: addr1,
			PoolParams: PoolParams{
				Lock:    false,
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			},
			Records: []Record{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test", sdk.NewInt(100)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test2", sdk.NewInt(100)),
				},
			},
		}

		return after(properMsg)
	}

	tests := []struct {
		name       string
		msg        MsgCreatePool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "has no record",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has no record2",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records = []Record{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has one record",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records = []Record{
					msg.Records[0],
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes 0 weight",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Weight = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the negative weight",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the negative weight",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the zero coin",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Token = sdk.NewCoin("test1", sdk.NewInt(0))
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the negative coin",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Token = sdk.Coin{
					Denom:  "test1",
					Amount: sdk.NewInt(-10),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "locked pool",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.Lock = true
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative swap fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.SwapFee = sdk.NewDecWithPrec(-1, 2)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative exit fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.ExitFee = sdk.NewDecWithPrec(-1, 2)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero swap fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.SwapFee = sdk.NewDec(0)
				return msg
			}),
			expectPass: true,
		},
		{
			name: "zero exit fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.ExitFee = sdk.NewDec(0)
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}
