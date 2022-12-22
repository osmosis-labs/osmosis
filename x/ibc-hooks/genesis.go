package ibc_hooks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v13/osmoutils"
=======
	"github.com/osmosis-labs/osmosis/v13/x/ibc-hooks/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
>>>>>>> 93181a45 (Move osmoutils into its own go mod (#3827))
)

var WasmHookModuleAccountAddr sdk.AccAddress = address.Module(ModuleName, []byte("wasm-hook intermediary account"))

func IbcHooksInitGenesis(ctx sdk.Context, ak osmoutils.AccountKeeper) {
	err := osmoutils.CreateModuleAccount(ctx, ak, WasmHookModuleAccountAddr)
	if err != nil {
		panic(err)
	}
}
