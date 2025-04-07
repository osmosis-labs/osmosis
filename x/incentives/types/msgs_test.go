package types_test

import (
	"testing"
	time "time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/incentives"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"

	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

// TestMsgCreateGauge tests if valid/invalid create gauge messages are properly validated/invalidated
func TestMsgCreateGauge(t *testing.T) {
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
			0,
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
		{
			name: "invalid: by time lock type",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.ByTime
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid: by duration with pool id set",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.ByDuration
				msg.PoolId = 1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid: no lock with pool id unset",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.NoLock
				msg.PoolId = 0
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid no lock with pool id unset",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.NoLock
				msg.DistributeTo.Denom = ""
				msg.DistributeTo.Duration = 0
				msg.PoolId = 1
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid due to denom being set",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.NoLock
				msg.DistributeTo.Denom = "stake"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid due to external denom being set",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.NoLock
				// This is set by the system. Client should provide empty string.
				msg.DistributeTo.Denom = types.NoLockExternalGaugeDenom(1)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid due to internal denom being set",
			msg: createMsg(func(msg incentivestypes.MsgCreateGauge) incentivestypes.MsgCreateGauge {
				msg.DistributeTo.LockQueryType = lockuptypes.NoLock
				// This is set by the system when creating internal gauges.
				// Client should provide empty string.
				msg.DistributeTo.Denom = types.NoLockInternalGaugeDenom(1)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
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

func TestMsgCreateGroup(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper createGroup message
	createMsg := func(after func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
		properMsg := *incentivestypes.NewMsgCreateGroup(
			sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			0,
			addr1,
			[]uint64{1, 2, 3},
		)

		return after(properMsg)
	}

	// validate createGroup message was created as intended
	msg := createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
		return msg
	})
	require.Equal(t, msg.Route(), incentivestypes.RouterKey)
	require.Equal(t, msg.Type(), "create_group")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        incentivestypes.MsgCreateGroup
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty owner",
			msg: createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
				msg.Owner = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "only one pool id",
			msg: createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
				msg.PoolIds = []uint64{1}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "greater than 30 pool ids",
			msg: createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
				msg.PoolIds = []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 1, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "repeated pool id",
			msg: createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
				msg.PoolIds = []uint64{1, 2, 1}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "non-perpetual group creation is disabled",
			msg: createMsg(func(msg incentivestypes.MsgCreateGroup) incentivestypes.MsgCreateGroup {
				msg.NumEpochsPaidOver = 2
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
	coin := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1))
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
			apptesting.TestMessageAuthzSerialization(t, tc.incentivesMsg, incentives.AppModuleBasic{})
		})
	}
}
