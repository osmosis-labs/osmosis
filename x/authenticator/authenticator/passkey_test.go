package authenticator_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	// multisig

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v21/app"
	"github.com/osmosis-labs/osmosis/v21/app/params"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"
)

type PassKeyAuthenticationSuite struct {
	suite.Suite
	OsmosisApp           *app.OsmosisApp
	Ctx                  sdk.Context
	EncodingConfig       params.EncodingConfig
	PassKeyAuthenticator authenticator.PassKeyAuthenticator
	TestKeys             []string
	TestAccAddress       []sdk.AccAddress
	TestPrivKeys         []*secp256r1.PrivKey
}

func TestPassKeyAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(PassKeyAuthenticationSuite))
}

func (s *PassKeyAuthenticationSuite) SetupTest() {
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
	for i := 0; i < len(TestKeys); i++ {
		priv, _ := secp256r1.GenPrivKey()

		// add the test private keys to array for later use
		s.TestPrivKeys = append(s.TestPrivKeys, priv)

		accAddress := sdk.AccAddress(priv.PubKey().Address())
		account := authtypes.NewBaseAccount(accAddress, nil, 0, 0)
		ak.SetAccount(s.Ctx, account)

		// add the test accounts to array for later use
		s.TestAccAddress = append(s.TestAccAddress, accAddress)
	}

	// Create a new Secp256k1SignatureAuthenticator for testing
	s.PassKeyAuthenticator = authenticator.NewPassKeyAuthenticator(
		ak,
		signModeHandler,
	)
	s.OsmosisApp.AuthenticatorManager.RegisterAuthenticator(s.PassKeyAuthenticator)

	err := s.OsmosisApp.AuthenticatorKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[0],
		"PassKeyAuthenticator",
		s.TestPrivKeys[0].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	err = s.OsmosisApp.AuthenticatorKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[1],
		"PassKeyAuthenticator",
		s.TestPrivKeys[1].PubKey().Bytes(),
	)
	s.Require().NoError(err)
	err = s.OsmosisApp.AuthenticatorKeeper.AddAuthenticator(
		s.Ctx,
		s.TestAccAddress[2],
		"PassKeyAuthenticator",
		s.TestPrivKeys[2].PubKey().Bytes(),
	)
	s.Require().NoError(err)
}

type PassKeyAuthenticatorTestData struct {
	Msgs                       []sdk.Msg
	AccNums                    []uint64
	AccSeqs                    []uint64
	Signers                    []cryptotypes.PrivKey
	Signatures                 []cryptotypes.PrivKey
	NumberOfExpectedSigners    int
	NumberOfExpectedSignatures int
	ShouldSucceedPassKey       bool
}

type PassKeyAuthenticatorTest struct {
	Description string
	TestData    PassKeyAuthenticatorTestData
}

// TestSignatureAuthenticator test a non-smart account signature verification
func (s *PassKeyAuthenticationSuite) TestSignatureAuthenticator() {
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
	//testMsg4 := &banktypes.MsgSend{
	//	FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[0]),
	//	ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[1]),
	//	Amount:      coins,
	//}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	tests := []PassKeyAuthenticatorTest{
		{
			Description: "Test: successfully verified authenticator with one signer: base case: PASS",
			TestData: PassKeyAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
				},
				[]uint64{0},
				[]uint64{0},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
				},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
				},
				1,
				1,
				true,
			},
		},
		{
			Description: "Test: successfully verified authenticator: multiple signers: PASS",
			TestData: PassKeyAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
					testMsg2,
					testMsg3,
				},
				[]uint64{0, 0, 0},
				[]uint64{0, 0, 0},
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
			},
		},
		{
			// This test case tests if there is two messages with the same signer
			// with two successful signatures.
			Description: "Test: verified authenticator with 2 messages signed correctly with the same address: PASS",
			TestData: PassKeyAuthenticatorTestData{
				[]sdk.Msg{
					testMsg1,
					testMsg2,
					testMsg2,
				},
				[]uint64{0, 0, 0},
				[]uint64{0, 0, 0},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
					s.TestPrivKeys[1],
				},
				[]cryptotypes.PrivKey{
					s.TestPrivKeys[0],
					s.TestPrivKeys[1],
					s.TestPrivKeys[1],
				},
				2,
				2,
				true,
			},
		},
		// TODO: This is failing because of the tx builder. Need to fix the test helper
		//{
		//	Description: "Test: unsuccessful signature authentication not enough signers: FAIL",
		//	TestData: PassKeyAuthenticatorTestData{
		//		[]sdk.Msg{
		//			testMsg1,
		//			testMsg2,
		//			testMsg3,
		//			testMsg4,
		//		},
		//		[]uint64{0, 0, 0, 0},
		//		[]uint64{0, 0, 0, 0},
		//		[]cryptotypes.PrivKey{
		//			s.TestPrivKeys[0],
		//			s.TestPrivKeys[1],
		//			s.TestPrivKeys[1],
		//			s.TestPrivKeys[2],
		//		},
		//		[]cryptotypes.PrivKey{
		//			s.TestPrivKeys[0],
		//			s.TestPrivKeys[2],
		//		},
		//		3,
		//		3,
		//		false,
		//	},
		//},

		// TODO: This is failing because of the tx builder. Need to fix the test helper
		//{
		//	Description: "Test: unsuccessful signature authentication not enough signatures: FAIL",
		//	TestData: PassKeyAuthenticatorTestData{
		//		[]sdk.Msg{
		//			testMsg1,
		//			testMsg2,
		//			testMsg3,
		//		},
		//		[]uint64{0, 0, 0},
		//		[]uint64{0, 0, 0},
		//		[]cryptotypes.PrivKey{
		//			s.TestPrivKeys[0],
		//			s.TestPrivKeys[1],
		//			s.TestPrivKeys[1],
		//		},
		//		[]cryptotypes.PrivKey{
		//			s.TestPrivKeys[0],
		//			s.TestPrivKeys[2],
		//		},
		//		3,
		//		3,
		//		false,
		//	},
		//},
		//{
		//	Description: "Test: unsuccessful signature authentication invalid signatures: FAIL",
		//	TestData: PassKeyAuthenticatorTestData{
		//		[]sdk.Msg{
		//			testMsg2,
		//		},
		//		[]uint64{0, 0},
		//		[]uint64{0, 0},
		//		[]cryptotypes.PrivKey{
		//			s.TestPrivKeys[1],
		//		},
		//		[]cryptotypes.PrivKey{
		//			s.TestPrivKeys[2],
		//		},
		//		1,
		//		1,
		//		false,
		//	},
		//},
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

			// Test GetAuthenticationData
			//authData, err := s.PassKeyAuthenticator.GetAuthenticationData(s.Ctx, tx, -1, false)
			//s.Require().NoError(err)
			//
			//// cast the interface as a concrete struct
			//sigData := authData.(authenticator.SignatureData)
			//
			//// the signer data should contain x signers
			//s.Require().Equal(tc.TestData.NumberOfExpectedSigners, len(sigData.Signers))

			// the signature data should contain x signatures
			//s.Require().Equal(tc.TestData.NumberOfExpectedSignatures, len(sigData.Signatures))

			ak := s.OsmosisApp.AccountKeeper
			sigModeHandler := s.EncodingConfig.TxConfig.SignModeHandler()

			// Test Authenticate method
			var success iface.AuthenticationResult
			for i, msg := range tx.GetMsgs() {
				accAddress := sdk.AccAddress(tc.TestData.Signers[i].PubKey().Address())
				allAuthenticators, err := s.OsmosisApp.AuthenticatorKeeper.GetAuthenticatorsForAccountOrDefault(s.Ctx, accAddress)
				s.Require().NoError(err)

				for _, a11r := range allAuthenticators {
					// Get the authentication data for the transaction
					request, err := authenticator.GenerateAuthenticationData(s.Ctx, ak, sigModeHandler, accAddress, msg, tx, i, false, authenticator.SequenceMatch)
					s.Require().NoError(err)

					success = a11r.Authenticator.Authenticate(s.Ctx, request)
				}
			}
			if tc.TestData.ShouldSucceedPassKey {
				s.Require().True(success.IsAuthenticated())
			} else {
				s.Require().False(success.IsAuthenticated())
			}
		})
	}
}
