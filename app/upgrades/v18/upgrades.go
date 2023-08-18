package v18

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v17/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		keepers.PoolManagerKeeper.SetParams(ctx, poolmanagertypes.DefaultParams())

		err = SetAllExistingPoolsTakerFee(ctx, keepers)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

func SetAllExistingPoolsTakerFee(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	lastPoolId := keepers.PoolManagerKeeper.GetNextPoolId(ctx) - 1

	for i := uint64(1); i <= lastPoolId; i++ {
		pool, err := keepers.PoolManagerKeeper.GetPool(ctx, i)
		if err != nil {
			return err
		}

		poolManagerParams := keepers.PoolManagerKeeper.GetParams(ctx)
		accAddress := pool.GetAddress()
		poolBalances := keepers.BankKeeper.GetAllBalances(ctx, accAddress)
		poolType := pool.GetType()

		takerFee := poolmanager.DetermineTakerFee(poolManagerParams, poolBalances, poolType)

		pool.SetTakerFee(takerFee)
		err = keepers.GAMMKeeper.OverwritePoolV15MigrationUnsafe(ctx, pool)
		if err != nil {
			return err
		}
	}
	return nil
}
