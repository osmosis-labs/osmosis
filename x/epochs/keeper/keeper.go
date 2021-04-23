package keeper

import (
	"fmt"

	"github.com/c-osmosis/osmosis/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc      codec.Marshaler
		storeKey sdk.StoreKey
		hooks    types.EpochHooks
	}
)

func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, hooks types.EpochHooks) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		hooks:    hooks,
	}
}

// Set the gamm hooks
func (k *Keeper) SetHooks(eh types.EpochHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set epochs hooks twice")
	}

	k.hooks = eh

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
