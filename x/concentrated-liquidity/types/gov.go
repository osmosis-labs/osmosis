package types

import (
	"fmt"
	"strings"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeTickSpacingDecrease = "TickSpacingDecrease"
)

func init() {
	govtypesv1.RegisterProposalType(ProposalTypeTickSpacingDecrease)
}

var (
	_ govtypesv1.Content = &TickSpacingDecreaseProposal{}
)

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

func NewTickSpacingDecreaseProposal(title, description string, records []PoolIdToTickSpacingRecord) govtypesv1.Content {
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
	err := govtypesv1.ValidateAbstract(p)
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
