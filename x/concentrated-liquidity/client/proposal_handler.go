package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/client/cli"
	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/client/rest"
)

var (
	TickSpacingDecreaseProposalHandler             = govclient.NewProposalHandler(cli.NewTickSpacingDecreaseProposal, rest.ProposalTickSpacingDecreaseRESTHandler)
	CreateConcentratedLiquidityPoolProposalHandler = govclient.NewProposalHandler(cli.NewCmdCreateConcentratedLiquidityPoolsProposal, rest.ProposalCreateConcentratedLiquidityPoolHandler)
)
