package types

import (
	"fmt"
	"strings"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeCreateGroups = "CreateGroups"
)

func init() {
	govtypesv1.RegisterProposalType(ProposalTypeCreateGroups)
}

var (
	_ govtypesv1.Content = &CreateGroupsProposal{}
)

// NewCreateGroupsProposal returns a new instance of a group creation proposal struct.
func NewCreateGroupsProposal(title, description string, groups []CreateGroup) govtypesv1.Content {
	return &CreateGroupsProposal{
		Title:        title,
		Description:  description,
		CreateGroups: groups,
	}
}

func (p *CreateGroupsProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *CreateGroupsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *CreateGroupsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *CreateGroupsProposal) ProposalType() string {
	return ProposalTypeCreateGroups
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *CreateGroupsProposal) ValidateBasic() error {
	if len(p.CreateGroups) == 0 {
		return fmt.Errorf("must provide at least one group")
	}

	for _, group := range p.CreateGroups {
		if len(group.PoolIds) <= 1 {
			return fmt.Errorf("each group much be comprised of at least two pool ids")
		}
	}
	return nil
}

// String returns a string to display the proposal.
func (p CreateGroupsProposal) String() string {
	recordsStr := ""
	for _, group := range p.CreateGroups {
		recordsStr = recordsStr + fmt.Sprintf("(PoolIDs: %d) ", group.PoolIds)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create Groups Proposal:
Title:       %s
Description: %s
Groups:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
