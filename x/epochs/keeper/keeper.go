package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc      codec.Codec
		storeKey sdk.StoreKey
		hooks    types.EpochHooks
	}
)

// NewKeeper returns a new keeper by codec and storeKey inputs.
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
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
