package twap

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1To2GeometricTwapAcc migrates all existing twap records
// by initializing their geometric twap accumulator to zero.
// Return nil on success, error otherwise.
func (m Migrator) Migrate1To2GeometricTwapAcc(ctx sdk.Context) error {
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

// initializeGeometricTwapAcc initializes all existing twap records in state
// with the given value. Returns nil on success, error otherwise.
// Returns error if:
// - fails to get historical time indexed records.
// - does not find any time indexed records.
// - the time indexed records returned from state are written back
// in incorrect order. The order must be ascending by time so that
// we update the "most recent" index with the right records correctly.
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

			isInvalidOrder := previousRecord.Time.After(record.Time)
			if isInvalidOrder {
				return fmt.Errorf("error: historical twap records are not in ascending order, (%v), was after (%v)", previousRecord, record)
			}
		}

		record.GeometricTwapAccumulator = value
		// The call to this updates all of:
		// - time-indexed records
		// - pool-indexed records
		// - most recent records
		k.storeNewRecord(ctx, record)
	}
	return nil
}
