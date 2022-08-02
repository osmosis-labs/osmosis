package types

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"
)

// // Test authz serialize and de-serializes for lockup msg.
func TestAuthzMsg(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("denom", sdk.NewInt(1))
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
			name: "MsgLockTokens",
			msg: &MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgBeginUnlocking",
			msg: &MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgBeginUnlockingAll",
			msg: &MsgBeginUnlockingAll{
				Owner: addr1,
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
