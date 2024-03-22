package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

type ChangeAssetStatusResult struct {
	OldStatus types.AssetStatus
	NewStatus types.AssetStatus
}

// ChangeAssetStatus changes the status of the provided asset to newStatus.
// Returns error if the provided asset is not found in the module params.
func (k Keeper) ChangeAssetStatus(
	ctx sdk.Context,
	assetID types.AssetID,
	newStatus types.AssetStatus,
) (ChangeAssetStatusResult, error) {
	// get current params
	params := k.GetParams(ctx)

	// check if the specified asset is known
	assetIdx := params.GetAssetIndex(assetID)
	if assetIdx == notFoundIdx {
		return ChangeAssetStatusResult{}, errorsmod.Wrapf(types.ErrInvalidAssetID, "Asset not found")
	}

	// update assetIdx asset status
	oldStatus := params.Assets[assetIdx].Status
	params.Assets[assetIdx].Status = newStatus
	k.SetParam(ctx, types.KeyAssets, params.Assets)

	return ChangeAssetStatusResult{
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}, nil
}

// createAssets creates tokenfactory denoms for all provided assets and properly sets
// last_transfer_height values.
func (k Keeper) createAssets(ctx sdk.Context, assets []types.Asset) ([]types.Asset, error) {
	handler := k.router.Handler(new(tokenfactorytypes.MsgCreateDenom))
	if handler == nil {
		return nil, errorsmod.Wrapf(types.ErrTokenfactory, "Can't route a create denom message")
	}

	bridgeModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	createdAssets := make([]types.Asset, 0, len(assets))

	for _, asset := range assets {
		msgCreateDenom := &tokenfactorytypes.MsgCreateDenom{
			Sender:   bridgeModuleAddr.String(),
			Subdenom: asset.Name(),
		}

		// ignore resp since it is not needed in this method
		// TODO: double-check if we need to handle the response
		_, err := handler(ctx, msgCreateDenom)
		if err != nil {
			return nil, errorsmod.Wrapf(
				types.ErrTokenfactory,
				"Can't execute a create denom message for %s: %s", asset.Name(), err,
			)
		}

		// TODO: set the last_transfer_height to the latest external blockchain height, since using 0
		//  doesn't really make sense. Should use corresponding chain clients here after they are implemented.
		asset.LastTransferHeight = 0

		createdAssets = append(createdAssets, asset)
	}

	return createdAssets, nil
}
