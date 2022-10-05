package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		storeKey sdk.StoreKey

		paramSpace paramtypes.Subspace

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
