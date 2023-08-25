package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/osmosis-labs/osmosis/v17/x/txfees/client/cli"
	"github.com/osmosis-labs/osmosis/v17/x/txfees/client/rest"
)

var (
	SubmitUpdateFeeTokenProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdateFeeTokenProposal, rest.ProposalUpdateFeeTokenProposalRESTHandler)
)
