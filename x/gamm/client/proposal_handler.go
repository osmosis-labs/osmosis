package client

import (
	"github.com/osmosis-labs/osmosis/v17/x/gamm/client/cli"
	"github.com/osmosis-labs/osmosis/v17/x/gamm/client/rest"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	ReplaceMigrationRecordsProposalHandler    = govclient.NewProposalHandler(cli.NewCmdSubmitReplaceMigrationRecordsProposal, rest.ProposalReplaceMigrationRecordsRESTHandler)
	UpdateMigrationRecordsProposalHandler     = govclient.NewProposalHandler(cli.NewCmdSubmitUpdateMigrationRecordsProposal, rest.ProposalUpdateMigrationRecordsRESTHandler)
	CreateCLPoolAndLinkToCFMMProposalHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitCreateCLPoolAndLinkToCFMMProposal, rest.ProposalCreateConcentratedLiquidityPoolAndLinkToCFMMHandler)
	SetScalingFactorControllerProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitSetScalingFactorControllerProposal, rest.ProposalSetScalingFactorController)
)
