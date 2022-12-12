package downtimedetector

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/downtime-detector/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	storeKey sdk.StoreKey

	paramSpace paramtypes.Subspace
}

func NewKeeper(storeKey sdk.StoreKey, paramSpace paramtypes.Subspace) *Keeper {
	// set KeyTable if it has not already been set
	// if !paramSpace.HasKeyTable() {
	// 	paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	// }

	return &Keeper{storeKey: storeKey, paramSpace: paramSpace}
}

// GetParams returns the total set of twap parameters.
// func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
// 	k.paramSpace.GetParamSet(ctx, &params)
// 	return params
// }

// // SetParams sets the total set of twap parameters.
// func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
// 	k.paramSpace.SetParamSet(ctx, &params)
// }

// InitGenesis initializes the twap module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	// k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the twap module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		// Params: k.GetParams(ctx),
	}
}
