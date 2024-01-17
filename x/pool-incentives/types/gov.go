package types

import (
	"fmt"
	"strings"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdatePoolIncentives  = "UpdatePoolIncentives"
	ProposalTypeReplacePoolIncentives = "ReplacePoolIncentives"
)

// Init registers proposals to update and replace pool incentives.
func init() {
	govtypesv1.RegisterProposalType(ProposalTypeUpdatePoolIncentives)
	govtypesv1.RegisterProposalType(ProposalTypeReplacePoolIncentives)
}

var (
	_ govtypesv1.Content = &UpdatePoolIncentivesProposal{}
	_ govtypesv1.Content = &ReplacePoolIncentivesProposal{}
)

// NewReplacePoolIncentivesProposal returns a new instance of a replace pool incentives proposal struct.
func NewReplacePoolIncentivesProposal(title, description string, records []DistrRecord) govtypesv1.Content {
	return &ReplacePoolIncentivesProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

// GetTitle gets the title of the proposal
func (p *ReplacePoolIncentivesProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *ReplacePoolIncentivesProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *ReplacePoolIncentivesProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *ReplacePoolIncentivesProposal) ProposalType() string {
	return ProposalTypeReplacePoolIncentives
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *ReplacePoolIncentivesProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return ErrEmptyProposalRecords
	}

	for _, record := range p.Records {
		if err := record.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p ReplacePoolIncentivesProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(GaugeId: %d, Weight: %s) ", record.GaugeId, record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Replace Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

// NewReplacePoolIncentivesProposal returns a new instance of a replace pool incentives proposal struct.
func NewUpdatePoolIncentivesProposal(title, description string, records []DistrRecord) govtypesv1.Content {
	return &UpdatePoolIncentivesProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

// GetTitle gets the title of the proposal
func (p *UpdatePoolIncentivesProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *UpdatePoolIncentivesProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *UpdatePoolIncentivesProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *UpdatePoolIncentivesProposal) ProposalType() string { return ProposalTypeUpdatePoolIncentives }

// ValidateBasic validates a governance proposal's abstract and basic contents.
func (p *UpdatePoolIncentivesProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return ErrEmptyProposalRecords
	}

	for _, record := range p.Records {
		if err := record.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p UpdatePoolIncentivesProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(GaugeId: %d, Weight: %s) ", record.GaugeId, record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
