package v13

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v12/app/keepers"
	"github.com/osmosis-labs/osmosis/v12/app/upgrades"
	gammkeeper "github.com/osmosis-labs/osmosis/v12/x/gamm/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v12/x/lockup/types"
	swaprouterkeeper "github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		keepers.LockupKeeper.SetParams(ctx, lockuptypes.DefaultParams())
		keepers.SwapRouterKeeper.SetParams(ctx, swaproutertypes.DefaultParams())

		// N.B: pool id in gamm is to be deprecated in the future
		// Instead,it is moved to swaprouter.
		migrateNextPoolId(ctx, keepers.GAMMKeeper, keepers.SwapRouterKeeper)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func migrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, swaprouterKeeper *swaprouterkeeper.Keeper) {
	// N.B: pool id in gamm is to be deprecated in the future
	// Instead,it is moved to swaprouter.
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	gammKeeper.SetPoolCount(ctx, nextPoolId-1)
	swaprouterKeeper.SetNextPoolId(ctx, nextPoolId)
}
