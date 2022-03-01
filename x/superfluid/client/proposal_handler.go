package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client/v1beta1"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/client/cli"
	//	"github.com/osmosis-labs/osmosis/v7/x/superfluid/client/rest"
)

var SetSuperfluidAssetsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitSetSuperfluidAssetsProposal)

var RemoveSuperfluidAssetsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitRemoveSuperfluidAssetsProposal)
