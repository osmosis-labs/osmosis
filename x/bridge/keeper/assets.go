package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type ChangeAssetStatusResult struct {
	OldStatus types.AssetStatus
	NewStatus types.AssetStatus
}

func (k Keeper) ChangeAssetStatus(
	ctx sdk.Context,
	asset types.Asset,
	newStatus types.AssetStatus,
) (ChangeAssetStatusResult, error) {
	// get current params
	params := k.GetParams(ctx)

	// check if the specified asset is known
	const notFound = -1
	var found = notFound
	for i := range params.Assets {
		if params.Assets[i].Asset == asset {
			found = i
		}
	}
	if found == notFound {
		return ChangeAssetStatusResult{}, errorsmod.Wrapf(types.ErrInvalidAsset, "Asset not found")
	}

	// update found asset status
	oldStatus := params.Assets[found].AssetStatus
	params.Assets[found].AssetStatus = newStatus
	k.SetParams(ctx, params)

	return ChangeAssetStatusResult{
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}, nil
}

// createAssets creates tokenfactory denoms for all provided assets
func (k Keeper) createAssets(ctx sdk.Context, assets []types.AssetWithStatus) error {
	bridgeModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	for _, asset := range assets {
		_, err := k.tokenFactoryKeeper.CreateDenom(ctx, bridgeModuleAddr.String(), asset.Asset.Name())
		if err != nil {
			return fmt.Errorf("can't create a new denom %s: %s", asset.Asset.Name(), err)
		}
	}

	return nil
}
