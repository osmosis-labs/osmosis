package client

import (
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	TickSpacingDecreaseProposalHandler             = govclient.NewProposalHandler(cli.NewTickSpacingDecreaseProposal)
	CreateConcentratedLiquidityPoolProposalHandler = govclient.NewProposalHandler(cli.NewCmdCreateConcentratedLiquidityPoolsProposal)
)
