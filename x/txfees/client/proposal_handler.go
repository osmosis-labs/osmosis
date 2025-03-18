package client

import (
	"github.com/osmosis-labs/osmosis/v29/x/txfees/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	SubmitUpdateFeeTokenProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitUpdateFeeTokenProposal)
)
