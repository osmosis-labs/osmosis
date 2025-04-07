package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/testutils"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	am *authenticator.AuthenticatorManager
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Reset()
	s.am = authenticator.NewAuthenticatorManager()

	// Register the SigVerificationAuthenticator
	s.am.InitializeAuthenticators([]authenticator.Authenticator{
		authenticator.SignatureVerification{},
		testutils.TestingAuthenticator{
			Approve:        testutils.Always,
			GasConsumption: 10,
			Confirm:        testutils.Always,
		},
	})
}

func (s *KeeperTestSuite) TestKeeper_AddAuthenticator() {
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	id, err := s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err, "Should successfully add a SignatureVerification")
	s.Require().Equal(id, uint64(1), "Adding authenticator returning incorrect id")

	id, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"MessageFilter",
		[]byte(`{"@type":"/cosmos.bank.v1beta1.MsgSend"}`),
	)
	s.Require().NoError(err, "Should successfully add a MessageFilter")
	s.Require().Equal(id, uint64(2), "Adding authenticator returning incorrect id")

	_, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		[]byte("BrokenBytes"),
	)
	s.Require().Error(err, "Should have failed as OnAuthenticatorAdded fails")

	s.App.AuthenticatorManager.ResetAuthenticators()
	_, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"MessageFilter",
		[]byte(`{"@type":"/cosmos.bank.v1beta1.MsgSend"`),
	)
	s.Require().Error(err, "Authenticator not registered so should fail")
}

func (s *KeeperTestSuite) TestKeeper_GetAuthenticatorDataForAccount() {
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	_, err := s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err, "Should successfully add a SignatureVerification")

	_, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err, "Should successfully add a MessageFilter")

	authenticators, err := s.App.SmartAccountKeeper.GetAuthenticatorDataForAccount(ctx, accAddress)
	s.Require().NoError(err)
	s.Require().Equal(len(authenticators), 2, "Getting authenticators returning incorrect data")
}

func (s *KeeperTestSuite) TestKeeper_GetAndSetAuthenticatorId() {
	ctx := s.Ctx

	authenticatorId := s.App.SmartAccountKeeper.InitializeOrGetNextAuthenticatorId(ctx)
	s.Require().Equal(uint64(1), authenticatorId, "Initialize/Get authenticator id returned incorrect id")

	authenticatorId = s.App.SmartAccountKeeper.InitializeOrGetNextAuthenticatorId(ctx)
	s.Require().Equal(uint64(1), authenticatorId, "Initialize/Get authenticator id returned incorrect id")

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	_, err := s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err, "Should successfully add a SignatureVerification")

	authenticatorId = s.App.SmartAccountKeeper.InitializeOrGetNextAuthenticatorId(ctx)
	s.Require().Equal(authenticatorId, uint64(2), "Initialize/Get authenticator id returned incorrect id")

	_, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err, "Should successfully add a MessageFilter")

	authenticatorId = s.App.SmartAccountKeeper.InitializeOrGetNextAuthenticatorId(ctx)
	s.Require().Equal(authenticatorId, uint64(3), "Initialize/Get authenticator id returned incorrect id")
}

func (s *KeeperTestSuite) TestKeeper_GetSelectedAuthenticatorForAccount() {
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	_, err := s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerification",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err, "Should successfully add a SignatureVerification")

	_, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"MessageFilter",
		[]byte(`{"@type":"/cosmos.bank.v1beta1.MsgSend"}`),
	)
	s.Require().NoError(err, "Should successfully add a MessageFilter")

	// Test getting a selected authenticator from the store
	selectedAuthenticator, err := s.App.SmartAccountKeeper.GetInitializedAuthenticatorForAccount(ctx, accAddress, 2)
	s.Require().NoError(err)
	s.Require().Equal(selectedAuthenticator.Authenticator.Type(), "MessageFilter", "Getting authenticators returning incorrect data")

	selectedAuthenticator, err = s.App.SmartAccountKeeper.GetInitializedAuthenticatorForAccount(ctx, accAddress, 1)
	s.Require().NoError(err)
	s.Require().Equal(selectedAuthenticator.Authenticator.Type(), "SignatureVerification", "Getting authenticators returning incorrect data")
	s.Require().Equal(selectedAuthenticator.Id, uint64(1), "Incorrect ID returned from store")

	_, err = s.App.SmartAccountKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"MessageFilter",
		[]byte(`{"@type":"/cosmos.bank.v1beta1.MsgSend"}`),
	)
	s.Require().NoError(err, "Should successfully add a MessageFilter")

	// Remove a registered authenticator from the authenticator manager
	s.App.AuthenticatorManager.UnregisterAuthenticator(authenticator.MessageFilter{})

	// Try to get an authenticator that has been removed from the store
	selectedAuthenticator, err = s.App.SmartAccountKeeper.GetInitializedAuthenticatorForAccount(ctx, accAddress, 2)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "authenticator id 2 failed to initialize, authenticator type MessageFilter not registered in manager: internal logic error")

	// Reset the authenticator manager to see how GetInitializedAuthenticatorForAccount behaves
	s.App.AuthenticatorManager.ResetAuthenticators()
	selectedAuthenticator, err = s.App.SmartAccountKeeper.GetInitializedAuthenticatorForAccount(ctx, accAddress, 2323)
	s.Require().Error(err)
	s.Require().Equal(selectedAuthenticator.Id, uint64(0), "Incorrect ID returned from store")
	s.Require().Equal(selectedAuthenticator.Authenticator, nil, "Returned authenticator from store but nothing registered in manager")
}
