package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v14/x/poolmanager"
)

func MigrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanagerkeeper.Keeper) {
	migrateNextPoolId(ctx, gammKeeper, poolmanagerKeeper)
}
