package cosmwasmpool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

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

// InitGenesis initializes the cosmwasmpool module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the cosmwasmpool module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

// Set the poolmanager keeper.
func (k *Keeper) SetPoolManagerKeeper(poolmanagerKeeper types.PoolManagerKeeper) {
	k.poolmanagerKeeper = poolmanagerKeeper
}

// Set the contract keeper.
func (k *Keeper) SetContractKeeper(contractKeeper types.ContractKeeper) {
	k.contractKeeper = contractKeeper
}

// Set the wasm keeper.
func (k *Keeper) SetWasmKeeper(wasmKeeper types.WasmKeeper) {
	k.wasmKeeper = wasmKeeper
}

// TODO: godoc and test
func (k *Keeper) convertToCosmwasmPool(poolI poolmanagertypes.PoolI) (types.CosmWasmExtension, error) {
	cosmwasmPool, ok := poolI.(types.CosmWasmExtension)
	if !ok {
		return nil, types.InvalidPoolTypeError{
			ActualPool: poolI,
		}
	}

	cosmwasmPool.SetWasmKeeper(k.wasmKeeper)

	return cosmwasmPool, nil
}
