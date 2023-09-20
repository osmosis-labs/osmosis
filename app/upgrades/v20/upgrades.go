package v20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
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

		// TODO START: DO NOT LEAVE THIS AS IS.
		// See my comment here https://github.com/osmosis-labs/osmosis/pull/6420#issuecomment-1726969370
		keepers.ConcentratedLiquidityKeeper.SetParams(ctx, concentratedliquiditytypes.DefaultParams())
		// TODO END

		return migrations, nil
	}
}
