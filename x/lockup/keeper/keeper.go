package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v11/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper provides a way to manage module storage.
type Keeper struct {
	cdc      codec.Codec
	storeKey sdk.StoreKey

	hooks types.LockupHooks

	ak types.AccountKeeper
	bk types.BankKeeper
	ck types.CommunityPoolKeeper
}

// NewKeeper returns an instance of Keeper.
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper, ck types.CommunityPoolKeeper) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
		bk:       bk,
		ck:       ck,
	}
}

// Logger returns a logger instance.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Set the lockup hooks.
func (k *Keeper) SetHooks(lh types.LockupHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set lockup hooks twice")
	}

	k.hooks = lh

	return k
}

// AdminKeeper defines a god privilege keeper functions to remove tokens from locks and create new locks
// For the governance system of token pools, we want a "ragequit" feature
// So governance changes will take 1 week to go into effect
// During that time, people can choose to "ragequit" which means they would leave the original pool
// and form a new pool with the old parameters but if they still had 2 months of lockup left,
// their liquidity still needs to be 2 month lockup-ed, just in the new pool
// And we need to replace their pool1 LP tokens with pool2 LP tokens with the same lock duration and end time.
type AdminKeeper struct {
	Keeper
}
