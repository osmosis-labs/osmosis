package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/keeper"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	am *types.AuthenticatorManager
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Reset()
	s.am = types.NewAuthenticatorManager()
	// Register the SigVerificationAuthenticator
	s.am.InitializeAuthenticators([]types.Authenticator{types.SigVerificationAuthenticator{}})
}

// ToDo: more and better tests

func (s *KeeperTestSuite) TestMsgServer_AddAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// Ensure the SigVerificationAuthenticator type is registered
	s.Require().True(s.am.IsAuthenticatorTypeRegistered(types.SigVerificationAuthenticator{}.Type()))

	// Create a test message
	msg := &types.MsgAddAuthenticator{
		Sender: s.TestAccs[0].String(),
		Type:   types.SigVerificationAuthenticator{}.Type(),
	}

	resp, err := msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), msg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}

func (s *KeeperTestSuite) TestMsgServer_RemoveAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(*s.App.AuthenticatorKeeper)
	ctx := s.Ctx

	// First add an authenticator so that we can attempt to remove it later
	addMsg := &types.MsgAddAuthenticator{
		Sender: s.TestAccs[0].String(),
		Type:   types.SigVerificationAuthenticator{}.Type(),
	}
	_, err := msgServer.AddAuthenticator(sdk.WrapSDKContext(ctx), addMsg)
	s.Require().NoError(err)

	// Now attempt to remove it
	removeMsg := &types.MsgRemoveAuthenticator{
		Sender: s.TestAccs[0].String(),
		Id:     0,
	}

	resp, err := msgServer.RemoveAuthenticator(sdk.WrapSDKContext(ctx), removeMsg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}
