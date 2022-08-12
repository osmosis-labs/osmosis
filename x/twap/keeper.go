package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v10/x/twap/types"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	transientKey *sdk.TransientStoreKey

	paramSpace paramtypes.Subspace

	ammkeeper types.AmmInterface
}

func NewKeeper(storeKey sdk.StoreKey, transientKey *sdk.TransientStoreKey, paramSpace paramtypes.Subspace, ammKeeper types.AmmInterface) *Keeper {
	return &Keeper{storeKey: storeKey, transientKey: transientKey, paramSpace: paramSpace, ammkeeper: ammKeeper}
}

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k *Keeper) PruneEpochIdentifier(ctx sdk.Context) string {
	return k.GetParams(ctx).PruneEpochIdentifier
}
