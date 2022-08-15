package v12

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v11/app/keepers"
	"github.com/osmosis-labs/osmosis/v11/app/upgrades"
)

// We set the app version to pre-upgrade because it will be incremented by one
// after the upgrade is applied by the handler.
const preUpgradeAppVersion = 11

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Although the app version was already set during the v9 upgrade, our v10 was a fork and
		// v11 was decided to be limited to the "gauge creation minimum fee" change only:
		// https://github.com/osmosis-labs/osmosis/pull/2202
		//
		// As a result, the upgrade handler was not executed to increment the app version.
		// This change helps to correctly set the app version for v12.
		if err := keepers.UpgradeKeeper.SetAppVersion(ctx, preUpgradeAppVersion); err != nil {
			return nil, err
		}

		// Set the max_age_num_blocks in the evidence params to reflect the 14 day
		// unbonding period.
		//
		// Ref: https://github.com/osmosis-labs/osmosis/issues/1160
		cp := bpm.GetConsensusParams(ctx)
		if cp != nil && cp.Evidence != nil {
			evParams := cp.Evidence
			evParams.MaxAgeNumBlocks = 186_092

			bpm.StoreConsensusParams(ctx, cp)
		}

		// Initialize TWAP state
		// TODO: Get allPoolIds from gamm keeper, and write test for migration.
		allPoolIds := []uint64{}
		err := keepers.TwapKeeper.MigrateExistingPools(ctx, allPoolIds)
		if err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
