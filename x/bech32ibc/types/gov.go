package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeUpdateHrpIBCRecord = "UpdateHrpIBCRecord"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateHrpIBCRecord)
	govtypes.RegisterProposalTypeCodec(&UpdateHrpIBCRecordProposal{}, "osmosis/UpdateHrpIBCRecord")
}

var _ govtypes.Content = &UpdateHrpIBCRecordProposal{}

func NewUpdateHrpIBCRecordProposal(title, description, hrp, sourceChannel string) govtypes.Content {
	return &UpdateHrpIBCRecordProposal{
		Title:         title,
		Description:   description,
		Hrp:           hrp,
		SourceChannel: sourceChannel,
	}
}

func (p *UpdateHrpIBCRecordProposal) GetTitle() string { return p.Title }

func (p *UpdateHrpIBCRecordProposal) GetDescription() string { return p.Description }

func (p *UpdateHrpIBCRecordProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateHrpIBCRecordProposal) ProposalType() string {
	return ProposalTypeUpdateHrpIBCRecord
}

func (p *UpdateHrpIBCRecordProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	return ValidateHRP(p.Hrp)
}

func (p UpdateHrpIBCRecordProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update HrpIBCRecord Proposal:
  Title:          %s
  Description:    %s
  HRP:            %s
  Source Channel: %s
`, p.Title, p.Description, p.Hrp, p.SourceChannel))
	return b.String()
}
