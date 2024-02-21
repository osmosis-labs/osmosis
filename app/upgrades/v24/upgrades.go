package v24

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/app/upgrades"
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

		// Now that the TWAP keys are refactored, we can delete all time indexed TWAPs
		// since we only need the pool indexed TWAPs.
		keepers.TwapKeeper.DeleteAllHistoricalTimeIndexedTWAPs(ctx)

		// Set the authenticator params in the store
		authenticatorParams := keepers.AuthenticatorKeeper.GetParams(ctx)
		authenticatorParams.MaximumUnauthenticatedGas = 50000
		authenticatorParams.AreSmartAccountsActive = true
		keepers.AuthenticatorKeeper.SetParams(ctx, authenticatorParams)

		return migrations, nil
	}
}
