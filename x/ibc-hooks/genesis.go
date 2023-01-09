package ibc_hooks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
)

var WasmHookModuleAccountAddr sdk.AccAddress = address.Module(ModuleName, []byte("wasm-hook intermediary account"))

func IbcHooksInitGenesis(ctx sdk.Context, ak osmoutils.AccountKeeper) {
	err := osmoutils.CanCreateModuleAccountAtAddr(ctx, ak, WasmHookModuleAccountAddr)
	if err != nil {
		return
	}
	err = osmoutils.CreateModuleAccount(ctx, ak, WasmHookModuleAccountAddr)
	if err != nil {
		panic(err)
	}
}
