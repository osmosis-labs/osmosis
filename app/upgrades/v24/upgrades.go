package v24

import (
	"context"
	"sort"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"

	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	concentratedliquidity "github.com/osmosis-labs/osmosis/v30/x/concentrated-liquidity"
	concentratedtypes "github.com/osmosis-labs/osmosis/v30/x/concentrated-liquidity/types"
	cwpooltypes "github.com/osmosis-labs/osmosis/v30/x/cosmwasmpool/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v30/x/incentives/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v30/x/txfees/types"
)

const (
	mainnetChainID = "osmosis-1"
	// Edgenet is to function exactly the same as mainnet, and expected
	// to be state-exported from mainnet state.
	edgenetChainID = "edgenet"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
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
		// keepers.TwapKeeper.SetDeprecatedHistoricalTWAPsIsPruning(ctx)

		// Set the new min value for distribution for the incentives module.
		// https://www.mintscan.io/osmosis/proposals/733
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyMinValueForDistr, incentivestypes.DefaultMinValueForDistr)

		chainID := ctx.ChainID()
		// We only perform the migration on mainnet pools since we hard-coded the pool IDs to migrate
		// in the types package. And the testnet was migrated in v24
		if chainID == mainnetChainID || chainID == edgenetChainID {
			if err := migrateMainnetPools(ctx, *keepers.ConcentratedLiquidityKeeper); err != nil {
				return nil, err
			}
		}
		// Enable ICA controllers
		keepers.ICAControllerKeeper.SetParams(ctx, icacontrollertypes.DefaultParams())

		// White Whale uploaded a broken contract. They later migrated cwpool via the governance
		// proposal in x/cosmwasmpool
		// However, there was a problem in the migration logic where the CosmWasmpool state CodeId did not get updated.
		// As a result, the CodeID for the contract that is tracked in x/wasmd was migrated correctly. However, the code ID that we track in the x/cosmwasmpool state did not.
		// Therefore, we should perform a migration for each of the hardcoded white whale pools.
		poolIds := []uint64{1584, 1575, 1514, 1463, 1462, 1461}
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
			if cwPool.GetCodeId() != 503 && cwPool.GetCodeId() != 572 {
				ctx.Logger().Error("Pool has incorrect code id", "poolId", poolId, "codeId", cwPool.GetCodeId())
				return nil, cwpooltypes.InvalidPoolTypeError{
					ActualPool: pool,
				}
			}
			cwPool.SetCodeId(641)
			keepers.CosmwasmPoolKeeper.SetPool(ctx, cwPool)
		}

		// Set whitelistedFeeTokenSetters param as per https://forum.osmosis.zone/t/temperature-check-add-a-permissioned-address-to-manage-the-fee-token-whitelist/2604
		keepers.TxFeesKeeper.SetParam(ctx, txfeestypes.KeyWhitelistedFeeTokenSetters, WhitelistedFeeTokenSetters)
		return migrations, nil
	}
}

// migrateMainnetPools migrates the specified mainnet pools to the new accumulator scaling factor.
func migrateMainnetPools(ctx sdk.Context, concentratedKeeper concentratedliquidity.Keeper) error {
	poolIDsToMigrate := make([]uint64, 0, len(concentratedtypes.MigratedIncentiveAccumulatorPoolIDsV24))
	for poolID := range concentratedtypes.MigratedIncentiveAccumulatorPoolIDsV24 {
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
		// This should never happen, this check is defence in depth in case we have wrong data by accident
		if poolID >= thresholdId {
			continue
		}

		// This should never happen, this check is defence in depth in case we have wrong data by accident
		_, isMigrated := concentratedtypes.MigratedIncentiveAccumulatorPoolIDs[poolID]
		if isMigrated {
			continue
		}

		if err := concentratedKeeper.MigrateIncentivesAccumulatorToScalingFactor(ctx, poolID); err != nil {
			return err
		}
	}

	return nil
}
