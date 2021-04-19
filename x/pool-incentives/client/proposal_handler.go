package client

import (
	"github.com/c-osmosis/osmosis/x/pool-incentives/client/cli"
	"github.com/c-osmosis/osmosis/x/pool-incentives/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var AddPoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitAddPoolIncentivesProposal, rest.ProposalAddPoolIncentivesRESTHandler)
var EditPoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitEditPoolIncentivesProposal, rest.ProposalEditPoolIncentivesRESTHandler)
var RemovePoolIncentiveHandler = govclient.NewProposalHandler(cli.NewCmdSubmitRemovePoolIncentivesProposal, rest.ProposalRemovePoolIncentivesRESTHandler)
