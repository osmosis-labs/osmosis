package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeAddPoolIncentives    = "AddPoolIncentives"
	ProposalTypeRemovePoolIncentives = "RemovePoolIncentives"
)

var _ govtypes.Content = &AddPoolIncentivesProposal{}

func NewAddPoolIncentivesProposal(title, description string, records []DistrRecord) *AddPoolIncentivesProposal {
	return &AddPoolIncentivesProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

func (p *AddPoolIncentivesProposal) GetTitle() string { return p.Title }

func (p *AddPoolIncentivesProposal) GetDescription() string { return p.Description }

func (p *AddPoolIncentivesProposal) ProposalRoute() string { return RouterKey }

func (p *AddPoolIncentivesProposal) ProposalType() string { return ProposalTypeAddPoolIncentives }

func (p *AddPoolIncentivesProposal) ValidateBasic() error {
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

func (p AddPoolIncentivesProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(FarmId: %d, Weight: %s) ", record.FarmId, record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

var _ govtypes.Content = &RemovePoolIncentivesProposal{}

func NewRemovePoolIncentivesProposal(title, description string, indexes []uint64) *RemovePoolIncentivesProposal {
	return &RemovePoolIncentivesProposal{
		Title:       title,
		Description: description,
		Indexes:     indexes,
	}
}

func (p *RemovePoolIncentivesProposal) GetTitle() string { return p.Title }

func (p *RemovePoolIncentivesProposal) GetDescription() string { return p.Description }

func (p *RemovePoolIncentivesProposal) ProposalRoute() string { return RouterKey }

func (p *RemovePoolIncentivesProposal) ProposalType() string { return ProposalTypeRemovePoolIncentives }

func (p *RemovePoolIncentivesProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Indexes) == 0 {
		return ErrEmptyProposalRecords
	}

	return nil
}

func (p RemovePoolIncentivesProposal) String() string {
	// TODO: Make this prettier
	indexesStr := ""
	for _, index := range p.Indexes {
		indexesStr = indexesStr + fmt.Sprintf("%d ", index)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Indexes:     %s
`, p.Title, p.Description, indexesStr))
	return b.String()
}
