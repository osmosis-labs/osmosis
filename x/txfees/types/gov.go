package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdateFeeToken = "UpdateFeeToken"
)

func init() {
	govtypesv1.RegisterProposalType(ProposalTypeUpdateFeeToken)
}

var _ govtypesv1.Content = &UpdateFeeTokenProposal{}

func NewUpdateFeeTokenProposal(title, description string, feeTokens []FeeToken) UpdateFeeTokenProposal {
	return UpdateFeeTokenProposal{
		Title:       title,
		Description: description,
		Feetokens:   feeTokens,
	}
}

func (p *UpdateFeeTokenProposal) GetTitle() string { return p.Title }

func (p *UpdateFeeTokenProposal) GetDescription() string { return p.Description }

func (p *UpdateFeeTokenProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateFeeTokenProposal) ProposalType() string {
	return ProposalTypeUpdateFeeToken
}

func (p *UpdateFeeTokenProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, feeToken := range p.Feetokens {
		if err := sdk.ValidateDenom(feeToken.Denom); err != nil {
			return err
		}
	}
	return nil
}

func (p UpdateFeeTokenProposal) String() string {
	var b strings.Builder
	for _, feeToken := range p.Feetokens {
		b.WriteString(fmt.Sprintf("(Denom: %s, PoolID: %d) ", feeToken.Denom, feeToken.PoolID))
	}

	recordsStr := b.String()
	b.Reset()

	b.WriteString(fmt.Sprintf(`Update Fee Token Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))

	return b.String()
}
