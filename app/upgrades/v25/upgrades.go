package v25

import (
	"errors"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	slashing "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/osmosis-labs/osmosis/v24/app/keepers"
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	concentratedliquidity "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity"
	concentratedtypes "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
)

const (
	mainnetChainID = "osmosis-1"
	// Edgenet is to function exactly the same as mainnet, and expected
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
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Snapshot the pool ID migration threshold
		// Get the next pool ID, and set the pool ID migration threshold to the last pool ID.
		nextPoolId := keepers.PoolManagerKeeper.GetNextPoolId(ctx)
		lastPoolID := nextPoolId - 1
		keepers.ConcentratedLiquidityKeeper.SetSpreadFactorPoolIDMigrationThreshold(ctx, lastPoolID)

		// We only perform the migration on mainnet pools since we hard-coded the pool IDs to migrate
		// in the types package. To ensure correctness, we will spin up a state-exported mainnet testnet
		// with the same chain ID.
		chainID := ctx.ChainID()
		if chainID == mainnetChainID || chainID == edgenetChainID {
			if err := migrateMainnetPoolsSpreadFactor(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		} else if chainID == testnetChainID || chainID == e2eChainIDA || chainID == e2eChainIDB {
			if err := migrateAllTestnetPoolsSpreadFactor(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("unsupported chain ID")
		}

		// update block-sdk params
		if err := keepers.AuctionKeeper.SetParams(ctx, AuctionParams); err != nil {
			return nil, err
		}

		// Now that all deprecated historical TWAPs have been pruned via v24, we can delete is isPruning state entry as well
		keepers.TwapKeeper.DeleteDeprecatedHistoricalTWAPsIsPruning(ctx)

		// Reset missed blocks counter for all validators
		resetMissedBlocksCounter(ctx, keepers.SlashingKeeper)

		// Set the authenticator params in the store
		authenticatorParams := keepers.SmartAccountKeeper.GetParams(ctx)
		authenticatorParams.MaximumUnauthenticatedGas = 120_000
		authenticatorParams.IsSmartAccountActive = false
		keepers.SmartAccountKeeper.SetParams(ctx, authenticatorParams)

		return migrations, nil
	}
}

// migrateMainnetPoolsSpreadFactor migrates the specified mainnet pools to the new spread factor accumulator scaling factor.
func migrateMainnetPoolsSpreadFactor(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	poolIDsToMigrate := make([]uint64, 0, len(concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDsV25))
	for poolID := range concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDsV25 {
		poolIDsToMigrate = append(poolIDsToMigrate, poolID)
	}

	// Sort for determinism
	sort.Slice(poolIDsToMigrate, func(i, j int) bool {
		return poolIDsToMigrate[i] < poolIDsToMigrate[j]
	})

	// Migrate concentrated pools
	for _, poolId := range poolIDsToMigrate {
		if err := concentratedKeeper.MigrateSpreadFactorAccumulatorToScalingFactor(ctx, poolId); err != nil {
			return err
		}
	}

	return nil
}

// migrates all pools. This is only for testnet.
// CONTRACT: called after setting the pool ID migration threshold since this overwrites the threshold to zero.
func migrateAllTestnetPoolsSpreadFactor(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	// Get all pools
	pools, err := concentratedKeeper.GetPools(ctx)
	if err != nil {
		return err
	}

	// Migrate each pool
	for _, pool := range pools {
		if err := concentratedKeeper.MigrateSpreadFactorAccumulatorToScalingFactor(ctx, pool.GetId()); err != nil {
			return err
		}
	}

	// Set to pool ID zero because all pools are migrated.
	concentratedKeeper.SetSpreadFactorPoolIDMigrationThreshold(ctx, 0)

	return nil
}

// resetMissedBlocksCounter resets the missed blocks counter for all validators back to zero.
// This corrects a mistake that was overlooked in v24, where we cleared all missedBlocks but did not reset the counter.
func resetMissedBlocksCounter(ctx sdk.Context, slashingKeeper *slashing.Keeper) {
	// Iterate over all validators signing info
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		missedBlocks, err := slashingKeeper.GetValidatorMissedBlocks(ctx, address)
		if err != nil {
			panic(err)
		}

		// Reset missed blocks counter
		info.MissedBlocksCounter = int64(len(missedBlocks))
		slashingKeeper.SetValidatorSigningInfo(ctx, address, info)

		return false
	})
}
