package concentrated_liquidity

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
)

// MigrateMainnetPools migrates the specified mainnet pools to the new accumulator scaling factor.
func MigrateMainnetPools(ctx sdk.Context, concentratedKeeper Keeper) error {
	// Define params to ensure the function runs as expected
	numberOfPoolsToMigratePerBlock := 2
	finalPoolID := uint64(1066)

	// Migrate concentrated pools, thresholdID starts at 1496
	thresholdID, err := concentratedKeeper.GetIncentivePoolIDMigrationThreshold(ctx)
	if err != nil {
		return err
	}

	// If the thresholdID is less than pool 1066, we know all pools have been migrated
	if thresholdID < finalPoolID {
		return nil
	}
	ctx.Logger().Info("Lazy migration started at", "pool_id", thresholdID)

	poolIDsToMigrate := make([]uint64, 0, len(types.FinalIncentiveAccumulatorPoolIDsToMigrate))
	// types.FinalIncentiveAccumulatorPoolIDsToMigrate is a map
	for poolID := range types.FinalIncentiveAccumulatorPoolIDsToMigrate {
		// Only include pools that haven't been migrated
		if poolID <= thresholdID {
			poolIDsToMigrate = append(poolIDsToMigrate, poolID)
		}
	}

	// Sort for determinism, starting with the highest pool
	sort.Slice(poolIDsToMigrate, func(i, j int) bool {
		return poolIDsToMigrate[i] > poolIDsToMigrate[j]
	})

	// Iterate over the sorted pools to ensure that pool migration is deterministic
	sortedPoolIDsToMigrate := make([]uint64, 0, numberOfPoolsToMigratePerBlock)
	for _, poolID := range poolIDsToMigrate {
		sortedPoolIDsToMigrate = append(sortedPoolIDsToMigrate, poolID)

		// Only iterate over pools that we need to migrate
		if len(sortedPoolIDsToMigrate) == numberOfPoolsToMigratePerBlock {
			break
		}
	}

	for i, poolID := range sortedPoolIDsToMigrate {
		// This should never happen, this check is defence in depth in case we have wrong data by accident
		if poolID > thresholdID {
			ctx.Logger().Info("Skipping cl incentive migration for", "pool_id", poolID)
			continue
		}

		// This should never happen, this check is defence in depth in case we have wrong data by accident
		_, isMigrated := types.MigratedIncentiveAccumulatorPoolIDs[poolID]
		if isMigrated {
			continue
		}

		// Run the migration on the selected pool
		if err := concentratedKeeper.MigrateAccumulatorToScalingFactor(ctx, poolID); err != nil {
			return err
		}

		// Only migrate 2 pools at a time or finish lazy migration if poolID is finalPoolID
		if i >= numberOfPoolsToMigratePerBlock-1 || poolID == finalPoolID {
			ctx.Logger().Info("Lazy migration stopped at", "pool_id", poolID)

			// Set the incentive pool id to be 1 below the latest migration
			concentratedKeeper.SetIncentivePoolIDMigrationThreshold(ctx, poolID-1)
			break
		}
	}

	return nil
}
