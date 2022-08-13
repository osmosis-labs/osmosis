package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/x/tokenfactory/types"

	"github.com/tendermint/tendermint/crypto/ed25519"
)

// // Test authz serialize and de-serializes for tokenfactory msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("denom", sdk.NewInt(1))

	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgCreateDenom",
			msg: &types.MsgCreateDenom{
				Sender:   addr1,
				Subdenom: "valoper1xyz",
			},
		},
		{
			name: "MsgBurn",
			msg: &types.MsgBurn{
				Sender: addr1,
				Amount: coin,
			},
		},
		{
			name: "MsgMint",
			msg: &types.MsgMint{
				Sender: addr1,
				Amount: coin,
			},
		},
		{
			name: "MsgChangeAdmin",
			msg: &types.MsgChangeAdmin{
				Sender:   addr1,
				Denom:    "denom",
				NewAdmin: "osmo1q8tq5qhrhw6t970egemuuwywhlhpnmdmts6xnu",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.msg)
		})
	}
}
