package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type ExchangeKeeper interface{}

type exchangeKeeper struct {
	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewExchangeKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) ExchangeKeeper {
	return exchangeKeeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
