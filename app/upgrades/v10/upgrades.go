package v9

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v7/app/keepers"
	"github.com/osmosis-labs/osmosis/v7/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// TODO: This upgrade is blocked on https://github.com/cosmos/cosmos-sdk/pull/11800

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
