package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdcutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/stretchr/testify/require"

	appParams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

type extMsg interface {
	sdk.Msg
	Route() string
	Type() string
	ValidateBasic() error
}

var (
	addr1       string
	addr2       string
	invalidAddr sdk.AccAddress
)

func init() {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 = sdk.AccAddress(pk1.Address()).String()
	pk2 := ed25519.GenPrivKey().PubKey()
	addr2 = sdk.AccAddress(pk2.Address()).String()
	invalidAddr = sdk.AccAddress("invalid")
}

func runValidateBasicTest(t *testing.T, name string, msg extMsg, expectPass bool, expType string) {
	if expectPass {
		require.NoError(t, msg.ValidateBasic(), "test: %v", name)
		require.Equal(t, msg.Route(), types.RouterKey)
		require.Equal(t, msg.Type(), expType)
		aminoCodec := codec.NewLegacyAmino()
		interfaceRegistry := cdcutil.CodecOptions{AccAddressPrefix: "osmo", ValAddressPrefix: "osmovaloper"}.NewInterfaceRegistry()
		codec := codec.NewProtoCodec(interfaceRegistry)

		encCfg := testutil.TestEncodingConfig{
			InterfaceRegistry: interfaceRegistry,
			Codec:             codec,
			TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
			Amino:             aminoCodec,
		}
		signers, _, err := encCfg.Codec.GetMsgV1Signers(msg)
		require.NoError(t, err)
		require.Equal(t, len(signers), 1)
		require.Equal(t, sdk.AccAddress(signers[0]).String(), addr1)
	} else {
		require.Error(t, msg.ValidateBasic(), "test: %v", name)
	}
}

func TestMsgSetFeeTokens(t *testing.T) {
	tests := []struct {
		name       string
		msg        types.MsgSetFeeTokens
		expectPass bool
	}{
		{
			name: "proper msg: set multiple fee tokens",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{
					{Denom: "foo", PoolID: 1},
					{Denom: "bar", PoolID: 2},
				},
				Sender: addr1,
			},
			expectPass: true,
		},
		{
			name: "proper msg: set single fee token",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{
					{Denom: "foo", PoolID: 1},
				},
				Sender: addr1,
			},
			expectPass: true,
		},
		{
			name: "improper msg: empty fee tokens",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{},
				Sender:    addr1,
			},
			expectPass: false,
		},
		{
			name: "improper msg: invalid sender address",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{
					{Denom: "foo", PoolID: 1},
				},
				Sender: invalidAddr.String(),
			},
			expectPass: false,
		},
		{
			name: "improper msg: empty sender address",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{
					{Denom: "foo", PoolID: 1},
				},
				Sender: "",
			},
			expectPass: false,
		},
		{
			name: "improper msg: empty fee tokens and invalid sender address",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{},
				Sender:    invalidAddr.String(),
			},
			expectPass: false,
		},
		{
			name: "improper msg: empty fee tokens and empty sender address",
			msg: types.MsgSetFeeTokens{
				FeeTokens: []types.FeeToken{},
				Sender:    "",
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		runValidateBasicTest(t, test.name, &test.msg, test.expectPass, types.TypeMsgSetFeeTokens)
	}
}
