package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	// paramSpace stores module's params
	paramSpace paramtypes.Subspace
	// router is used to access tokenfactory methods
	router *baseapp.MsgServiceRouter
	// accountKeeper helps get the module's address
	accountKeeper types.AccountKeeper
	// authority is used in UpdateParams method. It is
	// the only addr that can update bridge module params
	authority string
}

// NewKeeper returns a new instance of the x/bridge keeper.
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	paramSpace paramtypes.Subspace,
	router *baseapp.MsgServiceRouter,
	accountKeeper types.AccountKeeper,
	authority string,
) Keeper {
	// ensure bridge module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the bridge module account has not been set")
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid bridge authority address: %w", err))
	}

	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramSpace:    paramSpace,
		router:        router,
		accountKeeper: accountKeeper,
		authority:     authority,
	}
}

// Logger returns a logger for the x/bridge module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
