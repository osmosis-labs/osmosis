package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"
)

func TestMsgLockTokens(t *testing.T) {
	appParams.SetAddressPrefixes()
	addr1, invalidAddr := generateTestAddrs()

	tests := []struct {
		name       string
		msg        MsgLockTokens
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: MsgLockTokens{
				Owner:    invalidAddr,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
			},
		},
		{
			name: "invalid duration",
			msg: MsgLockTokens{
				Owner:    addr1,
				Duration: -1,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
			},
		},
		{
			name: "invalid coin length",
			msg: MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100000)), sdk.NewCoin("test2", sdk.NewInt(100000))),
			},
		},
		{
			name: "zero token amount",
			msg: MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(0))),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), RouterKey)
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
	addr1, invalidAddr := generateTestAddrs()

	tests := []struct {
		name       string
		msg        MsgBeginUnlockingAll
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: MsgBeginUnlockingAll{
				Owner: addr1,
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: MsgBeginUnlockingAll{
				Owner: invalidAddr,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), RouterKey)
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
	addr1, invalidAddr := generateTestAddrs()

	tests := []struct {
		name       string
		msg        MsgBeginUnlocking
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: MsgBeginUnlocking{
				Owner: invalidAddr,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
			},
		},
		{
			name: "invalid lockup ID",
			msg: MsgBeginUnlocking{
				Owner: addr1,
				ID:    0,
				Coins: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
			},
		},
		{
			name: "invalid coins length",
			msg: MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100000)), sdk.NewCoin("test2", sdk.NewInt(100000))),
			},
		},
		{
			name: "not positive coins amount",
			msg: MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(0))),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), RouterKey)
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
	addr1, invalidAddr := generateTestAddrs()

	tests := []struct {
		name       string
		msg        MsgExtendLockup
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: MsgExtendLockup{
				Owner:    addr1,
				ID:       1,
				Duration: time.Hour,
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: MsgExtendLockup{
				Owner:    invalidAddr,
				ID:       1,
				Duration: time.Hour,
			},
		},
		{
			name: "invalid lockup ID",
			msg: MsgExtendLockup{
				Owner:    addr1,
				ID:       0,
				Duration: time.Hour,
			},
		},
		{
			name: "invalid duration",
			msg: MsgExtendLockup{
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
				require.Equal(t, test.msg.Route(), RouterKey)
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

func generateTestAddrs() (string, string) {
	pk1 := ed25519.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid").String()
	return validAddr, invalidAddr 
}
