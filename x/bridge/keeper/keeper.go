package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type Keeper struct {
	// paramSpace stores module's params
	paramSpace paramtypes.Subspace
	// router is used to access tokenfactory methods
	router *baseapp.MsgServiceRouter
	// accountKeeper helps get the module's address
	accountKeeper types.AccountKeeper
	// govModuleAddr is used in UpdateParams method since it is
	// the only addr that can update bridge module params
	govModuleAddr string
}

// NewKeeper returns a new instance of the x/bridge keeper.
func NewKeeper(
	paramSpace paramtypes.Subspace,
	router *baseapp.MsgServiceRouter,
	accountKeeper types.AccountKeeper,
	govModuleAddr string,
) Keeper {
	// ensure bridge module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the bridge module account has not been set")
	}

	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		paramSpace:    paramSpace,
		router:        router,
		accountKeeper: accountKeeper,
		govModuleAddr: govModuleAddr,
	}
}

// Logger returns a logger for the x/bridge module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
