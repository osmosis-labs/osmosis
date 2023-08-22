package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeDenomPairTakerFee = "DenomPairTakerFee"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeDenomPairTakerFee)
	govtypes.RegisterProposalTypeCodec(&DenomPairTakerFeeProposal{}, "osmosis/DenomPairTakerFeeProposal")
}

var (
	_ govtypes.Content = &DenomPairTakerFeeProposal{}
)

// NewDenomPairTakerFeeProposal returns a new instance of a denom pair taker fee proposal struct.
func NewDenomPairTakerFeeProposal(title, description string, records []DenomPairTakerFee) govtypes.Content {
	return &DenomPairTakerFeeProposal{
		Title:             title,
		Description:       description,
		DenomPairTakerFee: records,
	}
}

func (p *DenomPairTakerFeeProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *DenomPairTakerFeeProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *DenomPairTakerFeeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *DenomPairTakerFeeProposal) ProposalType() string {
	return ProposalTypeDenomPairTakerFee
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *DenomPairTakerFeeProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, record := range p.DenomPairTakerFee {
		if record.Denom0 == record.Denom1 {
			return fmt.Errorf("denom0 and denom1 must be different")
		}

		if sdk.ValidateDenom(record.Denom0) != nil {
			return fmt.Errorf("denom0 is invalid: %s", sdk.ValidateDenom(record.Denom0))
		}

		if sdk.ValidateDenom(record.Denom1) != nil {
			return fmt.Errorf("denom1 is invalid: %s", sdk.ValidateDenom(record.Denom1))
		}

		takerFee := record.TakerFee
		if takerFee.IsNegative() || takerFee.GTE(sdk.OneDec()) {
			return fmt.Errorf("taker fee must be between 0 and 1: %s", takerFee.String())
		}
	}
	return nil
}

// String returns a string containing the denom pair taker fee proposal.
func (p DenomPairTakerFeeProposal) String() string {
	recordsStr := ""
	for _, record := range p.DenomPairTakerFee {
		recordsStr = recordsStr + fmt.Sprintf("(Denom0: %s, Denom1: %s, TakerFee: %s) ", record.Denom0, record.Denom1, record.TakerFee.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Denom Pair Taker Fee Proposal:
Title:       %s
Description: %s
Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
