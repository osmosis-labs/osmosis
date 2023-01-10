package v14

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v14/app/keepers"
	"github.com/osmosis-labs/osmosis/v14/app/upgrades"
	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		poolmanagerParams := poolmanagertypes.NewParams(keepers.GAMMKeeper.GetParams(ctx).PoolCreationFee)

		keepers.PoolManagerKeeper.SetParams(ctx, poolmanagerParams)

		// N.B: pool id in gamm is to be deprecated in the future
		// Instead,it is moved to poolmanager.
		migrateNextPoolId(ctx, keepers.GAMMKeeper, keepers.PoolManagerKeeper)

		//  N.B.: this is done to avoid initializing genesis for poolmanager module.
		// Otherwise, it would overwrite migrations with InitGenesis().
		// See RunMigrations() for details.
		fromVM[poolmanagertypes.ModuleName] = 0

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func migrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper) {
	// N.B: pool id in gamm is to be deprecated in the future
	// Instead,it is moved to poolmanager.
	// nolint: staticcheck
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	poolmanagerKeeper.SetNextPoolId(ctx, nextPoolId)

	for poolId := uint64(1); poolId < nextPoolId; poolId++ {
		poolType, err := gammKeeper.GetPoolType(ctx, poolId)
		if err != nil {
			panic(err)
		}

		poolmanagerKeeper.SetPoolRoute(ctx, poolId, poolType)
	}
}
