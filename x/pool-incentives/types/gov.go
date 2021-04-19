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

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddPoolIncentives)
	govtypes.RegisterProposalTypeCodec(&AddPoolIncentivesProposal{}, "osmosis/AddPoolIncentivesProposal")
	govtypes.RegisterProposalType(ProposalTypeRemovePoolIncentives)
	govtypes.RegisterProposalTypeCodec(&RemovePoolIncentivesProposal{}, "osmosis/RemovePoolIncentivesProposal")
}

var _ govtypes.Content = &AddPoolIncentivesProposal{}

func NewAddPoolIncentivesProposal(title, description string, records []DistrRecord) govtypes.Content {
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
		recordsStr = recordsStr + fmt.Sprintf("(PotId: %d, Weight: %s) ", record.PotId, record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

var _ govtypes.Content = &EditPoolIncentivesProposal{}

func NewEditPoolIncentivesProposal(title, description string, records []EditPoolIncentivesProposal_DistrRecordWithIndex) govtypes.Content {
	return &EditPoolIncentivesProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

func (p *EditPoolIncentivesProposal) GetTitle() string { return p.Title }

func (p *EditPoolIncentivesProposal) GetDescription() string { return p.Description }

func (p *EditPoolIncentivesProposal) ProposalRoute() string { return RouterKey }

func (p *EditPoolIncentivesProposal) ProposalType() string { return ProposalTypeRemovePoolIncentives }

func (p *EditPoolIncentivesProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return ErrEmptyProposalRecords
	}

	return nil
}

func (p EditPoolIncentivesProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(Index: %d, PotId: %d, Weight: %s) ", record.Index, record.Record.PotId, record.Record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

var _ govtypes.Content = &RemovePoolIncentivesProposal{}

func NewRemovePoolIncentivesProposal(title, description string, indexes []uint64) govtypes.Content {
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
		return ErrEmptyProposalIndexes
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
