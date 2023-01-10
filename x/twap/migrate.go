package twap

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate1To2(ctx sdk.Context) error {
	return m.keeper.initializeGeometricTwapAcc(ctx, sdk.ZeroDec())
}

// MigrateExistingPools iterates through all pools and creates state entry for the twap module.
func (k Keeper) MigrateExistingPools(ctx sdk.Context, latestPoolId uint64) error {
	for i := uint64(1); i <= latestPoolId; i++ {
		err := k.afterCreatePool(ctx, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) MigrateTwapRecordsToGeometric(ctx sdk.Context) error {

	// types

	allMostRecetRecords, err := k.getAllMostRecentRecords(ctx)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)

	for _, record := range allMostRecetRecords {
		record := record
		record.GeometricTwapAccumulator = sdk.ZeroDec()
		key := types.FormatMostRecentTWAPKey(record.PoolId, record.Asset0Denom, record.Asset1Denom)
		osmoutils.MustSet(store, key, &record)
	}

	// make available for GC
	allMostRecetRecords = nil

	allHistoricalRecords, err := k.getAllHistoricalPoolIndexedTWAPs(ctx)

	for _, record := range allHistoricalRecords {
		record := record
		record.GeometricTwapAccumulator = sdk.ZeroDec()
		k.storeHistoricalTWAP(ctx, record)
	}

	// make available for GC
	allHistoricalRecords = nil

	return nil
}

func (k Keeper) initializeGeometricTwapAcc(ctx sdk.Context, value sdk.Dec) error {
	// In ascending order by time.
	historicalTimeIndexed, err := k.getAllHistoricalTimeIndexedTWAPs(ctx)
	if err != nil {
		return err
	}

	if len(historicalTimeIndexed) == 0 {
		return errors.New("error: no historical twap records found")
	}

	// Since we are iterate over time-indexed records in ascending order,
	// most recent record should also be updated correctly.
	for i, record := range historicalTimeIndexed {
		record := record
		// Sanity check order.
		if i > 0 {
			previousRecord := historicalTimeIndexed[i-1]

			isSamePool := previousRecord.PoolId == record.PoolId
			isCorrectOrder := previousRecord.Time.Before(record.Time)
			if !isCorrectOrder && !isSamePool {
				return fmt.Errorf("error: historical twap records are not in ascending order, (%v), was after (%v)", previousRecord, record)
			}
		}

		record.GeometricTwapAccumulator = value
		k.storeNewRecord(ctx, record)
	}
	return nil
}
