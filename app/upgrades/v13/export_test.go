package v13

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v12/x/gamm/keeper"
	swaprouterkeeper "github.com/osmosis-labs/osmosis/v12/x/swaprouter"
)

func MigrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, swaprouterKeeper *swaprouterkeeper.Keeper) {
	migrateNextPoolId(ctx, gammKeeper, swaprouterKeeper)
}
