package v14

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	routertypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"

	"github.com/osmosis-labs/osmosis/v13/app/keepers"
	"github.com/osmosis-labs/osmosis/v13/app/upgrades"
	gammkeeper "github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		swaprouterParams := swaproutertypes.NewParams(keepers.GAMMKeeper.GetParams(ctx).PoolCreationFee)

		keepers.SwapRouterKeeper.SetParams(ctx, swaprouterParams)

		// N.B: pool id in gamm is to be deprecated in the future
		// Instead,it is moved to swaprouter.
		migrateNextPoolId(ctx, keepers.GAMMKeeper, keepers.SwapRouterKeeper)

		// Router module, set default param
		keepers.RouterKeeper.SetParams(ctx, routertypes.DefaultParams())

		//  N.B.: this is done to avoid initializing genesis for swaprouter module.
		// Otherwise, it would overwrite migrations with InitGenesis().
		// See RunMigrations() for details.
		fromVM[swaproutertypes.ModuleName] = 0

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func migrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, swaprouterKeeper *swaprouter.Keeper) {
	// N.B: pool id in gamm is to be deprecated in the future
	// Instead,it is moved to swaprouter.
	// nolint: staticcheck
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	swaprouterKeeper.SetNextPoolId(ctx, nextPoolId)

	for poolId := uint64(1); poolId < nextPoolId; poolId++ {
		poolType, err := gammKeeper.GetPoolType(ctx, poolId)
		if err != nil {
			panic(err)
		}

		swaprouterKeeper.SetPoolRoute(ctx, poolId, poolType)
	}
}
