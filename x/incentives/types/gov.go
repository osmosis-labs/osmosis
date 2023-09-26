package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeCreateGaugeGroups = "CreateGaugeGroupsProposal"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateGaugeGroups)
	govtypes.RegisterProposalTypeCodec(&CreateGaugeGroupsProposal{}, "osmosis/CreateGaugeGroupsProposal")
}

var (
	_ govtypes.Content = &CreateGaugeGroupsProposal{}
)

// NewCreateGaugeGroupsProposal returns a new instance of a gauge group creation proposal struct.
func NewCreateGaugeGroupsProposal(title, description string, groups []CreateGroup) govtypes.Content {
	return &CreateGaugeGroupsProposal{
		Title:        title,
		Description:  description,
		CreateGroups: groups,
	}
}

func (p *CreateGaugeGroupsProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *CreateGaugeGroupsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *CreateGaugeGroupsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *CreateGaugeGroupsProposal) ProposalType() string {
	return ProposalTypeCreateGaugeGroups
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *CreateGaugeGroupsProposal) ValidateBasic() error {
	// TODO: Complete
	return nil
}

// String returns a string containing the denom pair taker fee proposal.
func (p CreateGaugeGroupsProposal) String() string {
	recordsStr := ""
	for _, group := range p.CreateGroups {
		recordsStr = recordsStr + fmt.Sprintf("(Coins: %s, NumEpochsPaidOver: %d, FilledEpochs: %d, PoolIDs: %d) ", group.Coins, group.NumEpochsPaidOver, group.FilledEpochs, group.PoolIds)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create Gauge Groups Proposal:
Title:       %s
Description: %s
Groups:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
