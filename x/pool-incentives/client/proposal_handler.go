package client

import (
	"github.com/c-osmosis/osmosis/x/pool-incentives/client/cli"
	"github.com/c-osmosis/osmosis/x/pool-incentives/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var UpdatePoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdatePoolIncentivesProposal, rest.ProposalUpdatePoolIncentivesRESTHandler)
