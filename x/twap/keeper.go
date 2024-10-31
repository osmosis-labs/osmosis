package twap

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v27/x/twap/types"

	storetypes "cosmossdk.io/store/types"
)

type Keeper struct {
	storeKey     storetypes.StoreKey
	transientKey *storetypes.TransientStoreKey

	paramSpace paramtypes.Subspace

	poolmanagerKeeper types.PoolManagerInterface
}

func NewKeeper(storeKey storetypes.StoreKey, transientKey *storetypes.TransientStoreKey, paramSpace paramtypes.Subspace, poolmanagerKeeper types.PoolManagerInterface) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{storeKey: storeKey, transientKey: transientKey, paramSpace: paramSpace, poolmanagerKeeper: poolmanagerKeeper}
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

// SetParam sets a specific twap module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
}

func (k *Keeper) PruneEpochIdentifier(ctx sdk.Context) string {
	return k.GetParams(ctx).PruneEpochIdentifier
}

func (k *Keeper) RecordHistoryKeepPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).RecordHistoryKeepPeriod
}

// InitGenesis initializes the twap module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	// Most recent TWAP must be inserted last. This is required because
	// we maintain a separate index for the most recent records.
	// It is updated by storing new records.
	sort.Slice(genState.Twaps, func(i, j int) bool {
		return genState.Twaps[i].Time.Before(genState.Twaps[j].Time)
	})

	for _, twap := range genState.Twaps {
		k.StoreNewRecord(ctx, twap)
	}
}

// ExportGenesis returns the twap module's exported genesis.
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

// GetGeometricStrategy gets geometric TWAP keeper.
func (k Keeper) GetGeometricStrategy() *geometric {
	return &geometric{k}
}

// GetArithmeticStrategy gets arithmetic TWAP keeper.
func (k Keeper) GetArithmeticStrategy() *arithmetic {
	return &arithmetic{k}
}

// GetPruningState gets the current pruning state, which is used to determine
// whether to prune historical records in the EndBlock. This allows us to spread
// out the computational cost of pruning over time rather than all at once at epoch.
func (k Keeper) GetPruningState(ctx sdk.Context) types.PruningState {
	store := ctx.KVStore(k.storeKey)
	state := types.PruningState{}

	bz := store.Get(types.PruningStateKey)
	if bz == nil {
		return state
	}
	err := proto.Unmarshal(bz, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func (k Keeper) SetPruningState(ctx sdk.Context, state types.PruningState) {
	store := ctx.KVStore(k.storeKey)

	bz, err := proto.Marshal(&state)
	if err != nil {
		panic(err)
	}
	store.Set(types.PruningStateKey, bz)
}
