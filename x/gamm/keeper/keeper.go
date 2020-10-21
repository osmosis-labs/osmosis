package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/keeper/pool"
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

var _ Keeper = (*keeper)(nil)

// type alias
type poolService = pool.Service

type Keeper interface {
	poolService

	types.QueryServer
}

type keeper struct {
	poolService

	// stores
	poolStore pool.Store

	// keepers
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) Keeper {
	var (
		poolStore   = pool.NewStore(cdc, storeKey)
		poolService = pool.NewService(poolStore, accountKeeper, bankKeeper)
	)
	return keeper{
		// pool
		poolService: poolService,
		poolStore:   poolStore,

		// keepers
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
