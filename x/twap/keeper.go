package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	transientKey *sdk.TransientStoreKey

	paramSpace paramtypes.Subspace

	ammkeeper types.AmmInterface
}

func NewKeeper(storeKey sdk.StoreKey, transientKey *sdk.TransientStoreKey, paramSpace paramtypes.Subspace, ammKeeper types.AmmInterface) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{storeKey: storeKey, transientKey: transientKey, paramSpace: paramSpace, ammkeeper: ammKeeper}
}

// GetParams returns the total set of twap parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of twap parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k *Keeper) PruneEpochIdentifier(ctx sdk.Context) string {
	return k.GetParams(ctx).PruneEpochIdentifier
}

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, twap := range genState.Twaps {
		k.storeNewRecord(ctx, twap)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	twapRecords, err := k.getAllHistoricalPoolIndexedTWAPs(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Twaps:  twapRecords,
	}
}
