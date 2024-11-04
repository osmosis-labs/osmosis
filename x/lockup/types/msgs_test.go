package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/lockup"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	"github.com/cometbft/cometbft/crypto/ed25519"

	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
)

func TestMsgLockTokens(t *testing.T) {
	appParams.SetAddressPrefixes()
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        types.MsgLockTokens
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(100))),
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: types.MsgLockTokens{
				Owner:    invalidAddr,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(100))),
			},
		},
		{
			name: "invalid duration",
			msg: types.MsgLockTokens{
				Owner:    addr1,
				Duration: -1,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(100))),
			},
		},
		{
			name: "invalid coin length",
			msg: types.MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test1", osmomath.NewInt(100000)), sdk.NewCoin("test2", osmomath.NewInt(100000))),
			},
		},
		{
			name: "zero token amount",
			msg: types.MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(0))),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "lock_tokens")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

func TestMsgBeginUnlockingAll(t *testing.T) {
	appParams.SetAddressPrefixes()
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        types.MsgBeginUnlockingAll
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgBeginUnlockingAll{
				Owner: addr1,
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: types.MsgBeginUnlockingAll{
				Owner: invalidAddr,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "begin_unlocking_all")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

func TestMsgBeginUnlocking(t *testing.T) {
	appParams.SetAddressPrefixes()
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        types.MsgBeginUnlocking
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(100))),
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: types.MsgBeginUnlocking{
				Owner: invalidAddr,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(100))),
			},
		},
		{
			name: "invalid lockup ID",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    0,
				Coins: sdk.NewCoins(sdk.NewCoin("test", osmomath.NewInt(100))),
			},
		},
		{
			name: "invalid coins length",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test1", osmomath.NewInt(100000)), sdk.NewCoin("test2", osmomath.NewInt(100000))),
			},
		},
		{
			name: "zero coins (same as nil)",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test1", osmomath.NewInt(0))),
			},
			expectPass: true,
		},
		{
			name: "nil coins (unlock by ID)",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(),
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "begin_unlocking")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

func TestMsgExtendLockup(t *testing.T) {
	appParams.SetAddressPrefixes()
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        types.MsgExtendLockup
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgExtendLockup{
				Owner:    addr1,
				ID:       1,
				Duration: time.Hour,
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: types.MsgExtendLockup{
				Owner:    invalidAddr,
				ID:       1,
				Duration: time.Hour,
			},
		},
		{
			name: "invalid lockup ID",
			msg: types.MsgExtendLockup{
				Owner:    addr1,
				ID:       0,
				Duration: time.Hour,
			},
		},
		{
			name: "invalid duration",
			msg: types.MsgExtendLockup{
				Owner:    addr1,
				ID:       1,
				Duration: -1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "edit_lockup")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

// // Test authz serialize and de-serializes for lockup msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("denom", osmomath.NewInt(1))

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgLockTokens",
			msg: &types.MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgBeginUnlocking",
			msg: &types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgBeginUnlockingAll",
			msg: &types.MsgBeginUnlockingAll{
				Owner: addr1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.msg, lockup.AppModuleBasic{})
		})
	}
}
