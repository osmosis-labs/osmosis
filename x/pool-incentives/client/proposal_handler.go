package client

import (
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	UpdatePoolIncentivesHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitUpdatePoolIncentivesProposal)
	ReplacePoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitReplacePoolIncentivesProposal)
)
