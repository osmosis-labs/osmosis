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
	asset types.Asset,
	newStatus types.AssetStatus,
) (ChangeAssetStatusResult, error) {
	// get current params
	params := k.GetParams(ctx)

	// check if the specified asset is known
	const notFoundIdx = -1
	var assetIdx = notFoundIdx
	for i := range params.Assets {
		if params.Assets[i].Asset == asset {
			assetIdx = i
			break
		}
	}
	if assetIdx == notFoundIdx {
		return ChangeAssetStatusResult{}, errorsmod.Wrapf(types.ErrInvalidAsset, "Asset not found")
	}

	// update assetIdx asset status
	oldStatus := params.Assets[assetIdx].AssetStatus
	params.Assets[assetIdx].AssetStatus = newStatus
	k.SetParam(ctx, types.KeyAssets, params.Assets)

	return ChangeAssetStatusResult{
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}, nil
}

// createAssets creates tokenfactory denoms for all provided assets
func (k Keeper) createAssets(ctx sdk.Context, assets []types.AssetWithStatus) error {
	bridgeModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	handler := k.router.Handler(new(tokenfactorytypes.MsgCreateDenom))
	if handler == nil {
		return errorsmod.Wrapf(types.ErrTokenfactory, "Can't route a create denom message")
	}

	for _, asset := range assets {
		msgCreateDenom := &tokenfactorytypes.MsgCreateDenom{
			Sender:   bridgeModuleAddr.String(),
			Subdenom: asset.Asset.Name(),
		}

		// ignore resp since it is not needed in this method
		// TODO: double-check if we need to handle the response
		_, err := handler(ctx, msgCreateDenom)
		if err != nil {
			return errorsmod.Wrapf(
				types.ErrTokenfactory,
				"Can't execute a create denom message for %s: %s", asset.Asset.Name(), err,
			)
		}
	}

	return nil
}
