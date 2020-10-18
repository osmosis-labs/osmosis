package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/keeper/exchange"
	"github.com/c-osmosis/osmosis/x/gamm/keeper/pool"
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

var _ Keeper = (*keeper)(nil)

type Keeper interface {
	pool.Pool
	exchange.Exchange
}

type keeper struct {
	pool.Pool
	exchange.Exchange

	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewBaseKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) Keeper {
	return keeper{
		Pool:     pool.NewPool(cdc, storeKey, accountKeeper, bankKeeper),
		Exchange: exchange.NewExchange(cdc, storeKey, accountKeeper, bankKeeper),

		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
