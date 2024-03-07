package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type UpdateParamsResult struct {
	signersToCreate []string
	signersToDelete []string
	assetsToCreate  []types.AssetWithStatus
	assetsToDelete  []types.AssetWithStatus
}

// UpdateParams properly updates params of the module.
func (k Keeper) UpdateParams(ctx sdk.Context, newParams types.Params) UpdateParamsResult {
	var (
		oldParams = k.GetParams(ctx)

		signersToCreate = Difference(newParams.Signers, oldParams.Signers)
		signersToDelete = Difference(oldParams.Signers, newParams.Signers)
		assetsToCreate  = Difference(newParams.Assets, oldParams.Assets)
		assetsToDelete  = Difference(oldParams.Assets, newParams.Assets)

		bridgeModuleAddr = k.accountKeeper.GetModuleAddress(types.ModuleName)
	)

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

// Difference returns the slice of elements that are elements of a but not elements of b.
// TODO: Placed here temporarily. Delete after releasing the new osmoutils version.
func Difference[T comparable](a, b []T) []T {
	mb := make(map[T]struct{}, len(a))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	diff := make([]T, 0)
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// Map TODO: Placed here temporarily. Delete after releasing the new osmoutils version.
func Map[E, V any](s []E, f func(E) V) []V {
	res := make([]V, 0, len(s))
	for _, v := range s {
		res = append(res, f(v))
	}
	return res
}
