package keeper

import (
	"fmt"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc      codec.Marshaler
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey
		ak       authkeeper.AccountKeeper
		bk       types.BankKeeper
	}
)

func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, ak authkeeper.AccountKeeper, bk types.BankKeeper) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
		bk:       bk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
