package post_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/post"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/testutils"
	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/suite"
)

type AuthenticatorPostSuite struct {
	suite.Suite
	OsmosisApp                 *app.OsmosisApp
	Ctx                        sdk.Context
	EncodingConfig             params.EncodingConfig
	AuthenticatorPostDecorator post.AuthenticatorPostDecorator
	TestKeys                   []string
	TestAccAddress             []sdk.AccAddress
	TestPrivKeys               []*secp256k1.PrivKey
	HomeDir                    string
}

func TestAuthenticatorPostSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorPostSuite))
}

func (s *AuthenticatorPostSuite) SetupTest() {
	// Test data for authenticator signature verification
	TestKeys := []string{
		"6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159",
		"0dd4d1506e18a5712080708c338eb51ecf2afdceae01e8162e890b126ac190fe",
		"49006a359803f0602a7ec521df88bf5527579da79112bb71f285dd3e7d438033",
	}
	s.EncodingConfig = app.MakeEncodingConfig()

	s.HomeDir = fmt.Sprintf("%d", rand.Int())
	s.OsmosisApp = app.SetupWithCustomHome(false, s.HomeDir)

	s.Ctx = s.OsmosisApp.NewContextLegacy(false, tmproto.Header{})

	// Set up test accounts
	for _, key := range TestKeys {
		bz, _ := hex.DecodeString(key)
		priv := &secp256k1.PrivKey{Key: bz}

		// add the test private keys to array for later use
		s.TestPrivKeys = append(s.TestPrivKeys, priv)

		accAddress := sdk.AccAddress(priv.PubKey().Address())
		authtypes.NewBaseAccount(accAddress, priv.PubKey(), 0, 0)

		// add the test accounts to array for later use
		s.TestAccAddress = append(s.TestAccAddress, accAddress)
	}

	s.AuthenticatorPostDecorator = post.NewAuthenticatorPostDecorator(
		s.OsmosisApp.AppCodec(),
		s.OsmosisApp.SmartAccountKeeper,
		s.OsmosisApp.AccountKeeper,
		s.EncodingConfig.TxConfig.SignModeHandler(),
		// Add an empty handler here to enable a circuit breaker pattern
		sdk.ChainPostDecorators(sdk.Terminator{}), //nolint
	)
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(1_000_000))
}

func (s *AuthenticatorPostSuite) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

// TestAutenticatorPostHandlerSuccess tests that the post handler can succeed with the default authenticator
func (s *AuthenticatorPostSuite) TestAuthenticatorPostHandlerSuccess() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create a test messages for signing
	testMsg1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	testMsg2 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Add the authenticators for the accounts
	id, err := s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[0],
		"SignatureVerification",
		s.TestPrivKeys[0].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(id, uint64(1), "Adding authenticator returning incorrect id")

	id, err = s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[1],
		"SignatureVerification",
		s.TestPrivKeys[1].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(id, uint64(2), "Adding authenticator returning incorrect id")

	tx, _ := GenTx(s.Ctx, s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
		testMsg2,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []uint64{1, 2})

	postHandler := sdk.ChainPostDecorators(s.AuthenticatorPostDecorator)
	_, err = postHandler(s.Ctx, tx, false, true)

	s.Require().NoError(err, "Failed but should have passed as ConfirmExecution passed")
}

// TestAutenticatorPostHandlerReturnEarly tests that the post handler fails early on IsCircuitBreakActive
// the transaction should pass through the normal flow.
func (s *AuthenticatorPostSuite) TestAuthenticatorPostHandlerReturnEarly() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create a test messages for signing
	testMsg1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Generate a transaction that is signed incorrectly
	tx, _ := GenTx(s.Ctx, s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[1],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[1],
	}, []uint64{})

	postHandler := sdk.ChainPostDecorators(s.AuthenticatorPostDecorator)
	_, err := postHandler(s.Ctx, tx, false, true)

	s.Require().NoError(err, "Failed but should have passed with no-op")
}

// TestAuthenticatorPostHandlerFailConfirmExecution tests how the post handler behaves when ConfirmExecution fails.
func (s *AuthenticatorPostSuite) TestAuthenticatorPostHandlerFailConfirmExecution() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}
	approveAndBlock := testutils.TestingAuthenticator{
		Approve:        testutils.Always,
		GasConsumption: 10,
		Confirm:        testutils.Never,
	}
	s.OsmosisApp.AuthenticatorManager.RegisterAuthenticator(approveAndBlock)
	approveAndBlockId, err := s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(s.Ctx, s.TestAccAddress[0], approveAndBlock.Type(), []byte{})
	s.Require().NoError(err, "Should have been able to add an authenticator")

	// Create a test messages for signing
	testMsg1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Generate a transaction that is signed correctly
	tx, _ := GenTx(s.Ctx, s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
	}, []uint64{approveAndBlockId})

	postHandler := sdk.ChainPostDecorators(s.AuthenticatorPostDecorator)
	_, err = postHandler(s.Ctx, tx, false, true)
	s.Require().Error(err, "Should have failed on ConfirmExecution")
	s.Require().ErrorContains(err, "execution blocked by authenticator")
}

// GenTx generates a signed mock transaction.
func GenTx(
	ctx sdk.Context,
	gen client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums, accSeqs []uint64,
	signers, signatures []cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(signers))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))
	signMode, err := authsigning.APISignModeToInternal(gen.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range signers {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	baseTxBuilder := gen.NewTxBuilder()

	txBuilder, ok := baseTxBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, fmt.Errorf("expected authtx.ExtensionOptionsTxBuilder, got %T", baseTxBuilder)
	}
	if len(selectedAuthenticators) > 0 {
		value, err := types.NewAnyWithValue(&smartaccounttypes.TxExtension{
			SelectedAuthenticators: selectedAuthenticators,
		})
		if err != nil {
			return nil, err
		}
		txBuilder.SetNonCriticalExtensionOptions(value)
	}

	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)
	// TODO: set fee payer

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range signatures {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := authsigning.GetSignBytesAdapter(
			ctx, gen.SignModeHandler(), signMode, signerData, txBuilder.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = txBuilder.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return txBuilder.GetTx(), nil
}
