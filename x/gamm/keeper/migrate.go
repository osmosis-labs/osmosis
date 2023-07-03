package keeper

import (
	"fmt"
	"sort"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v16/x/gamm/types/migration"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateUnlockedPositionFromBalancerToConcentrated migrates unlocked lp tokens from a balancer pool to a concentrated liquidity pool.
// Fails if the lp tokens are locked (must instead utilize UnlockAndMigrate function in the superfluid module)
func (k Keeper) MigrateUnlockedPositionFromBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress, sharesToMigrate sdk.Coin,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, poolIdLeaving, poolIdEntering uint64, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err = types.GetPoolIdFromShareDenom(sharesToMigrate.Denom)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}

	// Find the governance sanctioned link between the balancer pool and a concentrated pool.
	poolIdEntering, err = k.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}

	// Get the concentrated pool from the message and type cast it to ConcentratedPoolExtension.
	concentratedPool, err := k.concentratedLiquidityKeeper.GetConcentratedPoolById(ctx, poolIdEntering)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}

	// Exit the balancer pool position.
	exitCoins, err := k.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, tokenOutMins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}
	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, fmt.Errorf("Balancer pool must have exactly two tokens")
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, err = k.concentratedLiquidityKeeper.CreateFullRangePosition(ctx, concentratedPool.GetId(), sender, exitCoins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, err
	}
	return positionId, amount0, amount1, liquidity, poolIdLeaving, poolIdEntering, nil
}

// GetAllMigrationInfo gets all existing links between Balancer Pool and Concentrated Pool,
// wraps and returns them in `MigrationRecords`.
func (k Keeper) GetAllMigrationInfo(ctx sdk.Context) (gammmigration.MigrationRecords, error) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixMigrationInfoBalancerPool)

	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	balancerToClPoolLinks := []gammmigration.BalancerToConcentratedPoolLink{}
	for ; iter.Valid(); iter.Next() {
		// balancer Pool Id
		balancerToClPoolLink := gammmigration.BalancerToConcentratedPoolLink{}
		balancerToClPoolLink.BalancerPoolId = sdk.BigEndianToUint64(iter.Key())

		// concentrated Pool Id
		balancerToClPoolLink.ClPoolId = sdk.BigEndianToUint64(iter.Value())

		balancerToClPoolLinks = append(balancerToClPoolLinks, balancerToClPoolLink)
	}

	migrationRecords := gammmigration.MigrationRecords{}
	migrationRecords.BalancerToConcentratedPoolLinks = balancerToClPoolLinks
	return migrationRecords, nil
}

// GetLinkedConcentratedPoolID returns the concentrated pool Id linked for the given balancer pool Id.
// Returns error if link for the given pool id does not exist.
func (k Keeper) GetLinkedConcentratedPoolID(ctx sdk.Context, balancerPoolId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	balancerToClPoolKey := types.GetKeyPrefixMigrationInfoBalancerPool(balancerPoolId)

	concentratedPoolIdBigEndian := store.Get(balancerToClPoolKey)
	if concentratedPoolIdBigEndian == nil {
		return 0, types.ConcentratedPoolMigrationLinkNotFoundError{PoolIdLeaving: balancerPoolId}
	}

	return sdk.BigEndianToUint64(concentratedPoolIdBigEndian), nil
}

// GetLinkedConcentratedPoolID returns the Balancer pool Id linked for the given concentrated pool Id.
// Returns error if link for the given pool id does not exist.
func (k Keeper) GetLinkedBalancerPoolID(ctx sdk.Context, concentratedPoolId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	concentratedToBalancerPoolKey := types.GetKeyPrefixMigrationInfoPoolCLPool(concentratedPoolId)

	balancerPoolIdBigEndian := store.Get(concentratedToBalancerPoolKey)
	if balancerPoolIdBigEndian == nil {
		return 0, types.BalancerPoolMigrationLinkNotFoundError{PoolIdEntering: concentratedPoolId}
	}

	return sdk.BigEndianToUint64(balancerPoolIdBigEndian), nil
}

// OverwriteMigrationRecordsAndRedirectDistrRecords sets the balancer to gamm pool migration info to the store and deletes all existing records
// migrationInfo in state is completely overwitten by the given migrationInfo.
// Additionally, the distribution record for the balancer pool is modified to redirect incentives to the new concentrated pool.
func (k Keeper) OverwriteMigrationRecordsAndRedirectDistrRecords(ctx sdk.Context, migrationInfo gammmigration.MigrationRecords) error {
	store := ctx.KVStore(k.storeKey)

	// delete all existing migration records
	// this is done for both replace and update migration calls because, regardless of whether we are replacing all or updating a few,
	// the resulting migrationInfo that gets passed into this function is the complete set of migration records.
	osmoutils.DeleteAllKeysFromPrefix(ctx, store, types.KeyPrefixMigrationInfoBalancerPool)
	osmoutils.DeleteAllKeysFromPrefix(ctx, store, types.KeyPrefixMigrationInfoCLPool)

	for _, balancerToCLPoolLink := range migrationInfo.BalancerToConcentratedPoolLinks {
		balancerToClPoolKey := types.GetKeyPrefixMigrationInfoBalancerPool(balancerToCLPoolLink.BalancerPoolId)
		store.Set(balancerToClPoolKey, sdk.Uint64ToBigEndian(balancerToCLPoolLink.ClPoolId))

		clToBalancerPoolKey := types.GetKeyPrefixMigrationInfoPoolCLPool(balancerToCLPoolLink.ClPoolId)
		store.Set(clToBalancerPoolKey, sdk.Uint64ToBigEndian(balancerToCLPoolLink.BalancerPoolId))

		err := k.redirectDistributionRecord(ctx, balancerToCLPoolLink.BalancerPoolId, balancerToCLPoolLink.ClPoolId)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetMigrationRecords is used in initGenesis, setting the balancer to gamm pool migration info in store.
func (k Keeper) SetMigrationRecords(ctx sdk.Context, migrationInfo gammmigration.MigrationRecords) {
	store := ctx.KVStore(k.storeKey)

	for _, balancerToCLPoolLink := range migrationInfo.BalancerToConcentratedPoolLinks {
		balancerToClPoolKey := types.GetKeyPrefixMigrationInfoBalancerPool(balancerToCLPoolLink.BalancerPoolId)
		store.Set(balancerToClPoolKey, sdk.Uint64ToBigEndian(balancerToCLPoolLink.ClPoolId))

		clToBalancerPoolKey := types.GetKeyPrefixMigrationInfoPoolCLPool(balancerToCLPoolLink.ClPoolId)
		store.Set(clToBalancerPoolKey, sdk.Uint64ToBigEndian(balancerToCLPoolLink.BalancerPoolId))
	}
}

// redirectDistributionRecord redirects the distribution record for the given balancer pool to the given concentrated pool.
func (k Keeper) redirectDistributionRecord(ctx sdk.Context, cfmmPoolId, clPoolId uint64) error {
	// Get CFMM gauges
	cfmmGauges, err := k.poolIncentivesKeeper.GetGaugesForCFMMPool(ctx, cfmmPoolId)
	if err != nil {
		return err
	}

	if len(cfmmGauges) == 0 {
		return fmt.Errorf("no gauges found for cfmm pool %d", cfmmPoolId)
	}

	// Get longest gauge duration from CFMM pool.
	longestDurationGauge := cfmmGauges[0]
	for i := 1; i < len(cfmmGauges); i++ {
		if cfmmGauges[i].DistributeTo.Duration > longestDurationGauge.DistributeTo.Duration {
			longestDurationGauge = cfmmGauges[i]
		}
	}

	// Get concentrated liquidity gauge duration.
	distributionEpochDuration := k.incentivesKeeper.GetEpochInfo(ctx).Duration

	// Get concentrated gauge corresponding to the distribution epoch duration.
	concentratedGaugeId, err := k.poolIncentivesKeeper.GetPoolGaugeId(ctx, clPoolId, distributionEpochDuration)
	if err != nil {
		return err
	}

	// Iterate through all the distr records, and redirect the old balancer gauge to the new concentrated gauge.
	distrInfo := k.poolIncentivesKeeper.GetDistrInfo(ctx)
	for i, distrRecord := range distrInfo.Records {
		if distrRecord.GaugeId == longestDurationGauge.Id {
			distrInfo.Records[i].GaugeId = concentratedGaugeId
		}
	}

	// Set the new distr info.
	k.poolIncentivesKeeper.SetDistrInfo(ctx, distrInfo)

	return nil
}

// validateRecords validates a list of BalancerToConcentratedPoolLink records to ensure that:
// 1) there are no duplicates
// 2) both the balancer and gamm pool IDs are valid
// 3) the balancer pool has exactly two tokens
// 4) the denoms of the tokens in the balancer pool match the denoms of the tokens in the gamm pool
// It also reorders records from lowest to highest balancer pool ID if they are not provided in order already.
func (k Keeper) validateRecords(ctx sdk.Context, records []gammmigration.BalancerToConcentratedPoolLink) error {
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
			clPool, err = k.concentratedLiquidityKeeper.GetConcentratedPoolById(ctx, record.ClPoolId)
			if err != nil {
				return err
			}
			poolType = clPool.GetType()
			if poolType != poolmanagertypes.Concentrated {
				return fmt.Errorf("Concentrated pool ID #%d is not of type concentrated", record.ClPoolId)
			}

			// Ensure the balancer pools denoms are the same as the concentrated pool denoms
			balancerPoolAssets, err := k.GetTotalPoolLiquidity(ctx, balancerPool.GetId())
			if err != nil {
				return err
			}

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
func (k Keeper) ReplaceMigrationRecords(ctx sdk.Context, records []gammmigration.BalancerToConcentratedPoolLink) error {
	err := k.validateRecords(ctx, records)
	if err != nil {
		return err
	}

	migrationInfo, err := k.GetAllMigrationInfo(ctx)
	if err != nil {
		return err
	}

	migrationInfo.BalancerToConcentratedPoolLinks = records

	// Remove all records from the distribution module and replace them with the new records
	err = k.OverwriteMigrationRecordsAndRedirectDistrRecords(ctx, migrationInfo)
	if err != nil {
		return err
	}
	return nil
}

// UpdateMigrationRecords gets the current migration records and only updates the records that are provided.
// It is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateMigrationRecords(ctx sdk.Context, records []gammmigration.BalancerToConcentratedPoolLink) error {
	err := k.validateRecords(ctx, records)
	if err != nil {
		return err
	}

	recordsMap := make(map[uint64]gammmigration.BalancerToConcentratedPoolLink, len(records))

	// Set up a map of the existing records
	migrationInfos, err := k.GetAllMigrationInfo(ctx)
	if err != nil {
		return err
	}
	for _, existingRecord := range migrationInfos.BalancerToConcentratedPoolLinks {
		recordsMap[existingRecord.BalancerPoolId] = existingRecord
	}

	// Update the map with the new records
	for _, record := range records {
		recordsMap[record.BalancerPoolId] = record
	}

	newRecords := []gammmigration.BalancerToConcentratedPoolLink{}

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

	// We now have a list of all previous records, as well as records that have been updated.
	// We can now remove all previous records and replace them with the new ones.
	err = k.OverwriteMigrationRecordsAndRedirectDistrRecords(ctx, gammmigration.MigrationRecords{
		BalancerToConcentratedPoolLinks: newRecords,
	})
	if err != nil {
		return err
	}
	return nil
}
