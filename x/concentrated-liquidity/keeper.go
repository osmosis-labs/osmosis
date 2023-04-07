package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	paramSpace paramtypes.Subspace
	listeners  types.ConcentratedLiquidityListeners

	// keepers
	poolmanagerKeeper types.PoolManagerKeeper
	bankKeeper        types.BankKeeper
	lockupKeeper      types.LockupKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper types.BankKeeper, lockupKeeper types.LockupKeeper, paramSpace paramtypes.Subspace) *Keeper {
	// ParamSubspace must be initialized within app/keepers/keepers.go
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		storeKey:     storeKey,
		paramSpace:   paramSpace,
		cdc:          cdc,
		bankKeeper:   bankKeeper,
		lockupKeeper: lockupKeeper,
	}
}

// GetParams returns the total set of concentrated-liquidity module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the concentrated-liquidity module's parameters with the provided parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// Set the poolmanager keeper.
func (k *Keeper) SetPoolManagerKeeper(poolmanagerKeeper types.PoolManagerKeeper) {
	k.poolmanagerKeeper = poolmanagerKeeper
}

// GetNextPositionId returns the next position id.
func (k Keeper) GetNextPositionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPositionId := gogotypes.UInt64Value{}
	osmoutils.MustGet(store, types.KeyNextGlobalPositionId, &nextPositionId)
	return nextPositionId.Value
}

// SetNextPositionId sets next position Id.
func (k Keeper) SetNextPositionId(ctx sdk.Context, positionId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextGlobalPositionId, &gogotypes.UInt64Value{Value: positionId})
}

// Set the concentrated-liquidity listeners.
func (k *Keeper) SetListeners(listeners types.ConcentratedLiquidityListeners) *Keeper {
	if k.listeners != nil {
		panic("cannot set concentrated liquidity listeners twice")
	}

	k.listeners = listeners

	return k
}
