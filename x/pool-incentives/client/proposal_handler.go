package client

import (
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/client/cli"
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var UpdatePoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdatePoolIncentivesProposal, rest.ProposalUpdatePoolIncentivesRESTHandler)
