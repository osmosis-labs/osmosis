package client

import (
	"github.com/osmosis-labs/osmosis/v28/x/poolmanager/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	DenomPairTakerFeeProposalHandler = govclient.NewProposalHandler(cli.NewCmdHandleDenomPairTakerFeeProposal)
)
