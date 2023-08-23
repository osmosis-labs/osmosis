package v18

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		poolsToRemoveIncentiveFrom := []uint64{1093, 1108, 1106, 1092, 1101, 1097}
		err = DisableIncentiveRecord(ctx, keepers, poolsToRemoveIncentiveFrom)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

func DisableIncentiveRecord(ctx sdk.Context, keepers *keepers.AppKeepers, poolsToRemoveIncentiveFrom []uint64) error {
	// check they have existing incentive record
	for _, poolId := range poolsToRemoveIncentiveFrom {
		incRecords, err := keepers.ConcentratedLiquidityKeeper.GetAllIncentiveRecordsForPool(ctx, poolId)
		if err != nil {
			return err
		}

		//remove all incentive records for incentive record
		for _, incRecord := range incRecords {
			err := keepers.ConcentratedLiquidityKeeper.RemoveIncentiveRecords(ctx, incRecord)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
