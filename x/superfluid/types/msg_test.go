package types_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

// // Test authz serialize and de-serializes for superfluid msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1))

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgLockAndSuperfluidDelegate",
			msg: &types.MsgLockAndSuperfluidDelegate{
				Sender:  addr1,
				Coins:   sdk.NewCoins(coin),
				ValAddr: "valoper1xyz",
			},
		},
		{
			name: "MsgSuperfluidDelegate",
			msg: &types.MsgSuperfluidDelegate{
				Sender:  addr1,
				LockId:  1,
				ValAddr: "valoper1xyz",
			},
		},
		{
			name: "MsgSuperfluidUnbondLock",
			msg: &types.MsgSuperfluidUnbondLock{
				Sender: addr1,
				LockId: 1,
			},
		},
		{
			name: "MsgSuperfluidUndelegateAndUnbondLock",
			msg: &types.MsgSuperfluidUndelegateAndUnbondLock{
				Sender: addr1,
				LockId: 1,
				Coin:   coin,
			},
		},
		{
			name: "MsgSuperfluidUndelegate",
			msg: &types.MsgSuperfluidUndelegate{
				Sender: addr1,
				LockId: 1,
			},
		},
		{
			name: "MsgUnPoolWhitelistedPool",
			msg: &types.MsgUnPoolWhitelistedPool{
				Sender: addr1,
				PoolId: 1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.msg, superfluid.AppModuleBasic{})
		})
	}
}

func TestUnbondConvertAndStakeMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()

	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address()).String()

	testCases := []struct {
		name          string
		msg           sdk.Msg
		expectedError bool
	}{
		{
			name: "happy case",
			msg: &types.MsgUnbondConvertAndStake{
				LockId:          2,
				Sender:          addr1,
				ValAddr:         valAddr,
				MinAmtToStake:   osmomath.NewInt(10),
				SharesToConvert: sdk.NewInt64Coin("foo", 10),
			},
		},
		{
			name: "lock id is 0 should not fail",
			msg: &types.MsgUnbondConvertAndStake{
				LockId:          0,
				Sender:          addr1,
				ValAddr:         valAddr,
				MinAmtToStake:   osmomath.NewInt(10),
				SharesToConvert: sdk.NewInt64Coin("foo", 10),
			},
		},
		{
			name: "no val address should not fail",
			msg: &types.MsgUnbondConvertAndStake{
				LockId:          0,
				Sender:          addr1,
				MinAmtToStake:   osmomath.NewInt(10),
				SharesToConvert: sdk.NewInt64Coin("foo", 10),
			},
		},
		{
			name: "err: sender is invalid",
			msg: &types.MsgUnbondConvertAndStake{
				LockId:          0,
				Sender:          "abcd",
				ValAddr:         valAddr,
				MinAmtToStake:   osmomath.NewInt(10),
				SharesToConvert: sdk.NewInt64Coin("foo", 10),
			},
			expectedError: true,
		},
		{
			name: "err: min amount to stake is negative",
			msg: &types.MsgUnbondConvertAndStake{
				LockId:          0,
				Sender:          addr1,
				MinAmtToStake:   osmomath.NewInt(10).Neg(),
				SharesToConvert: sdk.NewInt64Coin("foo", 10),
			},
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msgWithValBasic, ok := tc.msg.(sdk.HasValidateBasic)
			require.True(t, ok)
			err := msgWithValBasic.ValidateBasic()
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
