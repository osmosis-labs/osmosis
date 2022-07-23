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
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgLockTokens) MsgLockTokens) MsgLockTokens {
		properMsg := MsgLockTokens{
			Owner:    addr1,
			Duration: time.Hour,
			Coins:    sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgLockTokens) MsgLockTokens {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "lock_tokens")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgLockTokens
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgLockTokens) MsgLockTokens {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: createMsg(func(msg MsgLockTokens) MsgLockTokens {
				msg.Owner = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid duration",
			msg: createMsg(func(msg MsgLockTokens) MsgLockTokens {
				msg.Duration = -1
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid coin length",
			msg: createMsg(func(msg MsgLockTokens) MsgLockTokens {
				msg.Coins = sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100000)), sdk.NewCoin("test2", sdk.NewInt(100000)))
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero token amount",
			msg: createMsg(func(msg MsgLockTokens) MsgLockTokens {
				msg.Coins = sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(0)))
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

func TestMsgBeginUnlockingAll(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgBeginUnlockingAll) MsgBeginUnlockingAll) MsgBeginUnlockingAll {
		properMsg := MsgBeginUnlockingAll{
			Owner:    addr1,
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgBeginUnlockingAll) MsgBeginUnlockingAll {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "begin_unlocking_all")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgBeginUnlockingAll
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgBeginUnlockingAll) MsgBeginUnlockingAll {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: createMsg(func(msg MsgBeginUnlockingAll) MsgBeginUnlockingAll {
				msg.Owner = invalidAddr.String()
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

func TestMsgBeginUnlocking(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgBeginUnlocking) MsgBeginUnlocking) MsgBeginUnlocking {
		properMsg := MsgBeginUnlocking{
			Owner:    addr1,
			ID: 1,
			Coins: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(100))),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgBeginUnlocking) MsgBeginUnlocking {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "begin_unlocking")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgBeginUnlocking
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgBeginUnlocking) MsgBeginUnlocking {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: createMsg(func(msg MsgBeginUnlocking) MsgBeginUnlocking {
				msg.Owner = invalidAddr.String()

				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lockup ID",
			msg: createMsg(func(msg MsgBeginUnlocking) MsgBeginUnlocking {
				msg.ID = 0

				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid coins length",
			msg: createMsg(func(msg MsgBeginUnlocking) MsgBeginUnlocking {
				msg.Coins = sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(100000)), sdk.NewCoin("test2", sdk.NewInt(100000)))

				return msg
			}),
			expectPass: false,
		},
		{
			name: "not positive coins amount",
			msg: createMsg(func(msg MsgBeginUnlocking) MsgBeginUnlocking {
				msg.Coins = sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(0)))

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

func TestMsgExtendLockup(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExtendLockup) MsgExtendLockup) MsgExtendLockup {
		properMsg := MsgExtendLockup{
			Owner:    addr1,
			ID: 1,
			Duration: time.Hour,
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg MsgExtendLockup) MsgExtendLockup {
		// Do nothing
		return msg
	})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "edit_lockup")
	signers := msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgExtendLockup
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExtendLockup) MsgExtendLockup {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: createMsg(func(msg MsgExtendLockup) MsgExtendLockup {
				msg.Owner = invalidAddr.String()

				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid lockup ID",
			msg: createMsg(func(msg MsgExtendLockup) MsgExtendLockup {
				msg.ID = 0

				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid duration",
			msg: createMsg(func(msg MsgExtendLockup) MsgExtendLockup {
				msg.Duration = -1

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

