package keeper_test

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

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

	// assert event emitted
	s.Require().Equal(s.Ctx.EventManager().Events(), sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorType, msg.Type),
			sdk.NewAttribute(types.AttributeKeyAuthenticatorId, "1"),
		),
	})
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

	msg.Type = "PassKeyAuthenticator"
	_, err := msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), msg)
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
		Id:     1,
	}

	resp, err := msgServer.RemoveAuthenticator(sdk.WrapSDKContext(ctx), removeMsg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}

func (s *KeeperTestSuite) TestMsgServer_SetActiveState() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// prep params: circuit breaker controllers
	// keeper.Keeper.SetParams(ctx, types.Params {
	// })

	ak := s.App.AuthenticatorKeeper

	// activated by default
	initialParams := ak.GetParams(ctx)
	s.Require().True(initialParams.AreSmartAccountsActive)

	// Set up account
	key := "6cf5103c60c939a5f38e383b52239c5296c968579eec1c68a47d70fbf1d19159"
	bz, _ := hex.DecodeString(key)
	priv := &secp256k1.PrivKey{Key: bz}
	accAddress := sdk.AccAddress(priv.PubKey().Address())

	// deactivate
	_, err := msgServer.SetActiveState(
		sdk.WrapSDKContext(ctx),
		&types.MsgSetActiveState{
			Sender: accAddress.String(),
			Active: false,
		})

	s.Require().NoError(err)

	// active state should be false
	params := ak.GetParams(ctx)
	s.Require().False(params.AreSmartAccountsActive)
	// other params should remain the same
	s.Require().Equal(initialParams.MaximumUnauthenticatedGas, params.MaximumUnauthenticatedGas)

	// reactivate
	_, err = msgServer.SetActiveState(
		sdk.WrapSDKContext(ctx),
		&types.MsgSetActiveState{
			Sender: accAddress.String(),
			Active: true,
		})

	s.Require().NoError(err)

	// active state should be true
	params = ak.GetParams(ctx)
	s.Require().True(params.AreSmartAccountsActive)
	// other params should remain the same
	s.Require().Equal(initialParams.MaximumUnauthenticatedGas, params.MaximumUnauthenticatedGas)
}
