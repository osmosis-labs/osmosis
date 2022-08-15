package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/types"

	"github.com/tendermint/tendermint/crypto/ed25519"
)

// // Test authz serialize and de-serializes for superfluid msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))

	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

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
			apptesting.TestMessageAuthzSerialization(t, tc.msg)
		})
	}
}
