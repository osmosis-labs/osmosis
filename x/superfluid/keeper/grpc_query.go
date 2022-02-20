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

// AssetMultiplier returns superfluid asset multiplier
func (k Keeper) AssetMultiplier(goCtx context.Context, req *types.AssetMultiplierRequest) (*types.AssetMultiplierResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	epochInfo := k.ek.GetEpochInfo(ctx, params.RefreshEpochIdentifier)

	return &types.AssetMultiplierResponse{
		OsmoEquivalentMultiplier: &types.OsmoEquivalentMultiplierRecord{
			EpochNumber: epochInfo.CurrentEpoch,
			Denom:       req.Denom,
			Multiplier:  k.GetOsmoEquivalentMultiplier(ctx, req.Denom),
		},
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

// ConnectedIntermediaryAccount returns intermediary account connected to a superfluid staked lock by id
func (k Keeper) ConnectedIntermediaryAccount(goCtx context.Context, req *types.ConnectedIntermediaryAccountRequest) (*types.ConnectedIntermediaryAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := k.GetLockIdIntermediaryAccountConnection(ctx, req.LockId)
	acc := k.GetIntermediaryAccount(ctx, address)

	return &types.ConnectedIntermediaryAccountResponse{
		Account: &types.SuperfluidIntermediaryAccountInfo{
			Denom:   acc.Denom,
			ValAddr: acc.ValAddr,
			GaugeId: acc.GaugeId,
			Address: acc.GetAccAddress().String(),
		},
	}, nil
}
