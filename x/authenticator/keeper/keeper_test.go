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
	Keeper keeper.Keeper
	am     types.AuthenticatorManager
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Reset()
	// ToDo: when wired up, modify for s.App.AuthenticatorKeeper. Tests will fail for now because the store key doesn't exist
	ss, _ := s.App.ParamsKeeper.GetSubspace(types.ModuleName)
	am := types.NewAuthenticatorManager()
	s.Keeper = keeper.NewKeeper(s.App.AppCodec(), s.App.GetKey(types.StoreKey), ss, am)

	// Register the SigVerificationAuthenticator
	s.am.InitializeAuthenticators([]types.Authenticator{types.SigVerificationAuthenticator{}})
}

// ToDo: more and better tests

func (s *KeeperTestSuite) TestMsgServer_AddAuthenticator() {
	msgServer := keeper.NewMsgServerImpl(s.Keeper)
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
	msgServer := keeper.NewMsgServerImpl(s.Keeper)
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
		Id:     1, // assuming that the Id is 1 for simplicity
	}

	resp, err := msgServer.RemoveAuthenticator(sdk.WrapSDKContext(ctx), removeMsg)
	s.Require().NoError(err)
	s.Require().True(resp.Success)
}
