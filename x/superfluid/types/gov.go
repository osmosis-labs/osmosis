package types

import (
	"errors"
	"fmt"
	"strings"

	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

const (
	ProposalTypeSetSuperfluidAssets    = "SetSuperfluidAssets"
	ProposalTypeRemoveSuperfluidAssets = "RemoveSuperfluidAssets"
	ProposalTypeUpdateUnpoolWhitelist  = "UpdateUnpoolWhitelist"
)

func init() {
	govtypesv1.RegisterProposalType(ProposalTypeSetSuperfluidAssets)
	govtypesv1.RegisterProposalType(ProposalTypeRemoveSuperfluidAssets)
	govtypesv1.RegisterProposalType(ProposalTypeUpdateUnpoolWhitelist)
}

var (
	_ govtypesv1.Content = &SetSuperfluidAssetsProposal{}
	_ govtypesv1.Content = &RemoveSuperfluidAssetsProposal{}
	_ govtypesv1.Content = &UpdateUnpoolWhiteListProposal{}
)

func NewSetSuperfluidAssetsProposal(title, description string, assets []SuperfluidAsset) govtypesv1.Content {
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
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, asset := range p.Assets {
		switch asset.AssetType {
		case SuperfluidAssetTypeLPShare:
			if _, err := gammtypes.GetPoolIdFromShareDenom(asset.Denom); err != nil {
				return err
			}
			// Denom must be from GAMM
			if !strings.HasPrefix(asset.Denom, gammtypes.GAMMTokenPrefix) {
				return fmt.Errorf("denom %s must be from GAMM", asset.Denom)
			}
		case SuperfluidAssetTypeConcentratedShare:
			if _, err := cltypes.GetPoolIdFromShareDenom(asset.Denom); err != nil {
				return err
			}
			// Denom must be from CL
			if !strings.HasPrefix(asset.Denom, cltypes.ConcentratedLiquidityTokenPrefix) {
				return fmt.Errorf("denom %s must be from CL", asset.Denom)
			}
		default:
			return errors.New("unsupported superfluid asset type")
		}
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

func NewRemoveSuperfluidAssetsProposal(title, description string, denoms []string) govtypesv1.Content {
	return &RemoveSuperfluidAssetsProposal{
		Title:                 title,
		Description:           description,
		SuperfluidAssetDenoms: denoms,
	}
}

func (p *RemoveSuperfluidAssetsProposal) GetTitle() string { return p.Title }

func (p *RemoveSuperfluidAssetsProposal) GetDescription() string { return p.Description }

func (p *RemoveSuperfluidAssetsProposal) ProposalRoute() string { return RouterKey }

func (p *RemoveSuperfluidAssetsProposal) ProposalType() string {
	return ProposalTypeRemoveSuperfluidAssets
}

func (p *RemoveSuperfluidAssetsProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	return nil
}

func (p RemoveSuperfluidAssetsProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Superfluid Assets Proposal:
  Title:       %s
  Description: %s
  SuperfluidAssetDenoms:     %+v
`, p.Title, p.Description, p.SuperfluidAssetDenoms))
	return b.String()
}

func NewUpdateUnpoolWhitelistProposal(title, description string, poolIds []uint64, isOverwrite bool) govtypesv1.Content {
	return &UpdateUnpoolWhiteListProposal{
		Title:       title,
		Description: description,
		Ids:         poolIds,
		IsOverwrite: isOverwrite,
	}
}

func (p *UpdateUnpoolWhiteListProposal) GetTitle() string { return p.Title }

func (p *UpdateUnpoolWhiteListProposal) GetDescription() string { return p.Description }

func (p *UpdateUnpoolWhiteListProposal) ProposalRoute() string { return RouterKey }

func (p *UpdateUnpoolWhiteListProposal) ProposalType() string {
	return ProposalTypeUpdateUnpoolWhitelist
}

func (p *UpdateUnpoolWhiteListProposal) ValidateBasic() error {
	err := govtypesv1.ValidateAbstract(p)
	if err != nil {
		return err
	}

	for _, id := range p.Ids {
		if id == 0 {
			return errors.New("pool id cannot be 0")
		}
	}

	return nil
}

func (p UpdateUnpoolWhiteListProposal) String() string {
	return fmt.Sprintf(`Update Unpool Whitelist Assets Proposal:
	Title:       %s
	Description: %s
	Pool Ids:     %+v
	IsOverwrite:  %t
  `, p.Title, p.Description, p.Ids, p.IsOverwrite)
}
