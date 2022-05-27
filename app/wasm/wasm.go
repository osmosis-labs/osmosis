package wasm

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v8/x/tokenfactory/keeper"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v8/x/gamm/keeper"
)

func RegisterCustomPlugins(
	gammKeeper *gammkeeper.Keeper,
	bank *bankkeeper.BaseKeeper,
	tokenFactory *tokenfactorykeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(gammKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(gammKeeper, bank, tokenFactory),
	)

	return []wasm.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
