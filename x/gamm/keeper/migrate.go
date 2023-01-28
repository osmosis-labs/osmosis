package keeper

import (
	"fmt"
	"sort"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cl "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Migrate(ctx sdk.Context, sender sdk.AccAddress, sharesToMigrate sdk.Coin, poolIdEntering uint64) (amount0, amount1 sdk.Int, liquidity sdk.Dec, poolIdLeaving uint64, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err = getPoolIdFromSharesDenom(sharesToMigrate.Denom)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	// Ensure a governance sanctioned link exists between the balancer pool and the concentrated pool.
	migrationInfo := k.GetMigrationInfo(ctx)
	matchFound := false
	for _, info := range migrationInfo.BalancerToConcentratedPoolLinks {
		if info.BalancerPoolId == poolIdLeaving {
			if info.ClPoolId != poolIdEntering {
				return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, types.InvalidPoolMigrationLinkError{PoolIdEntering: poolIdEntering, CanonicalId: info.ClPoolId}
			}
			matchFound = true
			break
		}
	}
	if !matchFound {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, types.PoolMigrationLinkNotFoundError{PoolIdLeaving: poolIdLeaving}
	}

	// Get the concentrated pool from the message and type cast it to ConcentratedPoolExtension.
	poolI, err := k.clKeeper.GetPool(ctx, poolIdEntering)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}
	concentratedPool, ok := poolI.(cltypes.ConcentratedPoolExtension)
	if !ok {
		// If the conversion fails, return an error.
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, fmt.Errorf("given pool does not implement ConcentratedPoolExtension, implements %T", poolI)
	}

	// Exit the balancer pool position.
	exitCoins, err := k.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, sdk.NewCoins())
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	// Determine the max and min ticks for the concentrated pool we are migrating to.
	minTick, maxTick := cl.GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetPrecisionFactorAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	amount0, amount1, liquidity, err = k.clKeeper.CreatePosition(ctx, poolIdEntering, sender, exitCoins.AmountOf(concentratedPool.GetToken0()), exitCoins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}
	return amount0, amount1, liquidity, poolIdLeaving, nil
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
		if poolType.String() != "Balancer" {
			return fmt.Errorf("Balancer pool ID #%d is not of type balancer", record.BalancerPoolId)
		}

		// Ensure the provided ClPoolId exists and that it is of type concentrated.
		// If clPoolID is 0, this signals a removal, so we skip this check.
		var clPool poolmanagertypes.PoolI
		if record.ClPoolId != 0 {
			clPool, err = k.clKeeper.GetPool(ctx, record.ClPoolId)
			if err != nil {
				return err
			}
			poolType = clPool.GetType()
			if poolType.String() != "Concentrated" {
				return fmt.Errorf("Concentrated pool ID #%d is not of type concentrated", record.ClPoolId)
			}
		}

		// Ensure the balancer pools denoms are the same as the concentrated pool denoms
		balancerPoolAssets := balancerPool.GetTotalPoolLiquidity(ctx)

		// Type cast PoolI to ConcentratedPoolExtension
		clPoolExt, ok := clPool.(cltypes.ConcentratedPoolExtension)
		if !ok {
			return fmt.Errorf("pool type (%T) cannot be cast to ConcentratedPoolExtension", clPool)
		}

		if balancerPoolAssets.AmountOf(clPoolExt.GetToken0()).IsZero() {
			return fmt.Errorf("Balancer pool ID #%d does not contain token %s", record.BalancerPoolId, clPoolExt.GetToken0())
		}
		if balancerPoolAssets.AmountOf(clPoolExt.GetToken1()).IsZero() {
			return fmt.Errorf("Balancer pool ID #%d does not contain token %s", record.BalancerPoolId, clPoolExt.GetToken1())
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
