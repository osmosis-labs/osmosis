package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammmigration "github.com/osmosis-labs/osmosis/v27/x/gamm/types/migration"
)

const (
	ProposalTypeUpdateMigrationRecords                       = "UpdateMigrationRecords"
	ProposalTypeReplaceMigrationRecords                      = "ReplaceMigrationRecords"
	ProposalTypeCreateConcentratedLiquidityPoolAndLinktoCFMM = "CreateConcentratedLiquidityPoolAndLinktoCFMM"
	ProposalTypeSetScalingFactorController                   = "SetScalingFactorController"
)

// Init registers proposals to update and replace migration records.
func init() {
	govtypesv1.RegisterProposalType(ProposalTypeUpdateMigrationRecords)
	govtypesv1.RegisterProposalType(ProposalTypeReplaceMigrationRecords)
	govtypesv1.RegisterProposalType(ProposalTypeCreateConcentratedLiquidityPoolAndLinktoCFMM)
	govtypesv1.RegisterProposalType(ProposalTypeSetScalingFactorController)
}

var (
	_ govtypesv1.Content = &UpdateMigrationRecordsProposal{}
	_ govtypesv1.Content = &ReplaceMigrationRecordsProposal{}
	_ govtypesv1.Content = &CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal{}
	_ govtypesv1.Content = &SetScalingFactorControllerProposal{}
)

// NewReplacePoolIncentivesProposal returns a new instance of a replace migration record's proposal struct.
func NewReplaceMigrationRecordsProposal(title, description string, records []gammmigration.BalancerToConcentratedPoolLink) govtypesv1.Content {
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
	err := govtypesv1.ValidateAbstract(p)
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
func NewUpdatePoolIncentivesProposal(title, description string, records []gammmigration.BalancerToConcentratedPoolLink) govtypesv1.Content {
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
	err := govtypesv1.ValidateAbstract(p)
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

func NewCreateConcentratedLiquidityPoolsAndLinktoCFMMProposal(title, description string, records []PoolRecordWithCFMMLink) govtypesv1.Content {
	return &CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal{
		Title:                   title,
		Description:             description,
		PoolRecordsWithCfmmLink: records,
	}
}

// GetTitle gets the title of the proposal
func (p *CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) GetDescription() string {
	return p.Description
}

// ProposalRoute returns the router key for the proposal
func (p *CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns the type of the proposal
func (p *CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) ProposalType() string {
	return ProposalTypeCreateConcentratedLiquidityPoolAndLinktoCFMM
}

// ValidateBasic validates a governance proposal's abstract and basic contents.
func (p *CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, record := range p.PoolRecordsWithCfmmLink {
		if record.TickSpacing <= 0 {
			return fmt.Errorf("tick spacing must be positive")
		}

		if record.Denom0 == record.Denom1 {
			return fmt.Errorf("denom0 and denom1 must be different")
		}

		if sdk.ValidateDenom(record.Denom0) != nil {
			return fmt.Errorf("denom0 is invalid: %s", sdk.ValidateDenom(record.Denom0))
		}

		if sdk.ValidateDenom(record.Denom1) != nil {
			return fmt.Errorf("denom1 is invalid: %s", sdk.ValidateDenom(record.Denom1))
		}

		spreadFactor := record.SpreadFactor
		if spreadFactor.IsNegative() || spreadFactor.GTE(osmomath.OneDec()) {
			return fmt.Errorf("Invalid Spread factor")
		}

		if record.BalancerPoolId <= 0 {
			return fmt.Errorf("Invalid Balancer Pool Id")
		}
	}
	return nil
}

// String returns a string containing creating CL pool and linking it to an existing CFMM pool.
func (p CreateConcentratedLiquidityPoolsAndLinktoCFMMProposal) String() string {
	recordsStr := ""
	for _, record := range p.PoolRecordsWithCfmmLink {
		recordsStr = recordsStr + fmt.Sprintf("(Denom0: %s, Denom1: %s, TickSpacing: %d, ExponentAtPriceOne: %d, SpreadFactor: %d, BalancerPoolId: %d) ", record.Denom0, record.Denom1, record.TickSpacing, record.ExponentAtPriceOne, record.SpreadFactor, record.BalancerPoolId)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create Concentrated Liquidity Pool Proposal:
Title:       %s
Description: %s
Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

// NewSetScalingFactorControllerProposal returns a new instance of a replace migration record's proposal struct.
func NewSetScalingFactorControllerProposal(title, description string, poolId uint64, controllerAddress string) govtypesv1.Content {
	return &SetScalingFactorControllerProposal{
		Title:             title,
		Description:       description,
		PoolId:            poolId,
		ControllerAddress: controllerAddress,
	}
}

// GetTitle gets the title of the proposal
func (p *SetScalingFactorControllerProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *SetScalingFactorControllerProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *SetScalingFactorControllerProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *SetScalingFactorControllerProposal) ProposalType() string {
	return ProposalTypeSetScalingFactorController
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *SetScalingFactorControllerProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}
	_, err = sdk.AccAddressFromBech32(p.ControllerAddress)
	if err != nil {
		return fmt.Errorf("Invalid controller address (%s)", err)
	}

	return nil
}

// String returns a string containing the migration record's proposal.
func (p SetScalingFactorControllerProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Set Scaling Factor Controller Address Proposal:
  Title:             %s
  Description:       %s
  PoolId:            %d
  ControllerAddress: %s
`, p.Title, p.Description, p.PoolId, p.ControllerAddress))
	return b.String()
}
