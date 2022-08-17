package gov

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v11/x/superfluid/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HandleSetSuperfluidAssetsProposal(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper, p *types.SetSuperfluidAssetsProposal) error {
	for _, asset := range p.Assets {
		k.AddNewSuperfluidAsset(ctx, asset)
		events.EmitSetSuperfluidAssetEvent(ctx, asset.Denom, asset.AssetType)
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
		events.EmitRemoveSuperfluidAsset(ctx, denom)
	}
	return nil
}
