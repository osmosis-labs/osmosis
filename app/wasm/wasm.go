package wasm

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

func RegisterCustomPlugins(
	wasmQueryPlugin *QueryPlugin,
) []wasmkeeper.Option {
	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messagePluginOpt := wasmkeeper.WithMessageEncoders(&wasmkeeper.MessageEncoders{
		Custom: CustomEncoder(wasmQueryPlugin),
	})

	return []wasm.Option{
		queryPluginOpt,
		messagePluginOpt,
	}
}
