package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		storeKey sdk.StoreKey
		hooks    types.EpochHooks
	}
)

// NewKeeper returns a new keeper by codec and storeKey inputs.
func NewKeeper(storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		storeKey: storeKey,
	}
}

// Set the gamm hooks.
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
