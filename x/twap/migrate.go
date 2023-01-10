package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

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
