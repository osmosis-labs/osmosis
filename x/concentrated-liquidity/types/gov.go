package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	ProposalTypeCreateConcentratedLiquidityPool = "CreateConcentratedLiquidityPool"
	ProposalTypeTickSpacingDecrease             = "TickSpacingDecrease"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateConcentratedLiquidityPool)
	govtypes.RegisterProposalTypeCodec(&CreateConcentratedLiquidityPoolsProposal{}, "osmosis/CreateCLPoolsProposal")
	govtypes.RegisterProposalType(ProposalTypeTickSpacingDecrease)
	govtypes.RegisterProposalTypeCodec(&TickSpacingDecreaseProposal{}, "osmosis/TickSpacingDecreaseProposal")
}

var (
	_ govtypes.Content = &CreateConcentratedLiquidityPoolsProposal{}
	_ govtypes.Content = &TickSpacingDecreaseProposal{}
)

// NewCreateConcentratedLiquidityPoolsProposal returns a new instance of a create concentrated liquidity pool proposal struct.
func NewCreateConcentratedLiquidityPoolsProposal(title, description string, records []PoolRecord) govtypes.Content {
	return &CreateConcentratedLiquidityPoolsProposal{
		Title:       title,
		Description: description,
		PoolRecords: records,
	}
}

func (p *CreateConcentratedLiquidityPoolsProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *CreateConcentratedLiquidityPoolsProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *CreateConcentratedLiquidityPoolsProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *CreateConcentratedLiquidityPoolsProposal) ProposalType() string {
	return ProposalTypeCreateConcentratedLiquidityPool
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *CreateConcentratedLiquidityPoolsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, record := range p.PoolRecords {
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
			return InvalidSpreadFactorError{ActualSpreadFactor: spreadFactor}
		}
	}
	return nil
}

// String returns a string containing the pool incentives proposal.
func (p CreateConcentratedLiquidityPoolsProposal) String() string {
	recordsStr := ""
	for _, record := range p.PoolRecords {
		recordsStr = recordsStr + fmt.Sprintf("(Denom0: %s, Denom1: %s, TickSpacing: %d, SpreadFactor: %d) ", record.Denom0, record.Denom1, record.TickSpacing, record.SpreadFactor)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create Concentrated Liquidity Pool Proposal:
Title:       %s
Description: %s
Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

func NewTickSpacingDecreaseProposal(title, description string, records []PoolIdToTickSpacingRecord) govtypes.Content {
	return &TickSpacingDecreaseProposal{
		Title:                      title,
		Description:                description,
		PoolIdToTickSpacingRecords: records,
	}
}

// GetTitle gets the title of the proposal
func (p *TickSpacingDecreaseProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *TickSpacingDecreaseProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *TickSpacingDecreaseProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *TickSpacingDecreaseProposal) ProposalType() string {
	return ProposalTypeTickSpacingDecrease
}

// ValidateBasic validates a governance proposal's abstract and basic contents.
func (p *TickSpacingDecreaseProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.PoolIdToTickSpacingRecords) == 0 {
		return fmt.Errorf("empty proposal records")
	}

	for _, poolIdToTickSpacingRecord := range p.PoolIdToTickSpacingRecords {
		if poolIdToTickSpacingRecord.PoolId <= uint64(0) {
			return fmt.Errorf("Pool Id cannot be negative")
		}

		if poolIdToTickSpacingRecord.NewTickSpacing <= uint64(0) {
			return fmt.Errorf("tick spacing must be positive")
		}
	}
	return nil
}

// String returns a string containing the decrease tick spacing proposal.
func (p TickSpacingDecreaseProposal) String() string {
	recordsStr := ""
	for _, record := range p.PoolIdToTickSpacingRecords {
		recordsStr = recordsStr + fmt.Sprintf("(PoolID: %d, NewTickSpacing: %d) ", record.PoolId, record.NewTickSpacing)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Decrease Pools Tick Spacing Proposal:
Title:       %s
Description: %s
Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
