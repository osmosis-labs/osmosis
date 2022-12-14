package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
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

func (m Migrator) Migrate1To2(ctx sdk.Context) error {
	historicalPoolIndexedRecords, err := m.keeper.getAllHistoricalPoolIndexedTWAPs(ctx)
	if err != nil {
		return err
	}

	for _, record := range historicalPoolIndexedRecords {
		record.GeometricTwapAccumulator = sdk.ZeroDec()
		m.keeper.storeHistoricalTWAP(ctx, record)
	}

	historicalTimeIndexed, err := m.keeper.getAllHistoricalTimeIndexedTWAPs(ctx)
	if err != nil {
		return err
	}

	for _, record := range historicalTimeIndexed {
		record.GeometricTwapAccumulator = sdk.ZeroDec()
		m.keeper.storeHistoricalTWAP(ctx, record)
	}

	mostRecent, err := types.GetAllMostRecentTwaps(ctx.KVStore(m.keeper.storeKey))
	if err != nil {
		return err
	}

	for _, record := range mostRecent {
		record.GeometricTwapAccumulator = sdk.ZeroDec()
		m.keeper.storeHistoricalTWAP(ctx, record)
	}

	return nil
}
