package authenticator_test

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
)

// Test data for the sees to run
const (
	TestKey  = "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	TestKey2 = "0dd4d1506e18a5712080708c338eb51ecf2afdceae01e8162e890b126ac190fe"
	TestKey3 = "49006a359803f0602a7ec521df88bf5527579da79112bb71f285dd3e7d438033"
)

func TestSecp256k1SignatureAuthenticator(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	txConfig := encodingConfig.TxConfig
	signModeHandler := txConfig.SignModeHandler()
	osmosisApp := app.Setup(false)
	ak := osmosisApp.AccountKeeper
	ctx := osmosisApp.NewContext(false, tmproto.Header{})

	bz, _ := hex.DecodeString(TestKey)
	priv1 := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv1.PubKey().Address())
	account1 := authtypes.NewBaseAccount(accAddress, priv1.PubKey(), 0, 0)
	ak.SetAccount(ctx, account1)

	// decode the test private key
	bz, _ = hex.DecodeString(TestKey)
	priv2 := &secp256k1.PrivKey{Key: bz}
	accAddress2 := sdk.AccAddress(priv2.PubKey().Address())
	account2 := authtypes.NewBaseAccount(accAddress2, priv2.PubKey(), 0, 0)
	ak.SetAccount(ctx, account2)

	// Create a new Secp256k1SignatureAuthenticator for testing
	authenticator := authenticator.NewSignatureVerificationAuthenticator(ak, signModeHandler)
	coins := sdk.Coins{sdk.NewInt64Coin("osmo", 2500)}

	sendMsg := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes("osmo", accAddress),
		ToAddress:   sdk.MustBech32ifyAddressBytes("osmo", accAddress2),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin("osmo", 2500)}

	tx, err := GenTx(
		txConfig,
		[]sdk.Msg{
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
			sendMsg,
		},
		feeCoins,
		300000,
		"",
		[]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		priv1,
		priv1,
		priv1,
		priv1,
		priv1,
		priv1,
		priv1,
		priv1,
		priv1,
		priv1,
	)

	//fmt.Println(tx)

	// Test GetAuthenticationData
	authData, err := authenticator.GetAuthenticationData(ctx, tx, 0, false)
	require.NoError(t, err)
	//fmt.Println(authData)
	//require.Equal(t, testData, authData)

	// Test Authenticate
	for _, msg := range tx.GetMsgs() {
		success, err := authenticator.Authenticate(ctx, msg, authData)
		require.NoError(t, err)
		require.True(t, success)
	}
}

// GenTx generates a signed mock transaction.
func GenTx(
	gen client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums,
	accSeqs []uint64,
	priv ...cryptotypes.PrivKey,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))
	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}
