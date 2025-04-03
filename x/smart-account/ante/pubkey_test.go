package ante_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/ante"
)

// AuthenticatorSetPubKeyAnteSuite is a test suite for the authenticator and SetPubKey AnteDecorator.
type AuthenticatorSetPubKeyAnteSuite struct {
	suite.Suite
	OsmosisApp             *app.OsmosisApp
	Ctx                    sdk.Context
	EncodingConfig         params.EncodingConfig
	AuthenticatorDecorator ante.AuthenticatorDecorator
	TestKeys               []string
	TestAccAddress         []sdk.AccAddress
	TestPrivKeys           []*secp256k1.PrivKey
	HomeDir                string
}

// TestAuthenticatorSetPubKeyAnteSuite runs the test suite for the authenticator and SetPubKey AnteDecorator.
func TestAuthenticatorSetPubKeyAnteSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorSetPubKeyAnteSuite))
}

// SetupTest initializes the test data and prepares the test environment.
func (s *AuthenticatorSetPubKeyAnteSuite) SetupTest() {
	// Test data for authenticator signature verification
	TestKeys := []string{
		"6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159",
		"0dd4d1506e18a5712080708c338eb51ecf2afdceae01e8162e890b126ac190fe",
		"49006a359803f0602a7ec521df88bf5527579da79112bb71f285dd3e7d438033",
		"05d2f57e30fb44835da1cad5274cefd4c80f6652c425fb9e6cc9c6749126497c",
		"f98d0b79c0cc9805b905bfc5104f31293a270e60c6fc613a037eeb484fddb974",
	}
	s.EncodingConfig = app.MakeEncodingConfig()

	// Initialize the Osmosis application
	s.HomeDir = fmt.Sprintf("%d", rand.Int())
	s.OsmosisApp = app.SetupWithCustomHome(false, s.HomeDir)

	s.Ctx = s.OsmosisApp.NewContextLegacy(false, tmproto.Header{})

	// Set up test accounts
	for _, key := range TestKeys {
		bz, _ := hex.DecodeString(key)
		priv := &secp256k1.PrivKey{Key: bz}

		// Add the test private keys to an array for later use
		s.TestPrivKeys = append(s.TestPrivKeys, priv)

		// Generate an account address from the public key
		accAddress := sdk.AccAddress(priv.PubKey().Address())

		// Create a new BaseAccount for the test account
		authtypes.NewBaseAccount(accAddress, nil, 0, 0)

		// Add the test accounts' addresses to an array for later use
		s.TestAccAddress = append(s.TestAccAddress, accAddress)
	}
}

func (s *AuthenticatorSetPubKeyAnteSuite) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

// TestSetPubKeyAnte verifies that the SetPubKey AnteDecorator functions correctly.
func (s *AuthenticatorSetPubKeyAnteSuite) TestSetPubKeyAnte() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create test messages for signing
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

	// Generate a test transaction
	tx, _ := GenTx(s.Ctx, s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
		testMsg2,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []uint64{})

	// Create a SetPubKey AnteDecorator
	spkd := ante.NewEmitPubKeyDecoratorEvents(s.OsmosisApp.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(spkd)

	// Run the AnteDecorator on the transaction
	_, err := antehandler(s.Ctx, tx, false)
	s.Require().NoError(err)
}

// TestSetPubKeyAnteWithSenderNotSigner verifies that SetPubKey AnteDecorator correctly handles a non-signer sender.
func (s *AuthenticatorSetPubKeyAnteSuite) TestSetPubKeyAnteWithSenderNotSigner() {
	osmoToken := "osmo"
	coins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Create a test message with a sender that is not a signer
	testMsg1 := &banktypes.MsgSend{
		FromAddress: sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[4]),
		ToAddress:   sdk.MustBech32ifyAddressBytes(osmoToken, s.TestAccAddress[3]),
		Amount:      coins,
	}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(osmoToken, 2500)}

	// Generate a test transaction
	tx, _ := GenTx(s.Ctx, s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[3],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[3],
	}, []uint64{})

	// Create a SetPubKey AnteDecorator
	spkd := ante.NewEmitPubKeyDecoratorEvents(s.OsmosisApp.AccountKeeper)
	antehandler := sdk.ChainAnteDecorators(spkd)

	// Run the AnteDecorator on the transaction
	ctx, err := antehandler(s.Ctx, tx, false)
	s.Require().NoError(err)

	// Ensure that the public key has not been set for a non-signer sender
	pk, err := s.OsmosisApp.AccountKeeper.GetPubKey(ctx, s.TestAccAddress[4])
	s.Require().Equal(pk, nil, "Public Key has not been set")
}
