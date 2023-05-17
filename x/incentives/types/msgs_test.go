package types_test

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"

	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"

	appParams "github.com/osmosis-labs/osmosis/v15/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

// TestMsgCreatePool tests if valid/invalid create pool messages are properly validated/invalidated
func TestMsgCreatePool(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper createPool message
	createMsg := func(after func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
		distributeTo := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		}

		properMsg := *incentivestypes.NewMsgCreateGauge(
			false,
			addr1,
			distributeTo,
			sdk.Coins{},
			time.Now(),
			2,
		)

		return after(properMsg)
	}

	// validate createPool message was created as intended
	msg := createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
		return msg
	})
	require.Equal(t, msg.Route(), incentivestypes.RouterKey)
	require.Equal(t, msg.Type(), "create_gauge")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        incentivestypes.MsgCreateGauge
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty distribution denom",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.Denom = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution denom",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.Denom = "111"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lock query type",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lock query type",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid distribution start time",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.StartTime = time.Time{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.NumEpochsPaidOver = 0
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid num epochs paid over for perpetual gauge",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.NumEpochsPaidOver = 2
				msg.IsPerpetual = true
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid num epochs paid over for perpetual gauge",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
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

// TestMsgAddToGauge tests if valid/invalid add to gauge messages are properly validated/invalidated
func TestMsgAddToGauge(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper addToGauge message
	createMsg := func(after func(msg incentivestypes.MsgAddToGauge) incentivestypes.MsgAddToGauge) incentivestypes.MsgAddToGauge {
		properMsg := *incentivestypes.NewMsgAddToGauge(
			addr1,
			1,
			sdk.Coins{sdk.NewInt64Coin("stake", 10)},
		)

		return after(properMsg)
	}

	// validate addToGauge message was created as intended
	msg := createMsg(func(msg incentivestypes.MsgAddToGauge) incentivestypes.MsgAddToGauge {
		return msg
	})
	require.Equal(t, msg.Route(), incentivestypes.RouterKey)
	require.Equal(t, msg.Type(), "add_to_gauge")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        incentivestypes.MsgAddToGauge
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg incentivestypes.MsgAddToGauge) incentivestypes.MsgAddToGauge {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg incentivestypes.MsgAddToGauge) incentivestypes.MsgAddToGauge {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty rewards",
			msg: createMsg(func(msg incentivestypes.MsgAddToGauge) incentivestypes.MsgAddToGauge {
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

// // Test authz serialize and de-serializes for incentives msg.
func TestAuthzMsg(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)

	testCases := []struct {
		name          string
		incentivesMsg sdk.Msg
	}{
		{
			name: "MsgAddToGauge",
			incentivesMsg: &incentivestypes.MsgAddToGauge{
				Owner:   addr1,
				GaugeId: 1,
				Rewards: sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgCreateGauge",
			incentivesMsg: &incentivestypes.MsgCreateGauge{
				IsPerpetual: false,
				Owner:       addr1,
				DistributeTo: lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         "lptoken",
					Duration:      time.Second,
				},
				Coins:             sdk.NewCoins(coin),
				StartTime:         someDate,
				NumEpochsPaidOver: 1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.incentivesMsg)
		})
	}
}
