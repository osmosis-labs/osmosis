package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v15/app/keepers"
	"github.com/osmosis-labs/osmosis/v15/app/upgrades"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v15/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v15/x/tokenfactory/types"
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

		updateTokenFactoryParams(ctx, keepers.TokenFactoryKeeper)

		return migrations, nil
	}
}

func updateTokenFactoryParams(ctx sdk.Context, tokenFactoryKeeper *tokenfactorykeeper.Keeper) {
	tokenFactoryKeeper.SetParams(ctx, tokenfactorytypes.NewParams(nil, NewDenomCreationGasConsume))
}
