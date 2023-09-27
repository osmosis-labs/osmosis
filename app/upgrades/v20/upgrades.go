package v20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v19/app/keepers"
	"github.com/osmosis-labs/osmosis/v19/app/upgrades"
	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
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

		// Initialize the newly created param
		keepers.ConcentratedLiquidityKeeper.SetParam(ctx, cltypes.KeyUnrestrictedPoolCreatorWhitelist, []string{})

		// Initialize the new params in incentives for group creation.
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyGroupCreationFee, incentivestypes.DefaultGroupCreationFee)
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyCreatorWhitelist, []string{})

		return migrations, nil
	}
}
