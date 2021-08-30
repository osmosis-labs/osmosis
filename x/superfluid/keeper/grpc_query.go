package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

var _ types.QueryServer = Keeper{}

// AssetType Returns superfluid asset type
func (k Keeper) AssetType(goCtx context.Context, req *types.AssetTypeRequest) (*types.AssetTypeResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: implement this
	return &types.AssetTypeResponse{}, nil
}

// AllAssets Returns all superfluid assets info
func (k Keeper) AllAssets(goCtx context.Context, req *types.AllAssetsRequest) (*types.AllAssetsResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: implement this
	return &types.AllAssetsResponse{}, nil
}

// EnabledAssets Returns enabled superfluid assets
func (k Keeper) EnabledAssets(goCtx context.Context, req *types.EnabledAssetsRequest) (*types.EnabledAssetsResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: implement this
	return &types.EnabledAssetsResponse{}, nil
}

// AssetInfo Returns superfluid asset info
func (k Keeper) AssetInfo(goCtx context.Context, req *types.AssetInfoRequest) (*types.AssetInfoResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: implement this
	return &types.AssetInfoResponse{}, nil
}
