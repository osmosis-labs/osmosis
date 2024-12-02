package v23

import (
	"context"
	"errors"
	"sort"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	incentivestypes "github.com/osmosis-labs/osmosis/v28/x/incentives/types"

	"github.com/osmosis-labs/osmosis/v28/app/keepers"
	"github.com/osmosis-labs/osmosis/v28/app/upgrades"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v28/x/concentrated-liquidity"
	concentratedtypes "github.com/osmosis-labs/osmosis/v28/x/concentrated-liquidity/types"
)

const (
	mainnetChainID = "osmosis-1"
	// Edgenet is to function exactly the samas mainnet, and expected
	// to be state-exported from mainnet state.
	edgenetChainID = "edgenet"
	// Testnet will have its own state. Contrary to mainnet, we would
	// like to migrate all testnet pools at once.
	testnetChainID = "osmo-test-5"
	// E2E chain IDs which we expect to migrate all pools similar to testnet.
	e2eChainIDA = "osmo-test-a"
	e2eChainIDB = "osmo-test-b"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		before := time.Now()

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

		chainID := ctx.ChainID()
		// We only perform the migration on mainnet pools since we hard-coded the pool IDs to migrate
		// in the types package. To ensure correctness, we will spin up a state-exported mainnet testnet
		// with the same chain ID.
		if chainID == mainnetChainID || chainID == edgenetChainID {
			if err := migrateMainnetPools(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		} else if chainID == testnetChainID || chainID == e2eChainIDA || chainID == e2eChainIDB {
			if err := migrateAllTestnetPools(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("unsupported chain ID")
		}

		after := time.Now()

		ctx.Logger().Info("migration time", "duration_ms", after.Sub(before).Milliseconds())

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
		if err := concentratedKeeper.MigrateIncentivesAccumulatorToScalingFactor(ctx, poolId); err != nil {
			return err
		}
	}

	return nil
}

// migrates all pools. This is only for testnet.
// CONTRACT: called after setting the pool ID migration threshold since this overwrites the threshold to zero.
func migrateAllTestnetPools(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	// Get all pools
	pools, err := concentratedKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// Migrate each pool
	for _, pool := range pools {
		if err := concentratedKeeper.MigrateIncentivesAccumulatorToScalingFactor(ctx, pool.GetId()); err != nil {
			return err
		}
	}

	// Set to pool ID zero because all pools are migrated.
	concentratedKeeper.SetIncentivePoolIDMigrationThreshold(ctx, 0)

	return nil
}
