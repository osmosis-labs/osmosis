package v14

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammkeeper "github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"
	swaprouterkeeper "github.com/osmosis-labs/osmosis/v13/x/swaprouter"
)

func MigrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, swaprouterKeeper *swaprouterkeeper.Keeper) {
	migrateNextPoolId(ctx, gammKeeper, swaprouterKeeper)
}
