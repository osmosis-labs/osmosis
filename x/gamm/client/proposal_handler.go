package client

import (

	"github.com/osmosis-labs/osmosis/v7/x/gamm/client/cli"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	SetSwapFeeProposalHandler = govclient.NewProposalHandler(cli.NewSetSwapFeeProposalCmd, rest.ProposalSetSwapFeeRESTHandler)
	SetExitFeeProposalHandler = govclient.NewProposalHandler(cli.NewSetExitFeeProposalCmd, rest.ProposalSetExitFeeRESTHandler)
)
