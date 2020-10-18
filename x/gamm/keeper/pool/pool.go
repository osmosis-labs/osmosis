package pool

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Pool interface {
	Sender
	Viewer
}

type pool struct {
	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewPool(
	cdc codec.BinaryMarshaler,
	storeKey sdk.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
) Pool {
	return pool{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
