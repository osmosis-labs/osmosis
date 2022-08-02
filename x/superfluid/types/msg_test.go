package types

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// // Test authz serialize and de-serializes for superfluid msg.
func TestAuthzMsg(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("stake", sdk.NewInt(1))
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)

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
			msg: &MsgLockAndSuperfluidDelegate{
				Sender:  addr1,
				Coins:   sdk.NewCoins(coin),
				ValAddr: "valoper1xyz",
			},
		},
		{
			name: "MsgSuperfluidDelegate",
			msg: &MsgSuperfluidDelegate{
				Sender:  addr1,
				LockId:  1,
				ValAddr: "valoper1xyz",
			},
		},
		{
			name: "MsgSuperfluidUnbondLock",
			msg: &MsgSuperfluidUnbondLock{
				Sender: addr1,
				LockId: 1,
			},
		},
		{
			name: "MsgSuperfluidUndelegate",
			msg: &MsgSuperfluidUndelegate{
				Sender: addr1,
				LockId: 1,
			},
		},
		{
			name: "MsgUnPoolWhitelistedPool",
			msg: &MsgUnPoolWhitelistedPool{
				Sender: addr1,
				PoolId: 1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				mockMsgGrant  authz.MsgGrant
				mockMsgRevoke authz.MsgRevoke
				mockMsgExec   authz.MsgExec
			)

			// Authz: Grant Msg
			typeURL := sdk.MsgTypeURL(tc.msg)
			grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), someDate.Add(time.Hour))
			require.NoError(t, err)

			msgGrant := authz.MsgGrant{Granter: mockGranter, Grantee: mockGrantee, Grant: grant}
			msgGrantBytes := json.RawMessage(sdk.MustSortJSON(authzcodec.ModuleCdc.MustMarshalJSON(&msgGrant)))
			err = authzcodec.ModuleCdc.UnmarshalJSON(msgGrantBytes, &mockMsgGrant)
			require.NoError(t, err)

			// Authz: Revoke Msg
			msgRevoke := authz.MsgRevoke{Granter: mockGranter, Grantee: mockGrantee, MsgTypeUrl: typeURL}
			msgRevokeByte := json.RawMessage(sdk.MustSortJSON(authzcodec.ModuleCdc.MustMarshalJSON(&msgRevoke)))
			err = authzcodec.ModuleCdc.UnmarshalJSON(msgRevokeByte, &mockMsgRevoke)
			require.NoError(t, err)

			// Authz: Exec Msg
			msgAny, _ := cdctypes.NewAnyWithValue(tc.msg)
			msgExec := authz.MsgExec{Grantee: mockGrantee, Msgs: []*cdctypes.Any{msgAny}}
			execMsgByte := json.RawMessage(sdk.MustSortJSON(authzcodec.ModuleCdc.MustMarshalJSON(&msgExec)))
			err = authzcodec.ModuleCdc.UnmarshalJSON(execMsgByte, &mockMsgExec)
			require.NoError(t, err)
			require.Equal(t, msgExec.Msgs[0].Value, mockMsgExec.Msgs[0].Value)
		})
	}
}
