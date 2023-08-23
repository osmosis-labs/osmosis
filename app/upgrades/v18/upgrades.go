package v18

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v17/app/keepers"
	"github.com/osmosis-labs/osmosis/v17/app/upgrades"

	pooltypes "github.com/osmosis-labs/osmosis/v17/x/pool-incentives/types"
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
		err = DisableIncentiveForBalancerPool(ctx, keepers, poolsToRemoveIncentiveFrom)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

func DisableIncentiveForBalancerPool(ctx sdk.Context, keepers *keepers.AppKeepers, poolsToRemoveIncentiveFrom []uint64) error {
	longestDuration, err := keepers.PoolIncentivesKeeper.GetLongestLockableDuration(ctx)
	if err != nil {
		return err
	}

	// check they have existing incentive record
	for _, poolId := range poolsToRemoveIncentiveFrom {
		gaugeId, err := keepers.PoolIncentivesKeeper.GetPoolGaugeId(ctx, poolId, longestDuration)
		if err != nil {
			return err
		}

		distrRecord := pooltypes.DistrRecord{
			GaugeId: gaugeId,
			Weight:  sdk.NewInt(0), // this mean no incentives will be distribtued to this gauge
		}
		err = keepers.PoolIncentivesKeeper.ReplaceDistrRecords(ctx, distrRecord)
		if err != nil {
			return err
		}
	}

	return nil

}
