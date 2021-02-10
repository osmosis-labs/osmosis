package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper struct
type Keeper struct {
	cdc        codec.Marshaler
	storeKey   sdk.StoreKey
	bankKeeper types.BankKeeper
}

// NewKeeper returns keeper
func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
