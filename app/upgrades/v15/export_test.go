package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	icqkeeper "github.com/strangelove-ventures/async-icq/v4/keeper"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/poolmanager"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v14/x/poolmanager"
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
