package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type UpdateParamsResult struct {
	signersToCreate []string
	signersToDelete []string
	assetsToCreate  []types.AssetWithStatus
	assetsToDelete  []types.AssetWithStatus
}

// UpdateParams properly updates params of the module.
func (k Keeper) UpdateParams(ctx sdk.Context, newParams types.UpdateParams) UpdateParamsResult {
	var (
		newAssets = osmoutils.Map(newParams.Assets, func(v types.Asset) types.AssetWithStatus {
			return types.AssetWithStatus{
				Asset:       v,
				AssetStatus: types.AssetStatus_ASSET_STATUS_BLOCKED_BOTH,
			}
		})

		oldParams = k.GetParams(ctx)

		signersToCreate = osmoutils.Difference(newParams.Signers, oldParams.Signers)
		signersToDelete = osmoutils.Difference(oldParams.Signers, newParams.Signers)
		assetsToCreate  = osmoutils.Difference(newAssets, oldParams.Assets)
		assetsToDelete  = osmoutils.Difference(oldParams.Assets, newAssets)
	)

	bridgeModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// create denoms for all new assets
	for _, asset := range assetsToCreate {
		_, err := k.tokenFactoryKeeper.CreateDenom(ctx, bridgeModuleAddr.String(), asset.Asset.Denom)
		if err != nil {
			panic(fmt.Sprintf("can't create a new denom %s: %s", asset.Asset.Denom, err))
		}
	}

	// TODO: handle signers creation and deletion and asset deletion

	return UpdateParamsResult{
		signersToCreate: signersToCreate,
		signersToDelete: signersToDelete,
		assetsToCreate:  assetsToCreate,
		assetsToDelete:  assetsToDelete,
	}
}

// setParams sets the total set of params.
func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}
