package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func HandleSetSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper, p *types.SetSuperfluidAssetsProposal) error {
	for _, asset := range p.Assets {
		// initialize osmo equivalent multipliers
		epochIdentifier := k.GetParams(ctx).RefreshEpochIdentifier
		currentEpoch := ek.GetEpochInfo(ctx, epochIdentifier).CurrentEpoch
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			k.SetSuperfluidAsset(ctx, asset)
			err := k.UpdateOsmoEquivalentMultipliers(ctx, asset, currentEpoch)
			return err
		})
		event := sdk.NewEvent(
			types.TypeEvtSetSuperfluidAsset,
			sdk.NewAttribute(types.AttributeDenom, asset.Denom),
			sdk.NewAttribute(types.AttributeSuperfluidAssetType, asset.AssetType.String()),
		)
		ctx.EventManager().EmitEvent(event)
	}
	return nil
}

func HandleRemoveSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, p *types.RemoveSuperfluidAssetsProposal) error {
	for _, denom := range p.SuperfluidAssetDenoms {
		asset := k.GetSuperfluidAsset(ctx, denom)
		dummyAsset := types.SuperfluidAsset{}
		if asset == dummyAsset {
			return fmt.Errorf("superfluid asset %s doesn't exist", denom)
		}
		k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
		event := sdk.NewEvent(
			types.TypeEvtRemoveSuperfluidAsset,
			sdk.NewAttribute(types.AttributeDenom, denom),
		)
		ctx.EventManager().EmitEvent(event)
	}
	return nil
}
