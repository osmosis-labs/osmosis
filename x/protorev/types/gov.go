package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeSetProtoRevEnabled      = "SetProtoRevEnabledProposal"
	ProposalTypeSetProtoRevAdminAccount = "SetProtoRevAdminAccountProposal"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeSetProtoRevEnabled)
	govtypes.RegisterProposalTypeCodec(&SetProtoRevEnabledProposal{}, "osmosis/SetProtoRevEnabledProposal")
	govtypes.RegisterProposalType(ProposalTypeSetProtoRevAdminAccount)
	govtypes.RegisterProposalTypeCodec(&SetProtoRevAdminAccountProposal{}, "osmosis/SetProtoRevAdminAccountProposal")
}

var (
	_ govtypes.Content = &SetProtoRevEnabledProposal{}
	_ govtypes.Content = &SetProtoRevAdminAccountProposal{}
)

// ---------------- Interface for SetProtoRevEnabledProposal ---------------- //
func NewSetProtoRevEnabledProposal(title, description string, enabled bool) govtypes.Content {
	return &SetProtoRevEnabledProposal{title, description, enabled}
}

func (p *SetProtoRevEnabledProposal) GetTitle() string { return p.Title }

func (p *SetProtoRevEnabledProposal) GetDescription() string { return p.Description }

func (p *SetProtoRevEnabledProposal) ProposalRoute() string { return RouterKey }

func (p *SetProtoRevEnabledProposal) ProposalType() string {
	return ProposalTypeSetProtoRevEnabled
}

func (p *SetProtoRevEnabledProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return ValidateBoolean(p.Enabled)
}

func (p SetProtoRevEnabledProposal) String() string {
	return fmt.Sprintf(`Set ProtoRev Enabled Proposal:
	Title:       %s
	Description: %s
	ProtoRev Enabled:     %+v
  `, p.Title, p.Description, p.Enabled)
}

// ---------------- Interface for SetProtoRevAdminAccountProposal ---------------- //
func NewSetProtoRevAdminAccountProposal(title, description string, account string) govtypes.Content {
	return &SetProtoRevAdminAccountProposal{title, description, account}
}

func (p *SetProtoRevAdminAccountProposal) GetTitle() string { return p.Title }

func (p *SetProtoRevAdminAccountProposal) GetDescription() string { return p.Description }

func (p *SetProtoRevAdminAccountProposal) ProposalRoute() string { return RouterKey }

func (p *SetProtoRevAdminAccountProposal) ProposalType() string {
	return ProposalTypeSetProtoRevAdminAccount
}

func (p *SetProtoRevAdminAccountProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	_, err = sdk.AccAddressFromBech32(p.Account)
	return err
}

func (p SetProtoRevAdminAccountProposal) String() string {
	return fmt.Sprintf(`Set ProtoRev Admin Account Proposal:
	Title:       %s
	Description: %s
	ProtoRev Admin Account:     %+v
  `, p.Title, p.Description, p.Account)
}
