package v22

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v21/app/keepers"
	"github.com/osmosis-labs/osmosis/v21/app/upgrades"
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

		// Increase the tx size cost per byte to 20 to reduce the exploitability of bandwidth amplification problems.
		accountParams := keepers.AccountKeeper.GetParams(ctx)
		accountParams.TxSizeCostPerByte = 20 // Double from the default value of 10
		keepers.AccountKeeper.SetParams(ctx, accountParams)

		return migrations, nil
	}
}
