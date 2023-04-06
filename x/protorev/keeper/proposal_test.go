package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/protorev"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// TestSetProtoRevAdminAccount tests that the admin account can be set through a proposal
func (suite *KeeperTestSuite) TestSetProtoRevAdminAccount() {
	// Should be initialized to begin with
	account := suite.App.ProtoRevKeeper.GetAdminAccount(suite.Ctx)
	suite.Require().Equal(account, suite.adminAccount)

	// Set a new admin
	newAdmin := apptesting.CreateRandomAccounts(1)[0]
	err := protorev.HandleSetProtoRevAdminAccount(suite.Ctx, *suite.App.ProtoRevKeeper, &types.SetProtoRevAdminAccountProposal{
		Title:       "Updating the protorev admin account",
		Description: "This proposal is to update the protorev admin account",
		Account:     newAdmin.String(),
	})
	suite.Require().NoError(err)

	// Check that the admin account was updated
	account = suite.App.ProtoRevKeeper.GetAdminAccount(suite.Ctx)
	suite.Require().Equal(newAdmin, account)

	// Attempt to set a new admin with an invalid address
	err = protorev.HandleSetProtoRevAdminAccount(suite.Ctx, *suite.App.ProtoRevKeeper, &types.SetProtoRevAdminAccountProposal{
		Title:       "Updating the protorev admin account",
		Description: "This proposal is to update the protorev admin account",
		Account:     "invalid",
	})
	suite.Require().Error(err)
}

// TestSetProtoRevEnabledProposal tests that the enabled status can be set through a proposal
func (suite *KeeperTestSuite) TestSetProtoRevEnabledProposal() {
	// Should be enabled by default
	enabled := suite.App.ProtoRevKeeper.GetProtoRevEnabled(suite.Ctx)
	suite.Require().True(enabled)

	// Disable the protocol
	err := protorev.HandleEnabledProposal(suite.Ctx, *suite.App.ProtoRevKeeper, &types.SetProtoRevEnabledProposal{
		Title:       "Updating the protorev enabled status",
		Description: "This proposal is to update the protorev enabled status",
		Enabled:     false,
	})
	suite.Require().NoError(err)

	// Check that the enabled status was updated
	enabled = suite.App.ProtoRevKeeper.GetProtoRevEnabled(suite.Ctx)
	suite.Require().False(enabled)
}
