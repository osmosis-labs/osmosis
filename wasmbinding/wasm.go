package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v13/x/tokenfactory/keeper"
	twap "github.com/osmosis-labs/osmosis/v13/x/twap"
)

func RegisterCustomPlugins(
	gammKeeper *gammkeeper.Keeper,
	bank *bankkeeper.BaseKeeper,
	twap *twap.Keeper,
	tokenFactory *tokenfactorykeeper.Keeper,
	swaprouterKeeper *swaprouter.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(gammKeeper, twap, tokenFactory, swaprouterKeeper)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(gammKeeper, bank, tokenFactory, swaprouterKeeper),
	)

	return []wasm.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}

func RegisterStargateQueries(queryRouter baseapp.GRPCQueryRouter, codec codec.Codec) []wasmkeeper.Option {
	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Stargate: StargateQuerier(queryRouter, codec),
	})

	return []wasm.Option{
		queryPluginOpt,
	}
}
