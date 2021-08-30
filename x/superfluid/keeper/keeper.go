package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper provides a way to manage module storage
type Keeper struct {
	cdc        codec.Marshaler
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace

	ak authkeeper.AccountKeeper
	lk types.LockupKeeper
}

// NewKeeper returns an instance of Keeper
func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, ak authkeeper.AccountKeeper, lk types.LockupKeeper) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		ak:         ak,
		lk:         lk,
	}
}

// Logger returns a logger instance
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
