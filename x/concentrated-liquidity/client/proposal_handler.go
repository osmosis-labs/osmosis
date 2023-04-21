package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/cli"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/rest"
)

var (
	CreateConcentratedLiquidityPoolProposalHandler = govclient.NewProposalHandler(cli.NewCmdCreateConcentratedLiquidityPoolProposal, rest.ProposalCreateConcentratedLiquidityPoolHandler)
)
