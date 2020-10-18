package exchange

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Exchange interface {
	Sender
	Viewer
}

type exchange struct {
	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewExchange(
	cdc codec.BinaryMarshaler,
	storeKey sdk.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
) Exchange {
	return exchange{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
