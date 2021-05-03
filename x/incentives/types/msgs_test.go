package types

import (
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestMsgCreatePool(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	createMsg := func(after func(msg MsgCreatePot) MsgCreatePot) MsgCreatePot {
		properMsg := MsgCreatePot{
			IsPerpetual: false,
			Owner:       addr1,
			DistributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         "lptoken",
				Duration:      time.Second,
			},
			Coins:             sdk.Coins{},
			StartTime:         time.Now(),
			NumEpochsPaidOver: 2,
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgCreatePot) MsgCreatePot {
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "create_pot")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        MsgCreatePot
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.Owner = sdk.AccAddress{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty distribution denom",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.DistributeTo.Denom = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution denom",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.DistributeTo.Denom = "111"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lock query type",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.DistributeTo.LockQueryType = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lock query type",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.DistributeTo.LockQueryType = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution start time",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.StartTime = time.Time{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.NumEpochsPaidOver = 0
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over for perpetual pot",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.NumEpochsPaidOver = 2
				msg.IsPerpetual = true
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid num epochs paid over for perpetual pot",
			msg: createMsg(func(msg MsgCreatePot) MsgCreatePot {
				msg.NumEpochsPaidOver = 1
				msg.IsPerpetual = true
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

func TestMsgAddToPot(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	createMsg := func(after func(msg MsgAddToPot) MsgAddToPot) MsgAddToPot {
		properMsg := MsgAddToPot{
			Owner:   addr1,
			PotId:   1,
			Rewards: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgAddToPot) MsgAddToPot {
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "add_to_pot")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        MsgAddToPot
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgAddToPot) MsgAddToPot {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg MsgAddToPot) MsgAddToPot {
				msg.Owner = sdk.AccAddress{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty rewards",
			msg: createMsg(func(msg MsgAddToPot) MsgAddToPot {
				msg.Rewards = sdk.Coins{}
				return msg
			}),
			expectPass: false,
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
