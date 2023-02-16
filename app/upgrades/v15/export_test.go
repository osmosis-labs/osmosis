package v15

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	ibcratelimit "github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v14/x/poolmanager"
)

func MigrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanagerkeeper.Keeper) {
	migrateNextPoolId(ctx, gammKeeper, poolmanagerKeeper)
}

func RegisterOsmoIonMetadata(ctx sdk.Context, bankKeeper bankkeeper.Keeper) {
	registerOsmoIonMetadata(ctx, bankKeeper)
}

func SetRateLimits(ctx sdk.Context, accountKeeper *authkeeper.AccountKeeper, rateLimitingICS4Wrapper *ibcratelimit.ICS4Wrapper, wasmKeeper *wasmkeeper.Keeper) {
	setRateLimits(ctx, accountKeeper, rateLimitingICS4Wrapper, wasmKeeper)
}
