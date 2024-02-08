package v23

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	incentivestypes "github.com/osmosis-labs/osmosis/v23/x/incentives/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/app/upgrades"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity"
	concentratedtypes "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/types"
)

const mainnetChainID = "osmo-test-6"

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

		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyInternalUptime, incentivestypes.DefaultConcentratedUptime)

		// Snapshot the pool ID migration threshold
		// Get the next pool ID
		nextPoolId := keepers.PoolManagerKeeper.GetNextPoolId(ctx)

		lastPoolID := nextPoolId - 1

		keepers.ConcentratedLiquidityKeeper.SetIncentivePoolIDMigrationThreshold(ctx, lastPoolID)

		// We only perform the migration on mainnet pools since we hard-coded the pool IDs to migrate
		// in the types package. To ensure correctness, we will spin up a state-exported mainnet testnet
		// with the same chain ID.
		if ctx.ChainID() == mainnetChainID {
			if err := migrateMainnetPools(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		}

		return migrations, nil
	}
}

// migrateMainnetPools migrates the specified mainnet pools to the new accumulator scaling factor.
func migrateMainnetPools(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	poolIDsToMigrate := make([]uint64, 0, len(concentratedtypes.MigratedIncentiveAccumulatorPoolIDs))
	for poolID := range concentratedtypes.MigratedIncentiveAccumulatorPoolIDs {
		poolIDsToMigrate = append(poolIDsToMigrate, poolID)
	}

	// Sort for determinism
	sort.Slice(poolIDsToMigrate, func(i, j int) bool {
		return poolIDsToMigrate[i] < poolIDsToMigrate[j]
	})

	// Migrate concentrated pools
	for _, poolId := range poolIDsToMigrate {
		if err := concentratedKeeper.MigrateAccumulatorToScalingFactor(ctx, poolId); err != nil {
			return err
		}
	}

	return nil
}
