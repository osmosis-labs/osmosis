package v24

import (
	"errors"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"

	cwpooltypes "github.com/osmosis-labs/osmosis/v24/x/cosmwasmpool/types"

	"github.com/osmosis-labs/osmosis/v24/app/keepers"
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"

	concentratedliquidity "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity"
	concentratedtypes "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v24/x/incentives/types"
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
		// since we only need the pool indexed TWAPs. We set the is pruning store value to true
		// and spread the pruning time across multiple blocks to avoid a single block taking too long.
		keepers.TwapKeeper.SetDeprecatedHistoricalTWAPsIsPruning(ctx)

		// Set the new min value for distribution for the incentives module.
		// https://www.mintscan.io/osmosis/proposals/733
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyMinValueForDistr, incentivestypes.DefaultMinValueForDistr)

		// Enable ICA controllers
		keepers.ICAControllerKeeper.SetParams(ctx, icacontrollertypes.DefaultParams())

		// White Whale uploaded a broken contract. They later migrated cwpool via the governance
		// proposal in x/cosmwasmpool
		// However, there was a problem in the migration logic where the CosmWasmpool state CodeId  did not get updated.
		// As a result, the CodeID for the contract that is tracked in x/wasmd  was migrated correctly. However, the code ID that we track in the x/cosmwasmpool  state did not.
		// Therefore, we should perform a migration for each of the hardcoded white whale pools.
		poolIds := []uint64{1463, 1462, 1461}
		for _, poolId := range poolIds {
			pool, err := keepers.CosmwasmPoolKeeper.GetPool(ctx, poolId)
			if err != nil {
				// Skip non-existent pools. This way we don't need to create the pools on E2E tests
				continue
			}
			cwPool, ok := pool.(cwpooltypes.CosmWasmExtension)
			if !ok {
				ctx.Logger().Error("Pool has incorrect type", "poolId", poolId, "pool", pool)
				return nil, cwpooltypes.InvalidPoolTypeError{
					ActualPool: pool,
				}
			}
			if cwPool.GetCodeId() != 503 {
				ctx.Logger().Error("Pool has incorrect code id", "poolId", poolId, "codeId", cwPool.GetCodeId())
				return nil, cwpooltypes.InvalidPoolTypeError{
					ActualPool: pool,
				}
			}
			cwPool.SetCodeId(572)
			keepers.CosmwasmPoolKeeper.SetPool(ctx, cwPool)
		}

		// Snapshot the pool ID migration threshold
		// Get the next pool ID, and set the pool ID migration threshold to the last pool ID.
		nextPoolId := keepers.PoolManagerKeeper.GetNextPoolId(ctx)
		lastPoolID := nextPoolId - 1
		keepers.ConcentratedLiquidityKeeper.SetSpreadFactorPoolIDMigrationThreshold(ctx, lastPoolID)

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

		return migrations, nil
	}
}

// migrateMainnetPools migrates the specified mainnet pools to the new spread factor accumulator scaling factor.
func migrateMainnetPools(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	poolIDsToMigrate := make([]uint64, 0, len(concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDs))
	for poolID := range concentratedtypes.MigratedSpreadFactorAccumulatorPoolIDs {
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
func migrateAllTestnetPools(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
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
