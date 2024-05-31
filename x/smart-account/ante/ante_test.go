package ante_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v25/x/txfees/keeper"

	"github.com/cosmos/cosmos-sdk/codec/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	smartaccounttypes "github.com/osmosis-labs/osmosis/v25/x/smart-account/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v25/app"
	"github.com/osmosis-labs/osmosis/v25/app/params"
	"github.com/osmosis-labs/osmosis/v25/x/smart-account/ante"
	"github.com/osmosis-labs/osmosis/v25/x/smart-account/testutils"
)

type AuthenticatorAnteSuite struct {
	suite.Suite
	OsmosisApp             *app.OsmosisApp
	Ctx                    sdk.Context
	EncodingConfig         params.EncodingConfig
	AuthenticatorDecorator ante.AuthenticatorDecorator
	TestKeys               []string
	TestAccAddress         []sdk.AccAddress
	TestPrivKeys           []*secp256k1.PrivKey
}

func TestAuthenticatorAnteSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorAnteSuite))
}

func (s *AuthenticatorAnteSuite) SetupTest() {
	// Test data for authenticator signature verification
	TestKeys := []string{
		"6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159",
		"0dd4d1506e18a5712080708c338eb51ecf2afdceae01e8162e890b126ac190fe",
		"49006a359803f0602a7ec521df88bf5527579da79112bb71f285dd3e7d438033",
	}
	s.EncodingConfig = app.MakeEncodingConfig()

	s.OsmosisApp = app.Setup(false)

	ak := s.OsmosisApp.AccountKeeper
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})

	// Set up test accounts
	for _, key := range TestKeys {
		bz, _ := hex.DecodeString(key)
		priv := &secp256k1.PrivKey{Key: bz}

		// add the test private keys to array for later use
		s.TestPrivKeys = append(s.TestPrivKeys, priv)

		accAddress := sdk.AccAddress(priv.PubKey().Address())
		account := authtypes.NewBaseAccount(accAddress, priv.PubKey(), 0, 0)
		ak.SetAccount(s.Ctx, account)

		// add the test accounts to array for later use
		s.TestAccAddress = append(s.TestAccAddress, accAddress)
	}

	deductFeeDecorator := txfeeskeeper.NewDeductFeeDecorator(*s.OsmosisApp.TxFeesKeeper, s.OsmosisApp.AccountKeeper, s.OsmosisApp.BankKeeper, nil)

	s.AuthenticatorDecorator = ante.NewAuthenticatorDecorator(
		s.OsmosisApp.SmartAccountKeeper,
		s.OsmosisApp.AccountKeeper,
		s.EncodingConfig.TxConfig.SignModeHandler(),
		deductFeeDecorator,
	)
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))
}

// TestSignatureVerificationNoAuthenticatorInStore test a non-smart account signature verification
// with no authenticator in the store
func (s *AuthenticatorAnteSuite) TestSignatureVerificationNoAuthenticatorInStore() {
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

	tx, _ := GenTx(s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
		testMsg2,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []uint64{0, 0})

	anteHandler := sdk.ChainAnteDecorators(s.AuthenticatorDecorator)
	_, err := anteHandler(s.Ctx, tx, false)

	s.Require().Error(err, "Expected error when no authenticator is in the store")
}

// TestSignatureVerificationWithAuthenticatorInStore test a non-smart account signature verification
// with a single authenticator in the store
func (s *AuthenticatorAnteSuite) TestSignatureVerificationWithAuthenticatorInStore() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Ensure the feepayer has funds
	fees := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}
	feePayer := s.TestPrivKeys[0].PubKey().Address()
	err := testutil.FundAccount(s.OsmosisApp.BankKeeper, s.Ctx, feePayer.Bytes(), fees)
	s.Require().NoError(err)

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

	tx, _ := GenTx(s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
		testMsg2,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []uint64{1, 2})

	anteHandler := sdk.ChainAnteDecorators(s.AuthenticatorDecorator)
	_, err = anteHandler(s.Ctx, tx, false)

	s.Require().NoError(err)
}

// TestSignatureVerificationOutOfGas tests that the ante handler exits early by running out of gas if the
// fee payer has not been authenticated before consuming the parametrized max unauthenticated gas limit (even if the specified limit is 300k)
// This is to ensure that the amount of compute a non-authenticated user can execute is limited.
func (s *AuthenticatorAnteSuite) TestSignatureVerificationOutOfGas() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	maxUnauthenticatedGasLimit := s.OsmosisApp.SmartAccountKeeper.GetParams(s.Ctx).MaximumUnauthenticatedGas
	specifiedGasLimit := uint64(300_000)

	// Ensure the feepayers have funds
	fees := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}
	err := testutil.FundAccount(s.OsmosisApp.BankKeeper, s.Ctx, s.TestPrivKeys[0].PubKey().Address().Bytes(), fees)
	s.Require().NoError(err)
	err = testutil.FundAccount(s.OsmosisApp.BankKeeper, s.Ctx, s.TestPrivKeys[1].PubKey().Address().Bytes(), fees)
	s.Require().NoError(err)

	// This message will have several authenticators for s.TestPrivKeys[0] and one for s.TestPrivKeys[1] at the end
	testMsg1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}

	// fee payer is authenticated
	sigId, err := s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[1],
		"SignatureVerification",
		s.TestPrivKeys[1].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(sigId, uint64(1), "Adding authenticator returning incorrect id")

	alwaysHigher := testutils.TestingAuthenticator{Approve: testutils.Always, GasConsumption: int(maxUnauthenticatedGasLimit + 1)}
	s.OsmosisApp.AuthenticatorManager.RegisterAuthenticator(alwaysHigher)

	excessGasId, err := s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[0],
		alwaysHigher.Type(),
		[]byte{},
	)
	s.Require().NoError(err)
	s.Require().Equal(excessGasId, uint64(2), "Adding authenticator returning incorrect id")

	tx, _ := GenTx(s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
	}, feeCoins, specifiedGasLimit, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[1],
	}, []uint64{excessGasId})

	anteHandler := sdk.ChainAnteDecorators(s.AuthenticatorDecorator)
	_, err = anteHandler(s.Ctx, tx, false)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "FeePayer must be authenticated first because gas consumption has exceeded the free gas limit for authentication process. The gas limit has been reduced to 120000. Gas consumed: ")
	// Now, let's ensure the fee payer has been authenticated before checking all authenticators for s.TestPrivKeys[1]

	// This is a message that can only be aithenticated by its default authenticator (s.TestAccAddress[1])
	testMsg2 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}

	// Authenticate the fee payer and check gas limit is raised
	tx, _ = GenTx(s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg2,
		testMsg1,
	}, feeCoins, specifiedGasLimit, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[1],
		s.TestPrivKeys[0],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[1],
		s.TestPrivKeys[1],
	}, []uint64{sigId, excessGasId})

	// This authentication should succeed and consume gas over the max unauthenticated gas limit (because the fee payer
	// is authenticated in under max unauthenticated gas limit in the first message)
	res, err := anteHandler(s.Ctx, tx, false)
	s.Require().NoError(err)
	s.Require().Greater(res.GasMeter().GasConsumed(), maxUnauthenticatedGasLimit)
}

func (s *AuthenticatorAnteSuite) TestSpecificAuthenticator() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create a test messages for signing
	testMsg1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	sig1Id, err := s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[1],
		"SignatureVerification",
		s.TestPrivKeys[0].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(sig1Id, uint64(1), "Adding authenticator returning incorrect id")

	sig2Id, err := s.OsmosisApp.SmartAccountKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[1],
		"SignatureVerification",
		s.TestPrivKeys[1].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(sig2Id, uint64(2), "Adding authenticator returning incorrect id")

	testCases := []struct {
		name                  string
		senderKey             cryptotypes.PrivKey
		signKey               cryptotypes.PrivKey
		selectedAuthenticator []uint64
		shouldPass            bool
		checks                int
		checkGas              bool
	}{
		{"Correct authenticator 0", s.TestPrivKeys[0], s.TestPrivKeys[0], []uint64{sig1Id}, true, 1, true},
		{"Correct authenticator 1", s.TestPrivKeys[0], s.TestPrivKeys[1], []uint64{sig2Id}, true, 1, true},
		{"Incorrect authenticator 0", s.TestPrivKeys[0], s.TestPrivKeys[0], []uint64{sig2Id}, false, 1, true},
		{"Incorrect authenticator 1", s.TestPrivKeys[0], s.TestPrivKeys[1], []uint64{sig1Id}, false, 1, true},
		{"Not Specified for 0", s.TestPrivKeys[0], s.TestPrivKeys[0], []uint64{}, false, 0, true},
		{"Not Specified for 1", s.TestPrivKeys[0], s.TestPrivKeys[1], []uint64{}, false, 0, true},
		{"Bad selection", s.TestPrivKeys[0], s.TestPrivKeys[0], []uint64{3}, false, 0, false},
	}

	baseGas := 2891              // base gas consimed before starting to iterate through authenticators
	approachingGasPerSig := 5429 // Each signature consumes at least this amount (but not much more)

	// Ensure the feepayer has funds
	fees := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2_500_000)}
	err = testutil.FundAccount(s.OsmosisApp.BankKeeper, s.Ctx, s.TestPrivKeys[0].PubKey().Address().Bytes(), fees)
	s.Require().NoError(err)
	err = testutil.FundAccount(s.OsmosisApp.BankKeeper, s.Ctx, s.TestPrivKeys[1].PubKey().Address().Bytes(), fees)
	s.Require().NoError(err)

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tx, _ := GenTx(s.EncodingConfig.TxConfig, []sdk.Msg{
				testMsg1,
			}, feeCoins, 300000, "", []uint64{0}, []uint64{0}, []cryptotypes.PrivKey{
				tc.senderKey,
			}, []cryptotypes.PrivKey{
				tc.signKey,
			},
				tc.selectedAuthenticator,
			)

			anteHandler := sdk.ChainAnteDecorators(s.AuthenticatorDecorator)
			res, err := anteHandler(s.Ctx.WithGasMeter(sdk.NewGasMeter(300000)), tx, false)

			if tc.shouldPass {
				s.Require().NoError(err, "Expected to pass but got error")
			} else {
				s.Require().Error(err, "Expected to fail but got no error")
			}

			// ensure only the right amount of sigs have been checked
			if tc.checks > 0 {
				s.Require().GreaterOrEqual(res.GasMeter().GasConsumed(), uint64(baseGas+(tc.checks-1)*approachingGasPerSig))
				s.Require().LessOrEqual(res.GasMeter().GasConsumed(), uint64(baseGas+tc.checks*approachingGasPerSig))
			} else {
				if tc.checkGas {
					s.Require().LessOrEqual(res.GasMeter().GasConsumed(), uint64(baseGas))
				}
			}
		})
	}
}

// GenTx generates a signed mock transaction.
func GenTx(
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
	signMode := gen.SignModeHandler().DefaultMode()

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

	err := txBuilder.SetMsgs(msgs...)
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

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range signatures {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
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
