package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	paramSpace paramtypes.Subspace

	// keepers
	poolmanagerKeeper types.PoolManagerKeeper
	bankKeeper        types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper types.BankKeeper, paramSpace paramtypes.Subspace) *Keeper {
	// ParamSubspace must be initialized within app/keepers/keepers.go
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		cdc:        cdc,
		bankKeeper: bankKeeper,
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
