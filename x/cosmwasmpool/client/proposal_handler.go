package client

import (
	"github.com/osmosis-labs/osmosis/v30/x/cosmwasmpool/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	UploadCodeIdAndWhitelistProposalHandler = govclient.NewProposalHandler(cli.NewCmdUploadCodeIdAndWhitelistProposal)
	MigratePoolContractsProposalHandler     = govclient.NewProposalHandler(cli.NewCmdMigratePoolContractsProposal)
)
