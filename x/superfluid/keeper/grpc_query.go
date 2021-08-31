package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
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

// AssetInfo Returns superfluid asset info
func (k Keeper) AssetInfo(goCtx context.Context, req *types.AssetInfoRequest) (*types.AssetInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assetInfo := k.GetSuperfluidAssetInfo(ctx, req.Denom)
	return &types.AssetInfoResponse{
		Info: &assetInfo,
	}, nil
}
