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

// AuthenticatorCircuitBreakerAnteSuite is a test suite for the authenticator and CircuitBreaker AnteDecorator.
type AuthenticatorCircuitBreakerAnteSuite struct {
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

// TestAuthenticatorCircuitBreakerAnteSuite runs the test suite for the authenticator and CircuitBreaker AnteDecorator.
func TestAuthenticatorCircuitBreakerAnteSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorCircuitBreakerAnteSuite))
}

// SetupTest initializes the test data and prepares the test environment.
func (s *AuthenticatorCircuitBreakerAnteSuite) SetupTest() {
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

func (s *AuthenticatorCircuitBreakerAnteSuite) TearDownTest() {
	os.RemoveAll(s.HomeDir)
}

// MockAnteDecorator used to test the CircuitBreaker flow
type MockAnteDecorator struct {
	Called int
}

// AnteHandle increments the ctx.Priority() differently based on what flow is active
func (m MockAnteDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	prio := ctx.Priority()

	if m.Called == 1 {
		return ctx.WithPriority(prio + 1), nil
	} else {
		return ctx.WithPriority(prio + 2), nil
	}
}

// TestCircuitBreakerAnte verifies that the CircuitBreaker AnteDecorator functions correctly.
func (s *AuthenticatorCircuitBreakerAnteSuite) TestCircuitBreakerAnte() {
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

	mockTestClassic := MockAnteDecorator{Called: 1}
	mockTestAuthenticator := MockAnteDecorator{Called: 0}

	// Create a CircuitBreaker AnteDecorator
	cbd := ante.NewCircuitBreakerDecorator(
		s.OsmosisApp.SmartAccountKeeper,
		sdk.ChainAnteDecorators(mockTestAuthenticator),
		sdk.ChainAnteDecorators(mockTestClassic),
	)
	anteHandler := sdk.ChainAnteDecorators(cbd)

	// Deactivate smart accounts
	params := s.OsmosisApp.SmartAccountKeeper.GetParams(s.Ctx)
	params.IsSmartAccountActive = false
	s.OsmosisApp.SmartAccountKeeper.SetParams(s.Ctx, params)

	// Here we test when smart accounts are deactivated
	ctx, err := anteHandler(s.Ctx, tx, false)
	s.Require().NoError(err)
	s.Require().Equal(int64(1), ctx.Priority(), "Should have disabled the full authentication flow")

	// Reactivate smart accounts
	params = s.OsmosisApp.SmartAccountKeeper.GetParams(ctx)
	params.IsSmartAccountActive = true
	s.OsmosisApp.SmartAccountKeeper.SetParams(ctx, params)

	// Here we test when smart accounts are active and there is not selected authenticator
	ctx, err = anteHandler(ctx, tx, false)
	s.Require().Equal(int64(2), ctx.Priority(), "Will only go this way when a TxExtension is not included in the tx")
	s.Require().NoError(err)

	// Generate a test transaction with a selected authenticator
	tx, _ = GenTx(s.Ctx, s.EncodingConfig.TxConfig, []sdk.Msg{
		testMsg1,
		testMsg2,
	}, feeCoins, 300000, "", []uint64{0, 0}, []uint64{0, 0}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []cryptotypes.PrivKey{
		s.TestPrivKeys[0],
		s.TestPrivKeys[1],
	}, []uint64{1})

	// Test is smart accounts are active and the authenticator flow is selected
	ctx, err = anteHandler(ctx, tx, false)
	s.Require().NoError(err)
	s.Require().Equal(int64(4), ctx.Priority(), "Should have used the full authentication flow")
}
