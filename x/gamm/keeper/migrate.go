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
func (k Keeper) validateRecords(ctx sdk.Context, records ...types.GammToConcentratedPoolLink) error {
	lastGammPoolID := uint64(0)
	gammIdFlags := make(map[uint64]bool)

	for _, record := range records {
		if gammIdFlags[record.GammPoolId] {
			return fmt.Errorf(
				"Gamm pool ID #%d has duplications.",
				record.GammPoolId,
			)
		}

		// Ensure records are sorted
		if record.GammPoolId < lastGammPoolID {
			return fmt.Errorf(
				"Gamm pool ID #%d came after Gauge ID #%d.",
				record.GammPoolId, lastGammPoolID,
			)
		}

		// Ensure the first pool exists and that it is of type gamm
		poolType, err := k.GetPoolType(ctx, record.GammPoolId)
		if err != nil {
			return err
		}
		if poolType.String() != "Balancer" {
			return fmt.Errorf("Gamm pool ID #%d is not of type Gamm", record.GammPoolId)
		}

		// Ensure the concentrated pool exists
		// TODO: Get GetPoolType to work for cl pools from gamm
		_, err = k.poolManager.GetPoolModule(ctx, record.GammPoolId)
		if err != nil {
			return err
		}

		lastGammPoolID = record.GammPoolId

		gammIdFlags[record.GammPoolId] = true
	}
	return nil
}

// This is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) ReplaceMigrationRecords(ctx sdk.Context, records ...types.GammToConcentratedPoolLink) error {
	migrationInfo := k.GetMigrationInfo(ctx)

	err := k.validateRecords(ctx, records...)
	if err != nil {
		return err
	}

	migrationInfo.GammToConcentratedPoolLinks = records

	k.SetMigrationInfo(ctx, migrationInfo)
	return nil
}

// UpdateDistrRecords is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateMigrationRecords(ctx sdk.Context, records ...types.GammToConcentratedPoolLink) error {
	recordsMap := make(map[uint64]types.GammToConcentratedPoolLink)

	for _, existingRecord := range k.GetMigrationInfo(ctx).GammToConcentratedPoolLinks {
		recordsMap[existingRecord.GammPoolId] = existingRecord
	}

	err := k.validateRecords(ctx, records...)
	if err != nil {
		return err
	}

	for _, record := range records {
		recordsMap[record.GammPoolId] = record
	}

	newRecords := []types.GammToConcentratedPoolLink{}

	// if the clPoolId is 0, we remove the entire record
	for _, val := range recordsMap {
		if val.ClPoolId != 0 {
			newRecords = append(newRecords, val)
		}
	}

	sort.SliceStable(newRecords, func(i, j int) bool {
		return newRecords[i].GammPoolId < newRecords[j].GammPoolId
	})

	k.SetMigrationInfo(ctx, types.MigrationRecords{
		GammToConcentratedPoolLinks: newRecords,
	})
	return nil
}
