package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/protorev"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// TestSetProtoRevAdminAccount tests that the admin account can be set through a proposal
func (s *KeeperTestSuite) TestSetProtoRevAdminAccount() {
	// Should be initialized to begin with
	account := s.App.ProtoRevKeeper.GetAdminAccount(s.Ctx)
	s.Require().Equal(account, s.adminAccount)

	// Set a new admin
	newAdmin := apptesting.CreateRandomAccounts(1)[0]
	err := protorev.HandleSetProtoRevAdminAccount(s.Ctx, *s.App.ProtoRevKeeper, &types.SetProtoRevAdminAccountProposal{
		Title:       "Updating the protorev admin account",
		Description: "This proposal is to update the protorev admin account",
		Account:     newAdmin.String(),
	})
	s.Require().NoError(err)

	// Check that the admin account was updated
	account = s.App.ProtoRevKeeper.GetAdminAccount(s.Ctx)
	s.Require().Equal(newAdmin, account)

	// Attempt to set a new admin with an invalid address
	err = protorev.HandleSetProtoRevAdminAccount(s.Ctx, *s.App.ProtoRevKeeper, &types.SetProtoRevAdminAccountProposal{
		Title:       "Updating the protorev admin account",
		Description: "This proposal is to update the protorev admin account",
		Account:     "invalid",
	})
	s.Require().Error(err)
}

// TestSetProtoRevEnabledProposal tests that the enabled status can be set through a proposal
func (s *KeeperTestSuite) TestSetProtoRevEnabledProposal() {
	// Should be enabled by default
	enabled := s.App.ProtoRevKeeper.GetProtoRevEnabled(s.Ctx)
	s.Require().True(enabled)

	// Disable the protocol
	err := protorev.HandleEnabledProposal(s.Ctx, *s.App.ProtoRevKeeper, &types.SetProtoRevEnabledProposal{
		Title:       "Updating the protorev enabled status",
		Description: "This proposal is to update the protorev enabled status",
		Enabled:     false,
	})
	s.Require().NoError(err)

	// Check that the enabled status was updated
	enabled = s.App.ProtoRevKeeper.GetProtoRevEnabled(s.Ctx)
	s.Require().False(enabled)
}
