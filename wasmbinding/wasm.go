package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v10/x/gamm/keeper"
	twap "github.com/osmosis-labs/osmosis/v10/x/gamm/twap"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/keeper"
)

func RegisterCustomPlugins(
	gammKeeper *gammkeeper.Keeper,
	bank *bankkeeper.BaseKeeper,
	twap *twap.Keeper,
	tokenFactory *tokenfactorykeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(gammKeeper, twap, tokenFactory)

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
