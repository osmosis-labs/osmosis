package client

import (
	"github.com/c-osmosis/osmosis/x/pool-yield/client/cli"
	"github.com/c-osmosis/osmosis/x/pool-yield/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var AddPoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitAddPoolIncentivesProposal, rest.ProposalAddPoolIncentivesRESTHandler)
var RemovePoolIncentiveHandler = govclient.NewProposalHandler(cli.NewCmdSubmitRemovePoolIncentivesProposal, rest.ProposalRemovePoolIncentivesRESTHandler)
