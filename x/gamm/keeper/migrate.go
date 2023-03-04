package keeper

import (
	"fmt"
	"sort"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateFromBalancerToConcentrated migrates unlocked lp tokens from a balancer pool to a concentrated liquidity pool.
// Fails if the lp tokens are locked (must utilize UnlockAndMigrate function in the superfluid module)
func (k Keeper) MigrateFromBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, sharesToMigrate sdk.Coin) (amount0, amount1 sdk.Int, liquidity sdk.Dec, poolIdLeaving, poolIdEntering uint64, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err = types.GetPoolIdFromShareDenom(sharesToMigrate.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}

	// Find the governance sanctioned link between the balancer pool and a concentrated pool.
	poolIdEntering, err = k.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}

	// Get the concentrated pool from the message and type cast it to ConcentratedPoolExtension.
	concentratedPool, err := k.clKeeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, poolIdEntering)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}

	// Exit the balancer pool position.
	exitCoins, err := k.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, sdk.NewCoins())
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}
	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, fmt.Errorf("Balancer pool must have exactly two tokens")
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	amount0, amount1, liquidity, err = k.clKeeper.CreateFullRangePosition(ctx, concentratedPool, sender, exitCoins, 0)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}
	return amount0, amount1, liquidity, poolIdLeaving, poolIdEntering, nil
}

// GetMigrationInfo returns the balancer to gamm pool migration info from the store
// Returns an empty MigrationRecords struct if migration info does not exist
func (k Keeper) GetMigrationInfo(ctx sdk.Context) types.MigrationRecords {
	store := ctx.KVStore(k.storeKey)
	migrationInfo := types.MigrationRecords{}
	osmoutils.MustGet(store, types.KeyMigrationInfo, &migrationInfo)
	return migrationInfo
}

// SetMigrationInfo sets the balancer to gamm pool migration info to the store
func (k Keeper) SetMigrationInfo(ctx sdk.Context, migrationInfo types.MigrationRecords) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyMigrationInfo, &migrationInfo)
}

// validateRecords validates a list of BalancerToConcentratedPoolLink records to ensure that:
// 1) there are no duplicates
// 2) both the balancer and gamm pool IDs are valid
// 3) the balancer pool has exactly two tokens
// 4) the denoms of the tokens in the balancer pool match the denoms of the tokens in the gamm pool
// It also reorders records from lowest to highest balancer pool ID if they are not provided in order already.
func (k Keeper) validateRecords(ctx sdk.Context, records []types.BalancerToConcentratedPoolLink) error {
	lastBalancerPoolID := uint64(0)
	balancerIdFlags := make(map[uint64]bool, len(records))
	clIdFlags := make(map[uint64]bool, len(records))

	// Sort the provided records by balancer pool ID
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].BalancerPoolId < records[j].BalancerPoolId
	})

	for _, record := range records {
		// If the balancer ID has already been seen, we have a duplicate
		if balancerIdFlags[record.BalancerPoolId] {
			return fmt.Errorf(
				"Balancer pool ID #%d has duplications.",
				record.BalancerPoolId,
			)
		}

		// If the concentrated ID has already been seen, we have a duplicate
		if clIdFlags[record.ClPoolId] {
			return fmt.Errorf(
				"Concentrated pool ID #%d has duplications.",
				record.ClPoolId,
			)
		}

		// Ensure records are sorted from lowest to highest balancer pool ID
		if record.BalancerPoolId < lastBalancerPoolID {
			return fmt.Errorf(
				"Balancer pool ID #%d came after Balancer pool ID #%d.",
				record.BalancerPoolId, lastBalancerPoolID,
			)
		}

		// Ensure the provided balancerPoolId exists and that it is of type balancer
		balancerPool, err := k.GetPool(ctx, record.BalancerPoolId)
		if err != nil {
			return err
		}
		poolType := balancerPool.GetType()
		if poolType != poolmanagertypes.Balancer {
			return fmt.Errorf("Balancer pool ID #%d is not of type balancer", record.BalancerPoolId)
		}

		// If clPoolID is 0, this signals a removal, so we skip this check.
		var clPool cltypes.ConcentratedPoolExtension
		if record.ClPoolId != 0 {
			// Ensure the provided ClPoolId exists and that it is of type concentrated.
			clPool, err = k.clKeeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, record.ClPoolId)
			if err != nil {
				return err
			}
			poolType = clPool.GetType()
			if poolType != poolmanagertypes.Concentrated {
				return fmt.Errorf("Concentrated pool ID #%d is not of type concentrated", record.ClPoolId)
			}

			// Ensure the balancer pools denoms are the same as the concentrated pool denoms
			balancerPoolAssets := balancerPool.GetTotalPoolLiquidity(ctx)

			if len(balancerPoolAssets) != 2 {
				return fmt.Errorf("Balancer pool ID #%d does not contain exactly 2 tokens", record.BalancerPoolId)
			}

			if balancerPoolAssets.AmountOf(clPool.GetToken0()).IsZero() {
				return fmt.Errorf("Balancer pool ID #%d does not contain token %s", record.BalancerPoolId, clPool.GetToken0())
			}
			if balancerPoolAssets.AmountOf(clPool.GetToken1()).IsZero() {
				return fmt.Errorf("Balancer pool ID #%d does not contain token %s", record.BalancerPoolId, clPool.GetToken1())
			}
		}

		lastBalancerPoolID = record.BalancerPoolId

		balancerIdFlags[record.BalancerPoolId] = true
		clIdFlags[record.ClPoolId] = true
	}
	return nil
}

// ReplaceMigrationRecords gets the current migration records and replaces it in its entirety with the provided records.
// It is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) ReplaceMigrationRecords(ctx sdk.Context, records []types.BalancerToConcentratedPoolLink) error {
	err := k.validateRecords(ctx, records)
	if err != nil {
		return err
	}

	migrationInfo := k.GetMigrationInfo(ctx)

	migrationInfo.BalancerToConcentratedPoolLinks = records

	k.SetMigrationInfo(ctx, migrationInfo)
	return nil
}

// UpdateDistrRecords gets the current migration records and only updates the records that are provided.
// It is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateMigrationRecords(ctx sdk.Context, records []types.BalancerToConcentratedPoolLink) error {
	err := k.validateRecords(ctx, records)
	if err != nil {
		return err
	}

	recordsMap := make(map[uint64]types.BalancerToConcentratedPoolLink, len(records))

	// Set up a map of the existing records
	for _, existingRecord := range k.GetMigrationInfo(ctx).BalancerToConcentratedPoolLinks {
		recordsMap[existingRecord.BalancerPoolId] = existingRecord
	}

	// Update the map with the new records
	for _, record := range records {
		recordsMap[record.BalancerPoolId] = record
	}

	newRecords := []types.BalancerToConcentratedPoolLink{}

	// Iterate through the map and add all the records to a new list
	// if the clPoolId is 0, we remove the entire record
	for _, val := range recordsMap {
		if val.ClPoolId != 0 {
			newRecords = append(newRecords, val)
		}
	}

	// Sort the new records by balancer pool ID
	sort.SliceStable(newRecords, func(i, j int) bool {
		return newRecords[i].BalancerPoolId < newRecords[j].BalancerPoolId
	})

	k.SetMigrationInfo(ctx, types.MigrationRecords{
		BalancerToConcentratedPoolLinks: newRecords,
	})
	return nil
}

// GetLinkedConcentratedPoolID checks if a governance sanctioned link exists between the provided balancer pool and a concentrated pool.
// If a link exists, it returns the concentrated pool ID.
// If a link does not exist, it returns a 0 pool ID an error.
func (k Keeper) GetLinkedConcentratedPoolID(ctx sdk.Context, poolIdLeaving uint64) (poolIdEntering uint64, err error) {
	migrationInfo := k.GetMigrationInfo(ctx)
	for _, info := range migrationInfo.BalancerToConcentratedPoolLinks {
		if info.BalancerPoolId == poolIdLeaving {
			return info.ClPoolId, nil
		}
	}
	return 0, types.PoolMigrationLinkNotFoundError{PoolIdLeaving: poolIdLeaving}
}
