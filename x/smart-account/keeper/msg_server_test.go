package keeper_test

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/authenticator"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

func (s *KeeperTestSuite) TestMsgServer_AddAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(*s.App.SmartAccountKeeper)
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerification{}.Type()))

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// Create a test message
	msg := &types.MsgAddAuthenticator{
		Sender:            accAddress.String(),
		AuthenticatorType: authenticator.SignatureVerification{}.Type(),
		Data:              priv.PubKey().Bytes(),
	}

	resp, err := msgServer.AddAuthenticator(ctx, msg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)

	// assert event emitted
	s.Require().Equal(s.Ctx.EventManager().Events(), sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorType, msg.AuthenticatorType),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorId, "1"),
		),
	})
}

func (s *KeeperTestSuite) TestMsgServer_AddAuthenticatorFail() {
	msgServer := keeper.NewMsgServerImpl(*s.App.SmartAccountKeeper)
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(authenticator.SignatureVerification{}.Type()))

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// Create a test message
	msg := &types.MsgAddAuthenticator{
		Sender:            accAddress.String(),
		AuthenticatorType: authenticator.SignatureVerification{}.Type(),
		Data:              priv.PubKey().Bytes(),
	}

	msg.AuthenticatorType = "PassKeyAuthenticator"
	_, err := msgServer.AddAuthenticator(ctx, msg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestMsgServer_RemoveAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(*s.App.SmartAccountKeeper)
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// Create a test message
	addMsg := &types.MsgAddAuthenticator{
		Sender:            accAddress.String(),
		AuthenticatorType: authenticator.SignatureVerification{}.Type(),
		Data:              priv.PubKey().Bytes(),
	}
	_, err := msgServer.AddAuthenticator(ctx, addMsg)
	s.Require().NoError(err)

	// Now attempt to remove it
	removeMsg := &types.MsgRemoveAuthenticator{
		Sender: accAddress.String(),
		Id:     1,
	}

	resp, err := msgServer.RemoveAuthenticator(ctx, removeMsg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}

func (s *KeeperTestSuite) TestMsgServer_SetActiveState() {
	ak := *s.App.SmartAccountKeeper
	msgServer := keeper.NewMsgServerImpl(ak)
	ctx := s.Ctx

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	authorizedAccAddress := sdk.AccAddress(priv.PubKey().Address())

	key = "0dd4d1506e18a5712080708c338eb51ecf2afdceae01e8162e890b126ac190fe"
	bz, _ = hex.DecodeString(key)
	priv = &secp256k1.PrivKey{Key: bz}
	unauthorizedAccAddress := sdk.AccAddress(priv.PubKey().Address())

	// activated by default
	initialParams := s.App.SmartAccountKeeper.GetParams(ctx)
	s.Require().True(initialParams.IsSmartAccountActive)

	// Set the authorized account as the circuit breaker controller
	initialParams = s.App.SmartAccountKeeper.GetParams(ctx)
	initialParams.CircuitBreakerControllers = []string{authorizedAccAddress.String()}
	s.App.SmartAccountKeeper.SetParams(ctx, initialParams)

	// deactivate by unauthorized account
	_, err := msgServer.SetActiveState(
		ctx,
		&types.MsgSetActiveState{
			Sender: unauthorizedAccAddress.String(),
			Active: false,
		})

	s.Require().Error(err)
	s.Require().Equal(err.Error(), "signer is not a circuit breaker controller: unauthorized")

	// deactivate
	_, err = msgServer.SetActiveState(
		ctx,

		&types.MsgSetActiveState{
			Sender: authorizedAccAddress.String(),
			Active: false,
		})

	s.Require().NoError(err)

	// active state should be false
	params := ak.GetParams(ctx)
	s.Require().False(params.IsSmartAccountActive)
	// other params should remain the same
	params.IsSmartAccountActive = initialParams.IsSmartAccountActive
	s.Require().Equal(params, initialParams)

	// reactivate by a controller (unauthorized)
	_, err = msgServer.SetActiveState(
		ctx,
		&types.MsgSetActiveState{
			Sender: authorizedAccAddress.String(),
			Active: true,
		})
	s.Require().Error(err)
	s.Require().Equal(err.Error(), "signer is not the circuit breaker governor: unauthorized")

	// reactivate by gov
	governor := s.App.SmartAccountKeeper.CircuitBreakerGovernor
	_, err = msgServer.SetActiveState(
		ctx,
		&types.MsgSetActiveState{
			Sender: governor.String(),
			Active: true,
		})
	s.Require().NoError(err)

	// active state should be true
	params = ak.GetParams(ctx)
	s.Require().True(params.IsSmartAccountActive)
	// other params should remain the same
	params.IsSmartAccountActive = initialParams.IsSmartAccountActive
	s.Require().Equal(params, initialParams)
}

func (s *KeeperTestSuite) TestMsgServer_SmartAccountsNotActive() {
	msgServer := keeper.NewMsgServerImpl(*s.App.SmartAccountKeeper)
	ctx := s.Ctx

	s.App.SmartAccountKeeper.SetParams(s.Ctx, types.Params{IsSmartAccountActive: false})

	// Create a test message
	msg := &types.MsgAddAuthenticator{
		Sender:            "",
		AuthenticatorType: authenticator.SignatureVerification{}.Type(),
		Data:              []byte(""),
	}

	_, err := msgServer.AddAuthenticator(ctx, msg)
	s.Require().Error(err)
	s.Require().Equal(err.Error(), "smartaccount module is not active: unauthorized")

	removeMsg := &types.MsgRemoveAuthenticator{
		Sender: "",
		Id:     1,
	}

	_, err = msgServer.RemoveAuthenticator(ctx, removeMsg)
	s.Require().Error(err)
	s.Require().Equal(err.Error(), "smartaccount module is not active: unauthorized")
}
