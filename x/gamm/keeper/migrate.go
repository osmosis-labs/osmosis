package keeper

import (
	"fmt"
	"sort"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetMigrationInfo(ctx sdk.Context) types.MigrationRecords {
	store := ctx.KVStore(k.storeKey)
	migrationInfo := types.MigrationRecords{}
	osmoutils.MustGet(store, types.KeyMigrationInfo, &migrationInfo)
	return migrationInfo
}

func (k Keeper) SetMigrationInfo(ctx sdk.Context, migrationInfo types.MigrationRecords) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyMigrationInfo, &migrationInfo)
}

// validateRecords validates a list of records to ensure that:
// 1) there are no duplicates,
// 2) the records are in sorted order.
// 3) the pool IDs are valid
func (k Keeper) validateRecords(ctx sdk.Context, records ...types.BalancerToConcentratedPoolLink) error {
	lastBalancerPoolID := uint64(0)
	balancerIdFlags := make(map[uint64]bool)
	clIdFlags := make(map[uint64]bool)

	for _, record := range records {
		if balancerIdFlags[record.BalancerPoolId] {
			return fmt.Errorf(
				"Balancer pool ID #%d has duplications.",
				record.BalancerPoolId,
			)
		}

		if clIdFlags[record.ClPoolId] {
			return fmt.Errorf(
				"Concentrated pool ID #%d has duplications.",
				record.ClPoolId,
			)
		}

		// Ensure records are sorted
		if record.BalancerPoolId < lastBalancerPoolID {
			return fmt.Errorf(
				"Balancer pool ID #%d came after Gauge ID #%d.",
				record.BalancerPoolId, lastBalancerPoolID,
			)
		}

		// Ensure the first pool exists and that it is of type balancer
		poolType, err := k.GetPoolType(ctx, record.BalancerPoolId)
		if err != nil {
			return err
		}
		if poolType.String() != "Balancer" {
			return fmt.Errorf("Balancer pool ID #%d is not of type balancer", record.BalancerPoolId)
		}

		// Ensure the concentrated pool exists. If record is 0, its a removal, so we skip this check.
		// TODO: Get GetPoolType to work for cl pools from gamm
		if record.ClPoolId != 0 {
			_, err = k.poolManager.GetPoolModule(ctx, record.ClPoolId)
			if err != nil {
				return err
			}
		}

		lastBalancerPoolID = record.BalancerPoolId

		balancerIdFlags[record.BalancerPoolId] = true
		clIdFlags[record.ClPoolId] = true
	}
	return nil
}

// This is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) ReplaceMigrationRecords(ctx sdk.Context, records ...types.BalancerToConcentratedPoolLink) error {
	migrationInfo := k.GetMigrationInfo(ctx)

	err := k.validateRecords(ctx, records...)
	if err != nil {
		return err
	}

	migrationInfo.BalancerToConcentratedPoolLinks = records

	k.SetMigrationInfo(ctx, migrationInfo)
	return nil
}

// UpdateDistrRecords is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateMigrationRecords(ctx sdk.Context, records ...types.BalancerToConcentratedPoolLink) error {
	recordsMap := make(map[uint64]types.BalancerToConcentratedPoolLink)

	for _, existingRecord := range k.GetMigrationInfo(ctx).BalancerToConcentratedPoolLinks {
		recordsMap[existingRecord.BalancerPoolId] = existingRecord
	}

	err := k.validateRecords(ctx, records...)
	if err != nil {
		return err
	}

	for _, record := range records {
		recordsMap[record.BalancerPoolId] = record
	}

	newRecords := []types.BalancerToConcentratedPoolLink{}

	// if the clPoolId is 0, we remove the entire record
	for _, val := range recordsMap {
		if val.ClPoolId != 0 {
			newRecords = append(newRecords, val)
		}
	}

	sort.SliceStable(newRecords, func(i, j int) bool {
		return newRecords[i].BalancerPoolId < newRecords[j].BalancerPoolId
	})

	k.SetMigrationInfo(ctx, types.MigrationRecords{
		BalancerToConcentratedPoolLinks: newRecords,
	})
	return nil
}
