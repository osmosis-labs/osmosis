package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/osmosis-labs/osmosis/v20/x/authenticator/iface"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v20/app/apptesting"
	"github.com/osmosis-labs/osmosis/v20/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v20/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v20/x/authenticator/types"
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
	s.am.InitializeAuthenticators([]iface.Authenticator{authenticator.SignatureVerificationAuthenticator{}})
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

func (s *KeeperTestSuite) TestMsgServer_CreateAccount() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerificationAuthenticator{}.Type()))

	testCases := []struct {
		name            string
		msg             *types.MsgCreateAccount
		expectError     bool
		expectedAddress string
	}{
		{
			name: "valid input",
			msg: &types.MsgCreateAccount{
				Sender: "osmo1l4u56l7cvx8n0n6c7w650k02vz67qudjlcut89",
				Salt:   "testSalt",
				Authenticators: []*types.AuthenticatorData{
					{Type: "SignatureVerificationAuthenticator", Data: priv.PubKey().Bytes()},
				},
			},
			expectError: false,
		},

		{
			name: "invalid sender",
			msg: &types.MsgCreateAccount{
				Sender: "invalidSender",
				Salt:   "testSalt",
				Authenticators: []*types.AuthenticatorData{
					{Type: "SignatureVerificationAuthenticator", Data: priv.PubKey().Bytes()},
				},
			},
			expectError: true,
		},

		{
			name: "empty salt",
			msg: &types.MsgCreateAccount{
				Sender: "osmo1l4u56l7cvx8n0n6c7w650k02vz67qudjlcut89",
				Salt:   "",
				Authenticators: []*types.AuthenticatorData{
					{Type: "SignatureVerificationAuthenticator", Data: priv.PubKey().Bytes()},
				},
			},
			expectError: false,
		},

		{
			name: "valid input again (already exists)",
			msg: &types.MsgCreateAccount{
				Sender: "osmo1l4u56l7cvx8n0n6c7w650k02vz67qudjlcut89",
				Salt:   "testSalt",
				Authenticators: []*types.AuthenticatorData{
					{Type: "SignatureVerificationAuthenticator", Data: priv.PubKey().Bytes()},
				},
			},
			expectError: true,
		},

		{
			name: "valid input again (but with new salt)",
			msg: &types.MsgCreateAccount{
				Sender: "osmo1l4u56l7cvx8n0n6c7w650k02vz67qudjlcut89",
				Salt:   "testSalt2",
				Authenticators: []*types.AuthenticatorData{
					{Type: "SignatureVerificationAuthenticator", Data: priv.PubKey().Bytes()},
				},
			},
			expectError: false,
		},

		{
			name: "invalid authenticator data",
			msg: &types.MsgCreateAccount{
				Sender: "osmo1l4u56l7cvx8n0n6c7w650k02vz67qudjlcut89",
				Salt:   "testSalt2",
				Authenticators: []*types.AuthenticatorData{
					{Type: "SignatureVerificationAuthenticator", Data: []byte("testData")},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := msgServer.CreateAccount(sdk.WrapSDKContext(ctx), tc.msg)

			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp.Address)

				_, err = sdk.AccAddressFromBech32(resp.Address)
				s.Require().NoError(err)

				var data strings.Builder
				data.WriteString(tc.msg.Salt)
				for _, authenticatorData := range tc.msg.Authenticators {
					data.WriteString(authenticatorData.Type)
					data.Write(authenticatorData.Data)
				}

				hashResult := sha256.Sum256([]byte(data.String()))
				expectedAddress := sdk.AccAddress(hashResult[:]).String()
				s.Require().Equal(expectedAddress, resp.Address)

				// Check that the account has the right number of authenticators
				account, err := sdk.AccAddressFromBech32(resp.Address)
				s.Require().NoError(err)
				authenticators, err := s.App.AuthenticatorKeeper.GetAuthenticatorsForAccount(ctx, account)
				s.Require().NoError(err)
				s.Require().Equal(len(tc.msg.Authenticators), len(authenticators))
				for i, ator := range authenticators {
					s.Require().Equal(tc.msg.Authenticators[i].Type, ator.Type())
				}
			}
		})
	}
}
