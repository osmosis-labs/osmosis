package client

import (
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// TODO: Remove in v27 once comfortable with new gov message
var (
	SetSuperfluidAssetsProposalHandler    = govclient.NewProposalHandler(cli.NewCmdSubmitSetSuperfluidAssetsProposal)
	RemoveSuperfluidAssetsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitRemoveSuperfluidAssetsProposal)
	UpdateUnpoolWhitelistProposalHandler  = govclient.NewProposalHandler(cli.NewCmdUpdateUnpoolWhitelistProposal)
)
