package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeUpdatePoolIncentives  = "UpdatePoolIncentives"
	ProposalTypeReplacePoolIncentives = "ReplacePoolIncentives"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdatePoolIncentives)
	govtypes.RegisterProposalTypeCodec(&UpdatePoolIncentivesProposal{}, "osmosis/UpdatePoolIncentivesProposal")
	govtypes.RegisterProposalType(ProposalTypeReplacePoolIncentives)
	govtypes.RegisterProposalTypeCodec(&ReplacePoolIncentivesProposal{}, "osmosis/ReplacePoolIncentivesProposal")
}

var (
	_ govtypes.Content = &UpdatePoolIncentivesProposal{}
	_ govtypes.Content = &ReplacePoolIncentivesProposal{}
)

func NewReplacePoolIncentivesProposal(title, description string, records []DistrRecord) govtypes.Content {
	return &ReplacePoolIncentivesProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

func (p *ReplacePoolIncentivesProposal) GetTitle() string { return p.Title }

func (p *ReplacePoolIncentivesProposal) GetDescription() string { return p.Description }

func (p *ReplacePoolIncentivesProposal) ProposalRoute() string { return RouterKey }

func (p *ReplacePoolIncentivesProposal) ProposalType() string {
	return ProposalTypeReplacePoolIncentives
}

func (p *ReplacePoolIncentivesProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
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

func NewUpdatePoolIncentivesProposal(title, description string, records []DistrRecord) govtypes.Content {
	return &UpdatePoolIncentivesProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

func (p *UpdatePoolIncentivesProposal) GetTitle() string { return p.Title }

func (p *UpdatePoolIncentivesProposal) GetDescription() string { return p.Description }

func (p *UpdatePoolIncentivesProposal) ProposalRoute() string { return RouterKey }

func (p *UpdatePoolIncentivesProposal) ProposalType() string { return ProposalTypeUpdatePoolIncentives }

func (p *UpdatePoolIncentivesProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
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
