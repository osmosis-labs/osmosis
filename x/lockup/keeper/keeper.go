package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper provides a way to manage module storage
type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey
}

// NewKeeper returns an instance of Keeper
func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
