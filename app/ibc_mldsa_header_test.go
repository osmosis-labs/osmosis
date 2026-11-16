package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v32/app"
)

func TestIBCMLDSA65HeaderTransaction(t *testing.T) {
	// Public signed light block from the DoraVota v1.1.0 IBC compatibility lab.
	raw, err := os.ReadFile("testdata/dora_mldsa65_light_block.json")
	require.NoError(t, err)
	var fixture struct {
		LightBlock []byte `json:"light_block"`
	}
	require.NoError(t, json.Unmarshal(raw, &fixture))
	var lightBlock cmtproto.LightBlock
	require.NoError(t, lightBlock.Unmarshal(fixture.LightBlock))

	header := &ibctm.Header{
		SignedHeader:      lightBlock.SignedHeader,
		ValidatorSet:      lightBlock.ValidatorSet,
		TrustedValidators: lightBlock.ValidatorSet,
		TrustedHeight:     clienttypes.NewHeight(0, uint64(lightBlock.SignedHeader.Header.Height-1)),
	}
	require.NoError(t, header.ValidateBasic())
	clientMessage, err := codectypes.NewAnyWithValue(header)
	require.NoError(t, err)
	signer := sdk.AccAddress(bytes.Repeat([]byte{42}, 20)).String()
	update := &clienttypes.MsgUpdateClient{
		ClientId:      "07-tendermint-0",
		ClientMessage: clientMessage,
		Signer:        signer,
	}

	txConfig := app.MakeEncodingConfig().TxConfig
	for _, tc := range []struct {
		name string
		msgs []sdk.Msg
	}{
		{"client update", []sdk.Msg{update}},
		{"client update batched with connection open try", []sdk.Msg{
			update,
			&connectiontypes.MsgConnectionOpenTry{ClientId: update.ClientId, Signer: signer},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := txConfig.NewTxBuilder()
			require.NoError(t, builder.SetMsgs(tc.msgs...))
			encoded, err := txConfig.TxEncoder()(builder.GetTx())
			require.NoError(t, err)
			decoded, err := txConfig.TxDecoder()(encoded)
			require.NoError(t, err)
			require.Len(t, decoded.GetMsgs(), len(tc.msgs))
			decodedUpdate, ok := decoded.GetMsgs()[0].(*clienttypes.MsgUpdateClient)
			require.True(t, ok)
			require.Equal(t, clientMessage.Value, decodedUpdate.ClientMessage.Value)
		})
	}

	signed, err := cmttypes.SignedHeaderFromProto(header.SignedHeader)
	require.NoError(t, err)
	validators, err := cmttypes.ValidatorSetFromProto(header.ValidatorSet)
	require.NoError(t, err)
	verify := func() error {
		return validators.VerifyCommitLight(signed.ChainID, signed.Commit.BlockID, signed.Height, signed.Commit)
	}
	require.NoError(t, verify())
	signed.Commit.Signatures[0].Signature[0] ^= 1
	require.Error(t, verify())
}
