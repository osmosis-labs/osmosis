package v31

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v31/app/keepers"
	"github.com/osmosis-labs/osmosis/v31/app/upgrades"
	poolmanager "github.com/osmosis-labs/osmosis/v31/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v31/x/txfees/types"
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

		err = updateTakerFeeDistribution(sdkCtx, keepers.PoolManagerKeeper, keepers.AccountKeeper)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

// updateTakerFeeDistribution updates the community_pool and burn values in the osmo_taker_fee_distribution
// This changes taker fees from being sent to the community pool to being burned instead.
// It also sets up the staking rewards smoothing feature with a smoothing factor of 7.
func updateTakerFeeDistribution(ctx sdk.Context, poolManagerKeeper *poolmanager.Keeper, accountKeeper *authkeeper.AccountKeeper) error {
	// Set OSMO taker fee distribution: community_pool to 0, burn and staking rewards to 70%:30%
	osmoTakerFeeDistribution := poolmanagertypes.TakerFeeDistributionPercentage{
		CommunityPool:  osmomath.ZeroDec(),
		Burn:           osmomath.MustNewDecFromStr("0.7"),
		StakingRewards: osmomath.MustNewDecFromStr("0.3"),
	}
	poolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyOsmoTakerFeeDistribution, osmoTakerFeeDistribution)

	// Set non-OSMO taker fee distribution: staking_rewards=22.5%, burn=52.5%, community_pool=25%
	nonOsmoTakerFeeDistribution := poolmanagertypes.TakerFeeDistributionPercentage{
		StakingRewards: osmomath.MustNewDecFromStr("0.225"),
		Burn:           osmomath.MustNewDecFromStr("0.525"),
		CommunityPool:  osmomath.MustNewDecFromStr("0.25"),
	}
	poolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyNonOsmoTakerFeeDistribution, nonOsmoTakerFeeDistribution)

	// Set daily staking rewards smoothing factor to 7
	// This distributes 1/7th of the staking rewards buffer each day to smooth APR display
	dailyStakingRewardsSmoothingFactor := uint64(7)
	poolManagerKeeper.SetParam(ctx, poolmanagertypes.KeyDailyStakingRewardsSmoothingFactor, dailyStakingRewardsSmoothingFactor)

	// Ensure new module account exists for non‑native taker fee burn bucket. Error if it already exists.
	err := osmoutils.CreateModuleAccountByName(ctx, accountKeeper, txfeestypes.TakerFeeBurnName)
	if err != nil {
		return err
	}

	// Create the staking rewards smoothing buffer module account
	err = osmoutils.CreateModuleAccountByName(ctx, accountKeeper, txfeestypes.TakerFeeStakingRewardsBuffer)
	if err != nil {
		return err
	}

	return nil
}
