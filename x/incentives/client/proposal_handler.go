package client

import (
	"github.com/osmosis-labs/osmosis/v19/x/incentives/client/cli"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	HandleCreateGaugeGroupsProposal = govclient.NewProposalHandler(cli.NewCmdHandleCreateGaugeGroupsProposal, rest.ProposalCreateGaugeGroupsRESTHandler)
)
