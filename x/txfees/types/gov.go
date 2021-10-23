package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeUpdateFeeToken = "UpdateFeeToken"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateFeeToken)
	govtypes.RegisterProposalTypeCodec(&UpdateFeeTokenProposal{}, "osmosis/UpdateFeeTokenProposal")
}

var _ govtypes.Content = &UpdateFeeTokenProposal{}

func NewUpdateFeeTokenProposal(title, description string, feeToken FeeToken) UpdateFeeTokenProposal {
	return UpdateFeeTokenProposal{
		Title:       title,
		Description: description,
		Feetoken:    feeToken,
	}
}

func (p *UpdateFeeTokenProposal) GetTitle() string { return p.Title }

func (p *UpdateFeeTokenProposal) GetDescription() string { return p.Description }

func (p *UpdateFeeTokenProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateFeeTokenProposal) ProposalType() string {
	return ProposalTypeUpdateFeeToken
}

func (p *UpdateFeeTokenProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return sdk.ValidateDenom(p.Feetoken.Denom)
}

func (p UpdateFeeTokenProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Fee Token Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, p.Feetoken.String()))
	return b.String()
}
