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

// AssetTwap returns superfluid asset TWAP
func (k Keeper) AssetTwap(goCtx context.Context, req *types.AssetTwapRequest) (*types.AssetTwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	twap := k.GetCurrentEpochOsmoEquivalentTWAP(ctx, req.Denom)
	return &types.AssetTwapResponse{
		Twap: &twap,
	}, nil
}

// AllIntermediaryAccounts returns all superfluid intermediary accounts
func (k Keeper) AllIntermediaryAccounts(goCtx context.Context, req *types.AllIntermediaryAccountsRequest) (*types.AllIntermediaryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accounts := k.GetAllIntermediaryAccounts(ctx)
	accInfos := []types.SuperfluidIntermediaryAccountInfo{}
	for _, acc := range accounts {
		accInfos = append(accInfos, types.SuperfluidIntermediaryAccountInfo{
			Denom:   acc.Denom,
			ValAddr: acc.ValAddr,
			GaugeId: acc.GaugeId,
			Address: acc.GetAccAddress().String(),
		})
	}
	return &types.AllIntermediaryAccountsResponse{
		Accounts: accInfos,
	}, nil
}
