package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeUpdatePoolIncentives = "UpdatePoolIncentives"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdatePoolIncentives)
	govtypes.RegisterProposalTypeCodec(&UpdatePoolIncentivesProposal{}, "osmosis/UpdatePoolIncentivesProposal")
}

var _ govtypes.Content = &UpdatePoolIncentivesProposal{}

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
		if err := record.Validate(); err != nil {
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
