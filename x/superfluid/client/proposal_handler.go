package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1/client"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/client/cli"
)

var SetSuperfluidAssetsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitSetSuperfluidAssetsProposal)

var RemoveSuperfluidAssetsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitRemoveSuperfluidAssetsProposal)
