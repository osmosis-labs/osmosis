package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type PoolKeeper interface{}

type poolKeeper struct {
	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewPoolKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) PoolKeeper {
	return poolKeeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
