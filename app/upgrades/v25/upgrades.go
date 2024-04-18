package v25

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v24/app/keepers"
	"github.com/osmosis-labs/osmosis/v24/app/upgrades"
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

		// update block-sdk params
		if err := keepers.AuctionKeeper.SetParams(ctx, AuctionParams); err != nil {
			return nil, err
		}

		keepers.TwapKeeper.DeleteDeprecatedHistoricalTWAPsIsPruning(ctx)

		// Set the authenticator params in the store
		authenticatorParams := keepers.SmartAccountKeeper.GetParams(ctx)
		authenticatorParams.MaximumUnauthenticatedGas = 120_000
		authenticatorParams.IsSmartAccountActive = false
		keepers.SmartAccountKeeper.SetParams(ctx, authenticatorParams)

		return migrations, nil
	}
}
