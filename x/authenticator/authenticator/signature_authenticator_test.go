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
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/stretchr/testify/suite"
)

type SigVerifyAuthenticationSuite struct {
	suite.Suite
	OsmosisApp                   *app.OsmosisApp
	Ctx                          sdk.Context
	EncodingConfig               params.EncodingConfig
	SigVerificationAuthenticator authenticator.SignatureVerificationAuthenticator
	TestKeys                     []string
	TestAccAddress               []sdk.AccAddress
	TestPrivKeys                 []*secp256k1.PrivKey
}

func TestSigVerifyAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(SigVerifyAuthenticationSuite))
}

func (s *SigVerifyAuthenticationSuite) SetupTest() {
	// Test data for authenticator signature verification
	TestKeys := []string{
		"6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159",
		"0dd4d1506e18a5712080708c338eb51ecf2afdceae01e8162e890b126ac190fe",
		"49006a359803f0602a7ec521df88bf5527579da79112bb71f285dd3e7d438033",
	}
	s.EncodingConfig = app.MakeEncodingConfig()
	txConfig := s.EncodingConfig.TxConfig
	signModeHandler := txConfig.SignModeHandler()

	s.OsmosisApp = app.Setup(false)

	ak := s.OsmosisApp.AccountKeeper
	s.Ctx = s.OsmosisApp.NewContext(false, tmproto.Header{})
	s.Ctx = s.Ctx.WithGasMeter(sdk.NewGasMeter(1_000_000))

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

	// Create a new Secp256k1SignatureAuthenticator for testing
	s.SigVerificationAuthenticator = authenticator.NewSignatureVerificationAuthenticator(
		ak,
		signModeHandler,
	)
}

type SignatureVerificationAuthenticatorTestData struct {
	Msgs                               []sdk.Msg
	AccNums                            []uint64
	AccSeqs                            []uint64
	Signers                            []cryptotypes.PrivKey
	Signatures                         []cryptotypes.PrivKey
	NumberOfExpectedSigners            int
	NumberOfExpectedSignatures         int
	ShouldSucceedGettingData           bool
	ShouldSucceedSignatureVerification bool
}

type SignatureVerificationAuthenticatorTest struct {
	Description string
	TestData    SignatureVerificationAuthenticatorTestData
}

// TestSignatureAuthenticator test a non-smart account signature verification
func (s *SigVerifyAuthenticationSuite) TestSignatureAuthenticator() {
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
	testMsg3 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[2]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	testMsg4 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	tests := []SignatureVerificationAuthenticatorTest{
		{
			Description: "Successfully verified authenticator",
			TestData: SignatureVerificationAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
					testMsg2,
					testMsg3,
				},
				[]uint64{0, 0, 0, 0},
				[]uint64{0, 0, 0, 0},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
					s.TestPrivKeys[2],
				},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
					s.TestPrivKeys[2],
				},
				3,
				3,
				true,
				true,
			},
		},
		{
			Description: "Test: unsuccessful signature authentication not enough signers: FAIL",
			TestData: SignatureVerificationAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
					testMsg2,
					testMsg3,
					testMsg4,
				},
				[]uint64{0, 0, 0, 0},
				[]uint64{0, 0, 0, 0},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
					s.TestPrivKeys[2],
					s.TestPrivKeys[3],
				},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[2],
				},
				0,
				0,
				false,
				false,
			},
		},
		{
			Description: "Test: unsuccessful signature authentication not enough signatures: FAIL",
			TestData: SignatureVerificationAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
					testMsg2,
					testMsg3,
					testMsg4,
				},
				[]uint64{0, 0},
				[]uint64{0, 0},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
				},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[2],
				},
				0,
				0,
				false,
				false,
			},
		},
		{
			Description: "Test: unsuccessful signature authentication invalid signatures: FAIL",
			TestData: SignatureVerificationAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
					testMsg2,
				},
				[]uint64{0, 0},
				[]uint64{0, 0},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
				},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[2],
				},
				2,
				2,
				true,
				false,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.Description, func() {
			// Generate a transaction based on the test cases
			tx, _ := GenTx(
				s.EncodingConfig.TxConfig,
				tc.TestData.Msgs,
				feeCoins,
				300000,
				"",
				tc.TestData.AccNums,
				tc.TestData.AccSeqs,
				tc.TestData.Signers,
				tc.TestData.Signatures,
			)

			if tc.TestData.ShouldSucceedGettingData {
				// Test GetAuthenticationData
				authData, err := s.SigVerificationAuthenticator.GetAuthenticationData(s.Ctx, tx, -1, false)
				s.Require().NoError(err)

				// cast the interface as a concrete struct
				sigData := authData.(authenticator.SignatureData)

				// the signer data should contain x signers
				s.Require().Equal(len(sigData.Signers), tc.TestData.NumberOfExpectedSigners)

				// the signature data should contain x signatures
				s.Require().Equal(len(sigData.Signatures), tc.TestData.NumberOfExpectedSignatures)

				// Test Authenticate method
				if tc.TestData.ShouldSucceedSignatureVerification {
					success, err := s.SigVerificationAuthenticator.Authenticate(s.Ctx, nil, authData)
					s.Require().NoError(err)
					s.Require().True(success)

				} else {
					// TODO: check error here
					success, _ := s.SigVerificationAuthenticator.Authenticate(s.Ctx, nil, authData)
					s.Require().False(success)
				}
			} else {
				authData, err := s.SigVerificationAuthenticator.GetAuthenticationData(s.Ctx, tx, -1, false)
				s.Require().Error(err)

				// cast the interface as a concrete struct
				sigData := authData.(authenticator.SignatureData)

				// the signer data should contain x signers
				s.Require().Equal(len(sigData.Signers), tc.TestData.NumberOfExpectedSigners)

				// the signature data should contain x signatures
				s.Require().Equal(len(sigData.Signatures), tc.TestData.NumberOfExpectedSignatures)

			}
		})
	}
}

func (s *SigVerifyAuthenticationSuite) TestMultiSignatureAuthenticator() {
	// TODO: test multi sig
	s.Require().True(true)
	//osmoToken := "osmo"
	//coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	//// Create a test messages for signing
	//testMsg1 := &banktypes.MsgSend{
	//	FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
	//	ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
	//	Amount:      coins,
	//}
	//testMsg2 := &banktypes.MsgSend{
	//	FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
	//	ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
	//	Amount:      coins,
	//}
	//feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	//tx, err := GenValidTx(
	//	s.EncodingConfig.TxConfig,
	//	[]sdk.Msg{
	//		testMsg1,
	//		testMsg2,
	//	},
	//	feeCoins,
	//	300000,
	//	"",
	//	[]uint64{0, 0},
	//	[]uint64{0, 0},
	//	s.TestPrivKeys[0],
	//	s.TestPrivKeys[1],
	//)

	//// Test GetAuthenticationData
	//authData, err := s.SigVerificationAuthenticator.GetAuthenticationData(s.Ctx, tx, -1, false)
	//s.Require().NoError(err)

	//fmt.Println(s.Ctx.GasMeter().GasConsumed())

	//// the signer data should contain 2 signers
	//sigData := authData.(authenticator.SignatureData)
	//s.Require().Equal(len(sigData.Signers), 2)

	//// the signature data should contain 2 signatures
	//s.Require().Equal(len(sigData.Signatures), 2)

	//// Test Authenticate method
	//success, err := s.SigVerificationAuthenticator.Authenticate(s.Ctx, nil, authData)
	//s.Require().True(success)
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
	signers []cryptotypes.PrivKey,
	signatures []cryptotypes.PrivKey,
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
	for i, p := range signatures {
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

// GenTx generates a signed mock transaction.
func GenValidMultiSigTx(
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
		_, err = p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		// TODO: multisignature data
		//sigs[i].Data.(*signing.MultiSignatureData).BitArray = sig
		//sigs[i].Data.(*signing.MultiSignatureData).Signatures = []sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}
