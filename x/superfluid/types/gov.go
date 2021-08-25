package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeSetSuperfluidAssets     = "SetSuperfluidAssets"
	ProposalTypeEnableSuperfluidAssets  = "EnableSuperfluidAssets"
	ProposalTypeDisableSuperfluidAssets = "DisableSuperfluidAssets"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeSetSuperfluidAssets)
	govtypes.RegisterProposalTypeCodec(&SetSuperfluidAssetsProposal{}, "osmosis/SetSuperfluidAssetsProposal")
	govtypes.RegisterProposalType(ProposalTypeEnableSuperfluidAssets)
	govtypes.RegisterProposalTypeCodec(&EnableSuperfluidAssetsProposal{}, "osmosis/EnableSuperfluidAssetsProposal")
	govtypes.RegisterProposalType(ProposalTypeEnableSuperfluidAssets)
	govtypes.RegisterProposalTypeCodec(&DisableSuperfluidAssetsProposal{}, "osmosis/DisableSuperfluidAssetsProposal")
}

var _ govtypes.Content = &SetSuperfluidAssetsProposal{}
var _ govtypes.Content = &EnableSuperfluidAssetsProposal{}
var _ govtypes.Content = &DisableSuperfluidAssetsProposal{}

func NewSetSuperfluidAssetsProposal(title, description string, assets []SuperfluidAsset) govtypes.Content {
	return &SetSuperfluidAssetsProposal{
		Title:       title,
		Description: description,
		Assets:      assets,
	}
}

func (p *SetSuperfluidAssetsProposal) GetTitle() string { return p.Title }

func (p *SetSuperfluidAssetsProposal) GetDescription() string { return p.Description }

func (p *SetSuperfluidAssetsProposal) ProposalRoute() string { return RouterKey }

func (p *SetSuperfluidAssetsProposal) ProposalType() string {
	return ProposalTypeSetSuperfluidAssets
}

func (p *SetSuperfluidAssetsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return nil
}

func (p SetSuperfluidAssetsProposal) String() string {
	return fmt.Sprintf(`Set Superfluid Assets Proposal:
	Title:       %s
	Description: %s
	Assets:     %+v
  `, p.Title, p.Description, p.Assets)
}

func NewEnableSuperfluidAssetsProposal(title, description string, assetIds []uint64) govtypes.Content {
	return &EnableSuperfluidAssetsProposal{
		Title:              title,
		Description:        description,
		SuperfluidAssetIds: assetIds,
	}
}

func (p *EnableSuperfluidAssetsProposal) GetTitle() string { return p.Title }

func (p *EnableSuperfluidAssetsProposal) GetDescription() string { return p.Description }

func (p *EnableSuperfluidAssetsProposal) ProposalRoute() string { return RouterKey }

func (p *EnableSuperfluidAssetsProposal) ProposalType() string {
	return ProposalTypeEnableSuperfluidAssets
}

func (p *EnableSuperfluidAssetsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return nil
}

func (p EnableSuperfluidAssetsProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Enable Superfluid Assets Proposal:
  Title:       %s
  Description: %s
  SuperfluidAssetIds:     %+v
`, p.Title, p.Description, p.SuperfluidAssetIds))
	return b.String()
}

func NewDisableSuperfluidAssetsProposal(title, description string, assetIds []uint64) govtypes.Content {
	return &DisableSuperfluidAssetsProposal{
		Title:              title,
		Description:        description,
		SuperfluidAssetIds: assetIds,
	}
}

func (p *DisableSuperfluidAssetsProposal) GetTitle() string { return p.Title }

func (p *DisableSuperfluidAssetsProposal) GetDescription() string { return p.Description }

func (p *DisableSuperfluidAssetsProposal) ProposalRoute() string { return RouterKey }

func (p *DisableSuperfluidAssetsProposal) ProposalType() string {
	return ProposalTypeDisableSuperfluidAssets
}

func (p *DisableSuperfluidAssetsProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return nil
}

func (p DisableSuperfluidAssetsProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Disable Superfluid Assets Proposal:
  Title:       %s
  Description: %s
  SuperfluidAssetIds:     %+v
`, p.Title, p.Description, p.SuperfluidAssetIds))
	return b.String()
}
