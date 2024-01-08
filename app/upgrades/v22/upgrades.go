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

		// Migrate legacy taker fee tracker to new taker fee tracker (for performance reasons)

		oldTakerFeeTrackerForStakers := keepers.PoolManagerKeeper.GetLegacyTakerFeeTrackerForStakers(ctx)
		for _, coin := range oldTakerFeeTrackerForStakers {
			err := keepers.PoolManagerKeeper.UpdateTakerFeeTrackerForStakersByDenom(ctx, coin.Denom, coin.Amount)
			if err != nil {
				return nil, err
			}
		}

		oldTakerFeeTrackerForCommunityPool := keepers.PoolManagerKeeper.GetLegacyTakerFeeTrackerForCommunityPool(ctx)
		for _, coin := range oldTakerFeeTrackerForCommunityPool {
			err := keepers.PoolManagerKeeper.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, coin.Denom, coin.Amount)
			if err != nil {
				return nil, err
			}
		}

		return migrations, nil
	}
}
