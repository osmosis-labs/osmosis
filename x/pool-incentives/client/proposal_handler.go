package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client/v1beta1"
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/client/cli"
)

var UpdatePoolIncentivesHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdatePoolIncentivesProposal)
