package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeCreateConcentratedLiquidityPool = "CreateConcentratedLiquidityPool"
	ProposalTypeTickSpacingDecrease             = "TickSpacingDecrease"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateConcentratedLiquidityPool)
	govtypes.RegisterProposalTypeCodec(&CreateConcentratedLiquidityPoolProposal{}, "osmosis/CreateConcentratedLiquidityPoolProposal")
	govtypes.RegisterProposalType(ProposalTypeTickSpacingDecrease)
	govtypes.RegisterProposalTypeCodec(&TickSpacingDecreaseProposal{}, "osmosis/TickSpacingDecreaseProposal")
}

var (
	_ govtypes.Content = &CreateConcentratedLiquidityPoolProposal{}
	_ govtypes.Content = &TickSpacingDecreaseProposal{}
)

// NewCreateConcentratedLiquidityPoolProposal returns a new instance of a create concentrated liquidity pool proposal struct.
func NewCreateConcentratedLiquidityPoolProposal(title, description string, denom0, denom1 string, tickSpacing uint64, exponentAtPriceOne sdk.Int, swapFee sdk.Dec) govtypes.Content {
	return &CreateConcentratedLiquidityPoolProposal{
		Title:              title,
		Description:        description,
		Denom0:             denom0,
		Denom1:             denom1,
		TickSpacing:        tickSpacing,
		ExponentAtPriceOne: exponentAtPriceOne,
		SwapFee:            swapFee,
	}
}

func (p *CreateConcentratedLiquidityPoolProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *CreateConcentratedLiquidityPoolProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *CreateConcentratedLiquidityPoolProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *CreateConcentratedLiquidityPoolProposal) ProposalType() string {
	return ProposalTypeCreateConcentratedLiquidityPool
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *CreateConcentratedLiquidityPoolProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	if p.TickSpacing <= 0 {
		return fmt.Errorf("tick spacing must be positive")
	}

	if p.Denom0 == p.Denom1 {
		return fmt.Errorf("denom0 and denom1 must be different")
	}

	if sdk.ValidateDenom(p.Denom0) != nil {
		return fmt.Errorf("denom0 is invalid: %s", sdk.ValidateDenom(p.Denom0))
	}

	if sdk.ValidateDenom(p.Denom1) != nil {
		return fmt.Errorf("denom1 is invalid: %s", sdk.ValidateDenom(p.Denom1))
	}

	swapFee := p.SwapFee
	if swapFee.IsNegative() || swapFee.GTE(sdk.OneDec()) {
		return InvalidSwapFeeError{ActualFee: swapFee}
	}

	return nil
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

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p CreateConcentratedLiquidityPoolProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create Concentrated Liquidity Pool Proposal:
  Title:                 %s
  Description:           %s
  Denom0:                %s
  Denom1:                %s
  Tick Spacing:          %d
  ExponentAtPriceOne     %s
  Swap Fee:              %s
`, p.Title, p.Description, p.Denom0, p.Denom1, p.TickSpacing, p.ExponentAtPriceOne.String(), p.SwapFee.String()))
	return b.String()
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
