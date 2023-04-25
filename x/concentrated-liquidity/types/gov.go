package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeTickSpacingDecrease = "TickSpacingDecrease"
)

// Init registers proposal to decrease tick spacing.
func init() {
	govtypes.RegisterProposalType(ProposalTypeTickSpacingDecrease)
	govtypes.RegisterProposalTypeCodec(&TickSpacingDecreaseProposal{}, "osmosis/TickSpacingDecreaseProposal")
}

var (
	_ govtypes.Content = &TickSpacingDecreaseProposal{}
)

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
