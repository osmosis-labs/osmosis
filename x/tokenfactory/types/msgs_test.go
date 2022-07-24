package types

import (
	fmt "fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appParams "github.com/osmosis-labs/osmosis/v10/app/params"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestAuthzMsg(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("denom", sdk.NewInt(1))
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)

	testCases := []struct {
		expectedGrantSignByteMsg   string
		expectedRevokeSignByteMsg  string
		expectedExecStrSignByteMsg string
		msg                        sdk.Msg
	}{
		{
			expectedGrantSignByteMsg:   `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/GenericAuthorization","value":{"msg":"/osmosis.tokenfactory.v1beta1.MsgCreateDenom"}},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"cosmos1def","granter":"cosmos1abc"}}],"sequence":"1","timeout_height":"1"}`,
			expectedRevokeSignByteMsg:  `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgRevoke","value":{"grantee":"cosmos1def","granter":"cosmos1abc","msg_type_url":"/osmosis.tokenfactory.v1beta1.MsgCreateDenom"}}],"sequence":"1","timeout_height":"1"}`,
			expectedExecStrSignByteMsg: fmt.Sprintf(`{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgExec","value":{"grantee":"cosmos1def","msgs":[{"sender":"%s","subdenom":"valoper1xyz"}]}}],"sequence":"1","timeout_height":"1"}`, addr1),
			msg: &MsgCreateDenom{
				Sender:   addr1,
				Subdenom: "valoper1xyz",
			},
		},
		{
			expectedGrantSignByteMsg:   `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/GenericAuthorization","value":{"msg":"/osmosis.tokenfactory.v1beta1.MsgBurn"}},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"cosmos1def","granter":"cosmos1abc"}}],"sequence":"1","timeout_height":"1"}`,
			expectedRevokeSignByteMsg:  `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgRevoke","value":{"grantee":"cosmos1def","granter":"cosmos1abc","msg_type_url":"/osmosis.tokenfactory.v1beta1.MsgBurn"}}],"sequence":"1","timeout_height":"1"}`,
			expectedExecStrSignByteMsg: fmt.Sprintf(`{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgExec","value":{"grantee":"cosmos1def","msgs":[{"amount":{"amount":"1","denom":"denom"},"sender":"%s"}]}}],"sequence":"1","timeout_height":"1"}`, addr1),
			msg: &MsgBurn{
				Sender: addr1,
				Amount: coin,
			},
		},
		{
			expectedGrantSignByteMsg:   `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/GenericAuthorization","value":{"msg":"/osmosis.tokenfactory.v1beta1.MsgMint"}},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"cosmos1def","granter":"cosmos1abc"}}],"sequence":"1","timeout_height":"1"}`,
			expectedRevokeSignByteMsg:  `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgRevoke","value":{"grantee":"cosmos1def","granter":"cosmos1abc","msg_type_url":"/osmosis.tokenfactory.v1beta1.MsgMint"}}],"sequence":"1","timeout_height":"1"}`,
			expectedExecStrSignByteMsg: fmt.Sprintf(`{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgExec","value":{"grantee":"cosmos1def","msgs":[{"amount":{"amount":"1","denom":"denom"},"sender":"%s"}]}}],"sequence":"1","timeout_height":"1"}`, addr1),
			msg: &MsgMint{
				Sender: addr1,
				Amount: coin,
			},
		},
		{
			expectedGrantSignByteMsg:   `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgGrant","value":{"grant":{"authorization":{"type":"cosmos-sdk/GenericAuthorization","value":{"msg":"/osmosis.tokenfactory.v1beta1.MsgChangeAdmin"}},"expiration":"0001-01-01T02:01:01.000000001Z"},"grantee":"cosmos1def","granter":"cosmos1abc"}}],"sequence":"1","timeout_height":"1"}`,
			expectedRevokeSignByteMsg:  `{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgRevoke","value":{"grantee":"cosmos1def","granter":"cosmos1abc","msg_type_url":"/osmosis.tokenfactory.v1beta1.MsgChangeAdmin"}}],"sequence":"1","timeout_height":"1"}`,
			expectedExecStrSignByteMsg: fmt.Sprintf(`{"account_number":"1","chain_id":"foo","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgExec","value":{"grantee":"cosmos1def","msgs":[{"denom":"denom","new_admin":"osmo1q8tq5qhrhw6t970egemuuwywhlhpnmdmts6xnu","sender":"%s"}]}}],"sequence":"1","timeout_height":"1"}`, addr1),
			msg: &MsgChangeAdmin{
				Sender:   addr1,
				Denom:    "denom",
				NewAdmin: "osmo1q8tq5qhrhw6t970egemuuwywhlhpnmdmts6xnu",
			},
		},
	}
	for _, tc := range testCases {
		// Authz: Grant Msg
		typeURL := sdk.MsgTypeURL(tc.msg)
		grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), someDate.Add(time.Hour))
		require.NoError(t, err)
		msgGrant := &authz.MsgGrant{Granter: "cosmos1abc", Grantee: "cosmos1def", Grant: grant}
		require.Equal(t,
			tc.expectedGrantSignByteMsg,
			string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgGrant}, "memo")),
		)

		// Authz: Revoke Msg
		msgRevoke := &authz.MsgRevoke{Granter: "cosmos1abc", Grantee: "cosmos1def", MsgTypeUrl: typeURL}
		require.Equal(t,
			tc.expectedRevokeSignByteMsg,
			string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgRevoke}, "memo")),
		)

		// Authz: Exec Msg
		msgAny, _ := cdctypes.NewAnyWithValue(tc.msg)
		msgExec := &authz.MsgExec{Grantee: "cosmos1def", Msgs: []*cdctypes.Any{msgAny}}
		require.Equal(t,
			tc.expectedExecStrSignByteMsg,
			string(legacytx.StdSignBytes("foo", 1, 1, 1, legacytx.StdFee{}, []sdk.Msg{msgExec}, "memo")),
		)
	}
}
