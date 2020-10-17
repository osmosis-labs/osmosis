package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

var _ Keeper = (*keeper)(nil)

type Keeper interface {
	PoolKeeper
	ExchangeKeeper
}

type keeper struct {
	PoolKeeper
	ExchangeKeeper

	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewBaseKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) Keeper {
	return keeper{
		PoolKeeper:     NewPoolKeeper(cdc, storeKey, accountKeeper, bankKeeper),
		ExchangeKeeper: NewExchangeKeeper(cdc, storeKey, accountKeeper, bankKeeper),

		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
