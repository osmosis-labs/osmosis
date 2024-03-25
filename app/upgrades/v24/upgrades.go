package v24

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/app/upgrades"

	incentivestypes "github.com/osmosis-labs/osmosis/v23/x/incentives/types"
)

const (
	mainnetChainID = "osmosis-1"
	// Edgenet is to function exactly the samas mainnet, and expected
	// to be state-exported from mainnet state.
	edgenetChainID = "edgenet"
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

		// We no longer use the base denoms array and instead use the repeated base denoms field for performance reasons.
		// We retrieve the old base denoms array from the KVStore, delete the array from the KVStore, and set them as a repeated field in the new KVStore.
		baseDenoms, err := keepers.ProtoRevKeeper.DeprecatedGetAllBaseDenoms(ctx)
		if err != nil {
			return nil, err
		}
		keepers.ProtoRevKeeper.DeprecatedDeleteBaseDenoms(ctx)
		err = keepers.ProtoRevKeeper.SetBaseDenoms(ctx, baseDenoms)
		if err != nil {
			return nil, err
		}

		// Now that the TWAP keys are refactored, we can delete all time indexed TWAPs
		// since we only need the pool indexed TWAPs.
		keepers.TwapKeeper.DeleteAllHistoricalTimeIndexedTWAPs(ctx)

		// restrict lockable durations to 2 weeks per
		// https://wallet.keplr.app/chains/osmosis/proposals/400
		chainID := ctx.ChainID()

		if chainID == mainnetChainID || chainID == edgenetChainID {
			keepers.IncentivesKeeper.SetLockableDurations(ctx, []time.Duration{
				time.Hour * 24 * 14,
			})
			keepers.PoolIncentivesKeeper.SetLockableDurations(ctx, []time.Duration{
				time.Hour * 24 * 14,
			})
		}

		// Set the new min value for distribution for the incentives module.
		// https://www.mintscan.io/osmosis/proposals/733
		keepers.IncentivesKeeper.SetParam(ctx, incentivestypes.KeyMinValueForDistr, incentivestypes.DefaultMinValueForDistr)

		return migrations, nil
	}
}
