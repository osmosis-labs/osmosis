package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) HandleSetSuperfluidAssetsProposal(ctx sdk.Context, p *types.SetSuperfluidAssetsProposal) error {
	for _, asset := range p.Assets {
		k.SetSuperfluidAsset(ctx, asset)
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtSetSuperfluidAsset,
				sdk.NewAttribute(types.AttributeDenom, asset.Denom),
				sdk.NewAttribute(types.AttributeSuperfluidAssetType, asset.AssetType.String()),
			),
		})
	}
	return nil
}

func (k Keeper) HandleRemoveSuperfluidAssetsProposal(ctx sdk.Context, p *types.RemoveSuperfluidAssetsProposal) error {
	for _, denom := range p.SuperfluidAssetDenoms {
		k.DeleteSuperfluidAsset(ctx, denom)
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtRemoveSuperfluidAsset,
				sdk.NewAttribute(types.AttributeDenom, denom),
			),
		})
	}
	return nil
}
