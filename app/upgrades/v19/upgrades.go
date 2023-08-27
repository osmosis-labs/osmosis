package v19

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	gammtypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"

	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
)

// OSMO / DAI CL pool ID
const lastPoolToCorrect = 1066

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

		for id := 1; id < lastPoolToCorrect; id++ {
			resetSuperfluidSumtree(keepers, ctx, uint64(id))
		}
		return migrations, nil
	}
}

func resetSuperfluidSumtree(keepers *keepers.AppKeepers, ctx sdk.Context, id uint64) {
	denom := gammtypes.GetPoolShareDenom(id)
	keepers.LockupKeeper.RebuildSuperfluidAccumulationStoresForDenom(ctx, denom)
}
