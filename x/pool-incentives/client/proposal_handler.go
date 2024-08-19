package client

import (
	"github.com/osmosis-labs/osmosis/v25/x/pool-incentives/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// TODO: Remove in v27 once comfortable with new gov message
var (
	UpdatePoolIncentivesHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitUpdatePoolIncentivesProposal)
	ReplacePoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitReplacePoolIncentivesProposal)
)
