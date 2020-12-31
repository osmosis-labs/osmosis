package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	// keepers
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		// keepers
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}
