package v24

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/app/upgrades"

	concentratedliquidity "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity"
	concentratedtypes "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/types"
)

const (
	mainnetChainID = "osmosis-1"
	// Edgenet is to function exactly the samas mainnet, and expected
	// to be state-exported from mainnet state.
	edgenetChainID = "edgenet"
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

		// We no longer use the base denoms array and instead use the repeated base denoms field for performance reasons.
		// We retrieve the old base denoms array from the KVStore, delete the array from the KVStore, and set them as a repeated field in the new KVStore.
		baseDenoms, err := keepers.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(ctx)
		if err != nil {
			return nil, err
		}
		keepers.ProtoRevKeeper.DeprecatedDeleteBaseDenoms(ctx)
		err = keepers.ProtoRevKeeper.SetBaseDenoms(ctx, baseDenoms)
		if err != nil {
			return nil, err
		}

		// Now that the TWAP keys are refactored, we can delete all time indexed TWAPs
		// since we only need the pool indexed TWAPs.
		keepers.TwapKeeper.DeleteAllHistoricalTimeIndexedTWAPs(ctx)

		chainID := ctx.ChainID()
		// We only perform the migration on mainnet pools since we hard-coded the pool IDs to migrate
		// in the types package. And the testnet was migrated in v24
		if chainID == mainnetChainID || chainID == edgenetChainID {
			if err := migrateMainnetPools(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		}
		return migrations, nil
	}
}

// migrateMainnetPools migrates the specified mainnet pools to the new accumulator scaling factor.
func migrateMainnetPools(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	poolIDsToMigrate := make([]uint64, 0, len(concentratedtypes.FinalIncentiveAccumulatorPoolIDsToMigrated))
	for poolID := range concentratedtypes.FinalIncentiveAccumulatorPoolIDsToMigrated {
		poolIDsToMigrate = append(poolIDsToMigrate, poolID)
	}

	// Sort for determinism
	sort.Slice(poolIDsToMigrate, func(i, j int) bool {
		return poolIDsToMigrate[i] < poolIDsToMigrate[j]
	})

	// Migrate concentrated pools
	thresholdId, err := concentratedKeeper.GetIncentivePoolIDMigrationThreshold(ctx)
	if err != nil {
		return err
	}

	for _, poolID := range poolIDsToMigrate {
		// This should never happen, this check is defence in depth incase we have wrong data by accident
		if poolID >= thresholdId {
			continue
		}

		// This should never happen, this check is defence in depth incase we have wrong data by accident
		_, isMigrated := concentratedtypes.MigratedIncentiveAccumulatorPoolIDs[poolID]
		if isMigrated {
			continue
		}

		if err := concentratedKeeper.MigrateAccumulatorToScalingFactor(ctx, poolID); err != nil {
			return err
		}
	}

	return nil
}
