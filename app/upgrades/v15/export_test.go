package v15

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	icqkeeper "github.com/strangelove-ventures/async-icq/v4/keeper"

	ibcratelimit "github.com/osmosis-labs/osmosis/v15/x/ibc-rate-limit"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v15/x/poolmanager"
)

func MigrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanagerkeeper.Keeper) {
	migrateNextPoolId(ctx, gammKeeper, poolmanagerKeeper)
}

func RegisterOsmoIonMetadata(ctx sdk.Context, bankKeeper bankkeeper.Keeper) {
	registerOsmoIonMetadata(ctx, bankKeeper)
}

func SetICQParams(ctx sdk.Context, icqKeeper *icqkeeper.Keeper) {
	setICQParams(ctx, icqKeeper)
}

func MigrateBalancerPoolToSolidlyStable(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper, bankKeeper bankkeeper.Keeper, poolId uint64) {
	migrateBalancerPoolToSolidlyStable(ctx, gammKeeper, poolmanagerKeeper, bankKeeper, poolId)
}

func SetRateLimits(ctx sdk.Context, accountKeeper *authkeeper.AccountKeeper, rateLimitingICS4Wrapper *ibcratelimit.ICS4Wrapper, wasmKeeper *wasmkeeper.Keeper) {
	setRateLimits(ctx, accountKeeper, rateLimitingICS4Wrapper, wasmKeeper)
}
