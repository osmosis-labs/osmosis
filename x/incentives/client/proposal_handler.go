package client

import (
	"github.com/osmosis-labs/osmosis/v23/x/incentives/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	HandleCreateGroupsProposal = govclient.NewProposalHandler(cli.NewCmdHandleCreateGroupsProposal)
)
