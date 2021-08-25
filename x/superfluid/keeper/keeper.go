package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

// Keeper provides a way to manage module storage
type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	ak authkeeper.AccountKeeper
}

// NewKeeper returns an instance of Keeper
func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, ak authkeeper.AccountKeeper) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
	}
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
