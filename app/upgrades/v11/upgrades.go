package v11

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v10/app/keepers"
	"github.com/osmosis-labs/osmosis/v10/app/upgrades"
)

// We set the app version to pre-upgrade because it will be incremented by one
// after the upgrade is applied by the handler.
const preUpgradeAppVersion = 10

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Although the app version was already set during the v9 upgrade, our v10 was a fork.
		// As a result, the upgrade handler was not executed to increment the app version.
		// This change helps to correctly set the app version for v11.
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

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
