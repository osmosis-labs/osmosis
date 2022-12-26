package wasmbinding

import (
	"github.com/CosmWasm/wasmd/x/wasm"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

<<<<<<< HEAD
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v13/x/tokenfactory/keeper"
=======
	gammkeeper "github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v13/x/tokenfactory/keeper"
	twap "github.com/osmosis-labs/osmosis/v13/x/twap"
>>>>>>> concentrated-liquidity-main
)

func RegisterCustomPlugins(
	bank *bankkeeper.BaseKeeper,
	tokenFactory *tokenfactorykeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(tokenFactory)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(bank, tokenFactory),
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
