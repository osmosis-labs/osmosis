package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/client/cli"
)

// should we support legacy rest?
// general direction of the hub seems to be moving away from legacy rest
var (
	ProposalAddBurnTaxExemptionAddressHandler    = govclient.NewProposalHandler(cli.ProposalAddBurnTaxExemptionAddressCmd)
	ProposalRemoveBurnTaxExemptionAddressHandler = govclient.NewProposalHandler(cli.ProposalRemoveBurnTaxExemptionAddressCmd)
)
