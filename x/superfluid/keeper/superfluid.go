package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (k Keeper) SetSuperfluidAsset(ctx sdk.Context, asset types.SuperfluidAsset) {

}

func (k Keeper) GetSuperfluidAsset(ctx sdk.Context, denom string) types.SuperfluidAsset {
	return types.SuperfluidAsset{}
}

func (k Keeper) GetAllSuperfluidAssets(ctx sdk.Context) []types.SuperfluidAsset {
	return []types.SuperfluidAsset{}
}

func (k Keeper) SetEnabledSuperfluidAssetIds(ctx sdk.Context) {

}

func (k Keeper) SetSuperfluidAssetInfo(ctx sdk.Context, assetInfo types.SuperfluidAssetInfo) {

}

func (k Keeper) GetSuperfluidAssetInfo(ctx sdk.Context, denom string) types.SuperfluidAssetInfo {
	return types.SuperfluidAssetInfo{}
}

func (k Keeper) GetAllSuperfluidAssetInfos(ctx sdk.Context) []types.SuperfluidAssetInfo {
	return []types.SuperfluidAssetInfo{}
}

func (k Keeper) GetEnabledSuperfluidAssetInfos() []types.SuperfluidAssetInfo {
	return []types.SuperfluidAssetInfo{}
}
