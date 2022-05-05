package wasm

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	lockupkeeper "github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
)

func RegisterCustomPlugins(
	gammKeeper *gammkeeper.Keeper,
	bank *bankkeeper.BaseKeeper,
	lockup *lockupkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(gammKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(gammKeeper, bank, lockup),
	)

	return []wasm.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
