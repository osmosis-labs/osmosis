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

		// We no longer use the KVStore base denoms and instead use the param store for performance reasons.
		// We retrieve the base denoms from the KVStore, set them in the param store, then delete them here.
		baseDenoms, err := keepers.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(ctx)
		if err != nil {
			return nil, err
		}
		keepers.ProtoRevKeeper.SetBaseDenoms(ctx, baseDenoms)
		keepers.ProtoRevKeeper.DeprecatedDeleteBaseDenoms(ctx)

		return migrations, nil
	}
}
