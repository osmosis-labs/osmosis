package cosmwasmpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	storetypes "cosmossdk.io/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	// keepers
	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	poolmanagerKeeper types.PoolManagerKeeper
	contractKeeper    types.ContractKeeper
	wasmKeeper        types.WasmKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{cdc: cdc, storeKey: storeKey, paramSpace: paramSpace, accountKeeper: accountKeeper, bankKeeper: bankKeeper}
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

// SetParam sets a specific cosmwasmpool module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
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

// asCosmwasmPool converts a poolI to a CosmWasmExtension.
func (k *Keeper) asCosmwasmPool(poolI poolmanagertypes.PoolI) (types.CosmWasmExtension, error) {
	cosmwasmPool, ok := poolI.(types.CosmWasmExtension)
	if !ok {
		return nil, types.InvalidPoolTypeError{
			ActualPool: poolI,
		}
	}

	cosmwasmPool.SetWasmKeeper(k.wasmKeeper)

	return cosmwasmPool, nil
}

// GetCodeIdByPoolId returns the contract address and code id associated with the given pool.
func (k Keeper) GetCodeIdByPoolId(ctx sdk.Context, poolId uint64) (sdk.AccAddress, uint64, error) {
	pool, err := k.GetPoolById(ctx, poolId)
	if err != nil {
		return nil, 0, err
	}

	contractAddress := sdk.MustAccAddressFromBech32(pool.GetContractAddress())

	contractInfo := k.wasmKeeper.GetContractInfo(ctx, contractAddress)
	if contractInfo == nil {
		return nil, 0, fmt.Errorf("code id for pool id (%d) not found", poolId)
	}
	return contractAddress, contractInfo.CodeID, nil
}
