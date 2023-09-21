package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
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
	s.am.InitializeAuthenticators([]authenticator.Authenticator{authenticator.SignatureVerificationAuthenticator{}})
}

// ToDo: more and better tests

func (s *KeeperTestSuite) TestMsgServer_AddAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerificationAuthenticator{}.Type()))

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// Create a test message
	msg := &types.MsgAddAuthenticator{
		Sender: accAddress.String(),
		Type:   authenticator.SignatureVerificationAuthenticator{}.Type(),
		Data:   priv.PubKey().Bytes(),
	}

	resp, err := msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), msg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}

func (s *KeeperTestSuite) TestMsgServer_AddAuthenticatorFail() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerificationAuthenticator{}.Type()))

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// Create a test message
	msg := &types.MsgAddAuthenticator{
		Sender: accAddress.String(),
		Type:   authenticator.SignatureVerificationAuthenticator{}.Type(),
		Data:   priv.PubKey().Bytes(),
	}

	key = "6cf5103c60c939a5b38e383b52239c5296c968579eec1c68a47d70fbf1d19157"
	bz, _ = hex.DecodeString(key)
	priv = &secp256k1.PrivKey{Key: bz}
	accAddress = sdk.AccAddress(priv.PubKey().Address())
	msg.Data = priv.PubKey().Bytes()

	_, err := msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), msg)
	s.Require().Error(err)

	msg.Type = "PassKeyAuthenticator"
	_, err = msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), msg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestMsgServer_RemoveAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// Create a test message
	addMsg := &types.MsgAddAuthenticator{
		Sender: accAddress.String(),
		Type:   authenticator.SignatureVerificationAuthenticator{}.Type(),
		Data:   priv.PubKey().Bytes(),
	}
	_, err := msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), addMsg)
	s.Require().NoError(err)

	// Now attempt to remove it
	removeMsg := &types.MsgRemoveAuthenticator{
		Sender: accAddress.String(),
		Id:     0,
	}

	resp, err := msgServer.RemoveAuthenticator(sdk.WrapSDKContext(ctx), removeMsg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}
