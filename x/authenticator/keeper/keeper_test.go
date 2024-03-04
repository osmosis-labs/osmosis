package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/testutils"
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
		authenticator.SignatureVerificationAuthenticator{},
		testutils.TestingAuthenticator{
			Approve:        testutils.Always,
			GasConsumption: 10,
			Confirm:        testutils.Always,
		},
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAuthenticatorDataForAccount() {
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerificationAuthenticator{}.Type()))

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	id, err := s.App.AuthenticatorKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerificationAuthenticator",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(id, uint64(1), "Adding authenticator returning incorrect id")

	id, err = s.App.AuthenticatorKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerificationAuthenticator",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(id, uint64(2), "Adding authenticator returning incorrect id")

	authenticators, err := s.App.AuthenticatorKeeper.GetAuthenticatorDataForAccount(ctx, accAddress)
	s.Require().NoError(err)
	s.Require().Equal(len(authenticators), 2, "Getting authenticators returning incorrect data")
}

func (s *KeeperTestSuite) TestKeeper_GetAuthenticatorsForAccount() {
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerificationAuthenticator{}.Type()))

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	id, err := s.App.AuthenticatorKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerificationAuthenticator",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(id, uint64(1), "Adding authenticator returning incorrect id")

	id, err = s.App.AuthenticatorKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerificationAuthenticator",
		priv.PubKey().Bytes(),
	)
	s.Require().NoError(err)
	s.Require().Equal(id, uint64(2), "Adding authenticator returning incorrect id")

	authenticators, err := s.App.AuthenticatorKeeper.GetAuthenticatorsForAccount(ctx, accAddress)
	s.Require().NoError(err)
	s.Require().Equal(len(authenticators), 2, "Getting authenticators returning incorrect data")

	_, err = s.App.AuthenticatorKeeper.AddAuthenticator(
		ctx,
		accAddress,
		"SignatureVerificationAuthenticator",
		[]byte("BrokenBytes"),
	)
	s.Require().Error(err)

	s.App.AuthenticatorManager.ResetAuthenticators()
	authenticators, err = s.App.AuthenticatorKeeper.GetAuthenticatorsForAccount(ctx, accAddress)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "failed to initialize")
}
