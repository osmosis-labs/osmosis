package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

type GovTestSuite struct {
	suite.Suite
}

func TestGovTestSuite(t *testing.T) {
	suite.Run(t, new(GovTestSuite))
}

func (suite *GovTestSuite) TestGovKeysTypes() {
	suite.Require().Equal("SetProtoRevEnabledProposal", (&types.SetProtoRevEnabledProposal{}).ProposalType())
	suite.Require().Equal("SetProtoRevAdminAccountProposal", (&types.SetProtoRevAdminAccountProposal{}).ProposalType())
}

func (suite *GovTestSuite) TestEnableProposal() {
	testCases := []struct {
		description string
		enabled     bool
	}{
		{
			description: "enabled",
			enabled:     true,
		},
		{
			description: "disabled",
			enabled:     false,
		},
	}

	for _, tc := range testCases {
		proposal := types.NewSetProtoRevEnabledProposal("title", "description", tc.enabled)
		setProtoRevEnabledProposal, ok := proposal.(*types.SetProtoRevEnabledProposal)
		suite.Require().True(ok, "proposal is not a SetProtoRevEnabledProposal")
		suite.Require().Equal(tc.enabled, setProtoRevEnabledProposal.Enabled)
	}
}

func (suite *GovTestSuite) TestAdminAccountProposal() {
	testCases := []struct {
		description string
		address     string
		pass        bool
	}{
		{
			description: "valid address",
			address:     apptesting.CreateRandomAccounts(1)[0].String(),
			pass:        true,
		},
		{
			description: "invalid address",
			address:     "invalid",
			pass:        false,
		},
	}

	for _, tc := range testCases {
		proposal := types.NewSetProtoRevAdminAccountProposal("title", "description", tc.address)
		if tc.pass {
			suite.Require().NoError(proposal.ValidateBasic())
		} else {
			suite.Require().Error(proposal.ValidateBasic())
		}
	}
}
