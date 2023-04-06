package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace

	// keepers
	// TODO: remove nolint once added.
	// nolint: unused
	poolmanagerKeeper types.PoolManagerKeeper
	// TODO: remove nolint once added.
	// nolint: unused
	contractKeeper types.ContractKeeper
	// TODO: remove nolint once added.
	// nolint: unused
	wasmKeeper types.WasmKeeper
}

func NewKeeper(storeKey sdk.StoreKey, paramSpace paramtypes.Subspace) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{storeKey: storeKey, paramSpace: paramSpace}
}

// GetParams returns the total set of cosmwasmpool parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of cosmwasmpool parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
