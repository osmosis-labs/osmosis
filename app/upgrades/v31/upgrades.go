package v31

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v30/app/keepers"
	"github.com/osmosis-labs/osmosis/v30/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v30/x/poolmanager"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		updateTakerFeeDistribution(sdkCtx, keepers.PoolManagerKeeper)

		return migrations, nil
	}
}

// updateTakerFeeDistribution updates the community_pool and burn values in the osmo_taker_fee_distribution
// This changes taker fees from being sent to the community pool to being burned instead.
func updateTakerFeeDistribution(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper) {
	poolManagerParams := poolManagerKeeper.GetParams(ctx)

	// Set community_pool to 0, burn and staking rewards to 50% each
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool = osmomath.ZeroDec()
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.7")
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.3")

	// Set burn to 0 for non-OSMO taker fees
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn = osmomath.ZeroDec()

	poolManagerKeeper.SetParams(ctx, poolManagerParams)
}
