package keeper

import (
	"fmt"
	"sort"
	"time"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// MigrateFromBalancerToConcentrated migrates unlocked lp tokens from a balancer pool to a concentrated liquidity pool.
// Fails if the lp tokens are locked (must utilize UnlockAndMigrate function in the superfluid module)
func (k Keeper) MigrateFromBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, sharesToMigrate sdk.Coin) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, poolIdLeaving, poolIdEntering uint64, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err = types.GetPoolIdFromShareDenom(sharesToMigrate.Denom)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Find the governance sanctioned link between the balancer pool and a concentrated pool.
	poolIdEntering, err = k.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Get the concentrated pool from the message and type cast it to ConcentratedPoolExtension.
	concentratedPool, err := k.concentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, poolIdEntering)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Exit the balancer pool position.
	exitCoins, err := k.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, sdk.NewCoins())
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}
	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, fmt.Errorf("Balancer pool must have exactly two tokens")
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.concentratedLiquidityKeeper.CreateFullRangePosition(ctx, concentratedPool, sender, exitCoins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}
	return positionId, amount0, amount1, liquidity, joinTime, poolIdLeaving, poolIdEntering, nil
}

func (k Keeper) GetAllMigrationInfo(ctx sdk.Context) (types.MigrationRecords, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixMigrationInfoBalancerPool)

	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	balancerToClPoolLinks := []types.BalancerToConcentratedPoolLink{}
	for ; iter.Valid(); iter.Next() {
		// balancer Pool Id
		balancerToClPoolLink := types.BalancerToConcentratedPoolLink{}
		balancerToClPoolLink.BalancerPool.PoolId = sdk.BigEndianToUint64(iter.Key())

		// concentrated Pool Id
		concentratedPoolId := types.PoolID{}
		err := proto.Unmarshal(iter.Value(), &concentratedPoolId)
		if err != nil {
			return types.MigrationRecords{}, err
		}

		balancerToClPoolLink.ClPool.PoolId = concentratedPoolId.PoolId

		balancerToClPoolLinks = append(balancerToClPoolLinks, balancerToClPoolLink)
	}

	migrationRecords := types.MigrationRecords{}
	migrationRecords.BalancerToConcentratedPoolLinks = balancerToClPoolLinks
	return migrationRecords, nil

}

func (k Keeper) GetLinkedConcentratedPoolID(ctx sdk.Context, balancerPoolId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	balancerToClPoolKey := types.GetKeyPrefixMigrationInfoBalancerPool(balancerPoolId)

	concentratedPoolId := types.PoolID{}
	found, err := osmoutils.Get(store, balancerToClPoolKey, &concentratedPoolId)
	if err != nil {
		return 0, err
	}

	if !found {
		return 0, types.ErrPoolNotFound
	}

	return concentratedPoolId.PoolId, nil
}

func (k Keeper) GetLinkedBalancerPoolID(ctx sdk.Context, concentratedPoolId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	concentratedToBalancerPoolKey := types.GetKeyPrefixMigrationInfoBalancerPool(concentratedPoolId)

	balancerPoolId := types.PoolID{}
	found, err := osmoutils.Get(store, concentratedToBalancerPoolKey, &balancerPoolId)
	if err != nil {
		return 0, err
	}

	if !found {
		return 0, types.ErrPoolNotFound
	}

	return balancerPoolId.PoolId, nil
}

// SetMigrationInfo sets the balancer to gamm pool migration info to the store
func (k Keeper) SetMigrationInfo(ctx sdk.Context, migrationInfo types.MigrationRecords) {
	store := ctx.KVStore(k.storeKey)

	for _, balancerToCLPoolLink := range migrationInfo.BalancerToConcentratedPoolLinks {
		balancerToClPoolKey := types.GetKeyPrefixMigrationInfoBalancerPool(balancerToCLPoolLink.BalancerPool.PoolId)
		osmoutils.MustSet(store, balancerToClPoolKey, &balancerToCLPoolLink.ClPool)

		clToBalancerPoolKey := types.GetKeyPrefixMigrationInfoPoolCLPool(balancerToCLPoolLink.ClPool.PoolId)
		osmoutils.MustSet(store, clToBalancerPoolKey, &balancerToCLPoolLink.BalancerPool)
	}
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
		return records[i].BalancerPool.PoolId < records[j].ClPool.PoolId
	})

	for _, record := range records {
		// If the balancer ID has already been seen, we have a duplicate
		if balancerIdFlags[record.BalancerPool.PoolId] {
			return fmt.Errorf(
				"Balancer pool ID #%d has duplications.",
				record.BalancerPool.PoolId,
			)
		}

		// If the concentrated ID has already been seen, we have a duplicate
		if clIdFlags[record.ClPool.PoolId] {
			return fmt.Errorf(
				"Concentrated pool ID #%d has duplications.",
				record.ClPool.PoolId,
			)
		}

		// Ensure records are sorted from lowest to highest balancer pool ID
		if record.BalancerPool.PoolId < lastBalancerPoolID {
			return fmt.Errorf(
				"Balancer pool ID #%d came after Balancer pool ID #%d.",
				record.BalancerPool.PoolId, lastBalancerPoolID,
			)
		}

		// Ensure the provided balancerPoolId exists and that it is of type balancer
		balancerPool, err := k.GetPool(ctx, record.BalancerPool.PoolId)
		if err != nil {
			return err
		}
		poolType := balancerPool.GetType()
		if poolType != poolmanagertypes.Balancer {
			return fmt.Errorf("Balancer pool ID #%d is not of type balancer", record.BalancerPool.PoolId)
		}

		// If clPoolID is 0, this signals a removal, so we skip this check.
		var clPool cltypes.ConcentratedPoolExtension
		if record.ClPool.PoolId != 0 {
			// Ensure the provided ClPoolId exists and that it is of type concentrated.
			clPool, err = k.concentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(ctx, record.ClPool.PoolId)
			if err != nil {
				return err
			}
			poolType = clPool.GetType()
			if poolType != poolmanagertypes.Concentrated {
				return fmt.Errorf("Concentrated pool ID #%d is not of type concentrated", record.ClPool.PoolId)
			}

			// Ensure the balancer pools denoms are the same as the concentrated pool denoms
			balancerPoolAssets, err := k.GetTotalPoolLiquidity(ctx, balancerPool.GetId())
			if err != nil {
				return err
			}

			if len(balancerPoolAssets) != 2 {
				return fmt.Errorf("Balancer pool ID #%d does not contain exactly 2 tokens", record.BalancerPool.PoolId)
			}

			if balancerPoolAssets.AmountOf(clPool.GetToken0()).IsZero() {
				return fmt.Errorf("Balancer pool ID #%d does not contain token %s", record.BalancerPool.PoolId, clPool.GetToken0())
			}
			if balancerPoolAssets.AmountOf(clPool.GetToken1()).IsZero() {
				return fmt.Errorf("Balancer pool ID #%d does not contain token %s", record.BalancerPool.PoolId, clPool.GetToken1())
			}
		}

		lastBalancerPoolID = record.BalancerPool.PoolId

		balancerIdFlags[record.BalancerPool.PoolId] = true
		clIdFlags[record.ClPool.PoolId] = true
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

	migrationInfo, err := k.GetAllMigrationInfo(ctx)
	if err != nil {
		return err
	}

	migrationInfo.BalancerToConcentratedPoolLinks = records

	k.SetMigrationInfo(ctx, migrationInfo)
	return nil
}

// UpdateMigrationRecords gets the current migration records and only updates the records that are provided.
// It is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateMigrationRecords(ctx sdk.Context, records []types.BalancerToConcentratedPoolLink) error {
	err := k.validateRecords(ctx, records)
	if err != nil {
		return err
	}

	recordsMap := make(map[uint64]types.BalancerToConcentratedPoolLink, len(records))

	// Set up a map of the existing records
	migrationInfos, err := k.GetAllMigrationInfo(ctx)
	if err != nil {
		return err
	}
	for _, existingRecord := range migrationInfos.BalancerToConcentratedPoolLinks {
		recordsMap[existingRecord.BalancerPool.PoolId] = existingRecord
	}

	// Update the map with the new records
	for _, record := range records {
		recordsMap[record.BalancerPool.PoolId] = record
	}

	newRecords := []types.BalancerToConcentratedPoolLink{}

	// Iterate through the map and add all the records to a new list
	// if the clPoolId is 0, we remove the entire record
	for _, val := range recordsMap {
		if val.ClPool.PoolId != 0 {
			newRecords = append(newRecords, val)
		}
	}

	// Sort the new records by balancer pool ID
	sort.SliceStable(newRecords, func(i, j int) bool {
		return newRecords[i].BalancerPool.PoolId < newRecords[j].BalancerPool.PoolId
	})

	k.SetMigrationInfo(ctx, types.MigrationRecords{
		BalancerToConcentratedPoolLinks: newRecords,
	})
	return nil
}
