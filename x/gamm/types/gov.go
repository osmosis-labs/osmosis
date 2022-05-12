package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeUpdatePoolParams = "UpdatePoolParams"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdatePoolParams)
	govtypes.RegisterProposalTypeCodec(&UpdatePoolParamsProposal{}, "osmosis/UpdatePoolParamsProposal")
}

var (
	_ govtypes.Content = &UpdatePoolParamsProposal{}
)

func NewUpdatePoolParamsProposal(title, description string, updates []UpdatePoolParam) govtypes.Content {
	return &UpdatePoolParamsProposal{
		Title:       title,
		Description: description,
		Updates:     updates,
	}
}

func (p *UpdatePoolParamsProposal) GetTitle() string { return p.Title }

func (p *UpdatePoolParamsProposal) GetDescription() string { return p.Description }

func (p *UpdatePoolParamsProposal) ProposalRoute() string { return RouterKey }

func (p *UpdatePoolParamsProposal) ProposalType() string {
	return ProposalTypeUpdatePoolParams
}

func (p *UpdatePoolParamsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Updates) == 0 {
		return ErrEmptyProposalUpdates
	}

	for _, record := range p.Updates {
		if err := record.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func (p UpdatePoolParamsProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, metadata := range p.Updates {
		recordsStr = recordsStr + fmt.Sprintf("(%+v), ", metadata.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Pool Metadata Proposal:
  Title:       %s
  Description: %s
  Updates:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

func (data UpdatePoolParam) ValidateBasic() error {
	if data.PoolId == 0 {
		return fmt.Errorf("Pool ID is empty")
	}
	return nil
}

func (risk RiskLevel) IsEmpty() bool {
	return risk == RiskLevel_DEFAULT_SAFE
}

func (risk RiskLevel) IsSafe() bool {
	return risk <= RiskLevel_SAFE
}

func (risk RiskLevel) IsUnpoolAllowed() bool {
	return risk >= RiskLevel_UNPOOL_ALLOWED
}
