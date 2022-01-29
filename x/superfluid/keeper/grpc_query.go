package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

var _ types.QueryServer = Keeper{}

// AssetType Returns superfluid asset type
func (k Keeper) AssetType(goCtx context.Context, req *types.AssetTypeRequest) (*types.AssetTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	asset := k.GetSuperfluidAsset(ctx, req.Denom)
	return &types.AssetTypeResponse{
		AssetType: asset.AssetType,
	}, nil
}

// AllAssets Returns all superfluid assets info
func (k Keeper) AllAssets(goCtx context.Context, req *types.AllAssetsRequest) (*types.AllAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assets := k.GetAllSuperfluidAssets(ctx)
	return &types.AllAssetsResponse{
		Assets: assets,
	}, nil
}
