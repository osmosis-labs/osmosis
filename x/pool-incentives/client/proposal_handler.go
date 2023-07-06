package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/osmosis-labs/osmosis/v16/x/pool-incentives/client/cli"
	"github.com/osmosis-labs/osmosis/v16/x/pool-incentives/client/rest"
)

var (
	UpdatePoolIncentivesHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitUpdatePoolIncentivesProposal, rest.ProposalUpdatePoolIncentivesRESTHandler)
	ReplacePoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitReplacePoolIncentivesProposal, rest.ProposalReplacePoolIncentivesRESTHandler)
)
