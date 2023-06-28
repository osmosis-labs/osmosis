package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gammmigration "github.com/osmosis-labs/osmosis/v16/x/gamm/types/migration"
)

const (
	ProposalTypeUpdateMigrationRecords  = "UpdateMigrationRecords"
	ProposalTypeReplaceMigrationRecords = "ReplaceMigrationRecords"
)

// Init registers proposals to update and replace migration records.
func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateMigrationRecords)
	govtypes.RegisterProposalTypeCodec(&UpdateMigrationRecordsProposal{}, "osmosis/UpdateMigrationRecordsProposal")
	govtypes.RegisterProposalType(ProposalTypeReplaceMigrationRecords)
	govtypes.RegisterProposalTypeCodec(&ReplaceMigrationRecordsProposal{}, "osmosis/ReplaceMigrationRecordsProposal")
}

var (
	_ govtypes.Content = &UpdateMigrationRecordsProposal{}
	_ govtypes.Content = &ReplaceMigrationRecordsProposal{}
)

// NewReplacePoolIncentivesProposal returns a new instance of a replace migration record's proposal struct.
func NewReplaceMigrationRecordsProposal(title, description string, records []gammmigration.BalancerToConcentratedPoolLink) govtypes.Content {
	return &ReplaceMigrationRecordsProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

// GetTitle gets the title of the proposal
func (p *ReplaceMigrationRecordsProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *ReplaceMigrationRecordsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *ReplaceMigrationRecordsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *ReplaceMigrationRecordsProposal) ProposalType() string {
	return ProposalTypeReplaceMigrationRecords
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *ReplaceMigrationRecordsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return fmt.Errorf("empty proposal records")
	}

	return nil
}

// String returns a string containing the migration record's proposal.
func (p ReplaceMigrationRecordsProposal) String() string {
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(BalancerPoolID: %d, ClPoolID: %d) ", record.BalancerPoolId, record.ClPoolId)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Replace Migration Records Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

// NewReplacePoolIncentivesProposal returns a new instance of a replace migration record's proposal struct.
func NewUpdatePoolIncentivesProposal(title, description string, records []gammmigration.BalancerToConcentratedPoolLink) govtypes.Content {
	return &UpdateMigrationRecordsProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

// GetTitle gets the title of the proposal
func (p *UpdateMigrationRecordsProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *UpdateMigrationRecordsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *UpdateMigrationRecordsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *UpdateMigrationRecordsProposal) ProposalType() string {
	return ProposalTypeUpdateMigrationRecords
}

// ValidateBasic validates a governance proposal's abstract and basic contents.
func (p *UpdateMigrationRecordsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return fmt.Errorf("empty proposal records")
	}

	return nil
}

// String returns a string containing the migration record's proposal.
func (p UpdateMigrationRecordsProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(BalancerPoolID: %d, ClPoolID: %d) ", record.BalancerPoolId, record.ClPoolId)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Migration Records Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
