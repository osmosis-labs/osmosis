package twap

import (
	"sort"

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

// InitGenesis initializes the twap module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	// Most recent TWAP must be inserted last. This is required because
	// we maintain a separate index for the most recent TWAP records that
	// is updated by storing new records.
	sort.Slice(genState.Twaps, func(i, j int) bool {
		return genState.Twaps[i].Time.Before(genState.Twaps[j].Time)
	})

	for _, twap := range genState.Twaps {
		k.storeNewRecord(ctx, twap)
	}
}

// ExportGenesis returns the twap module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// These are ordered in increasing order, guaranteed but the iterator
	// that is prefixed by time.
	twapRecords, err := k.getAllHistoricalTimeIndexedTWAPs(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Twaps:  twapRecords,
	}
}
