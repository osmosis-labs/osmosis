package v33

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v32/x/incentives/keeper"
	mintkeeper "github.com/osmosis-labs/osmosis/v32/x/mint/keeper"
	minttypes "github.com/osmosis-labs/osmosis/v32/x/mint/types"
)

// CreateUpgradeHandler returns the v33 upgrade handler. After the module migrations it removes the
// developer rewards share from mint distribution and recovers the HUAHUA held by gauges 1954-1959
// (see HuahuaStuckGauges).
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

		// Removing the developer rewards share changes the mint schedule, so a failure halts the upgrade.
		if err := removeDeveloperRewardsFromMint(sdkCtx, keepers.MintKeeper); err != nil {
			return nil, err
		}

		// The recovery is not critical to the chain, so a failure is logged instead of halting the upgrade.
		// It runs in a cache context so that it is applied either entirely or not at all.
		cacheCtx, write := sdkCtx.CacheContext()
		recovered, err := recoverHuahuaFromStuckGauges(cacheCtx, keepers.IncentivesKeeper)
		if err != nil {
			sdkCtx.Logger().Error("v33 upgrade: HUAHUA recovery skipped", "error", err)
		} else {
			write()
			sdkCtx.Logger().Info("v33 upgrade: HUAHUA recovered", "amount", recovered.String(), "recipient", HuahuaRecoveryAddress)
		}

		return migrations, nil
	}
}

// recoverHuahuaFromStuckGauges finishes the stuck HUAHUA gauges and sends their undistributed coins
// to the recovery address. See HuahuaStuckGauges.
//
// All undistributed coins of each gauge are sent, not only HUAHUA. On mainnet these gauges hold only
// HUAHUA, but MsgAddToGauge can add other denoms before the upgrade; once the gauge is finished those
// coins could never be distributed or withdrawn, so they are sent along with the HUAHUA.
func recoverHuahuaFromStuckGauges(ctx sdk.Context, incentivesKeeper *incentiveskeeper.Keeper) (sdk.Coins, error) {
	recipient, err := sdk.AccAddressFromBech32(HuahuaRecoveryAddress)
	if err != nil {
		return nil, err
	}

	recovered := sdk.NewCoins()
	for _, stuck := range HuahuaStuckGauges {
		gauge, err := incentivesKeeper.GetGaugeByID(ctx, stuck.GaugeID)
		if err != nil {
			return nil, err
		}
		// Guard against the IDs pointing to anything other than the expected gauges.
		if gauge.DistributeTo.Denom != stuck.Denom {
			return nil, fmt.Errorf("gauge %d distributes to %q, expected %q", stuck.GaugeID, gauge.DistributeTo.Denom, stuck.Denom)
		}
		if gauge.Coins.AmountOf(HuahuaDenom).IsZero() {
			return nil, fmt.Errorf("gauge %d holds no HUAHUA", stuck.GaugeID)
		}

		coins, err := incentivesKeeper.ForceFinishGaugeAndSendUndistributed(ctx, stuck.GaugeID, recipient)
		if err != nil {
			return nil, err
		}
		recovered = recovered.Add(coins...)
	}
	return recovered, nil
}

// removeDeveloperRewardsFromMint takes the developer rewards share out of mint distribution without
// changing the amount of OSMO that enters circulation each epoch.
//
// Previously the full epoch provisions were minted, the developer rewards share was burned and the same
// amount was paid out of the developer vesting module account. Only (1 - developer_rewards) of the
// provisions was new supply. The epoch provisions are scaled by that factor and the remaining
// proportions are rescaled by its inverse, so every recipient keeps its per-epoch amount. The
// community pool takes the remainder so the proportions sum to exactly one.
//
// The developer vesting module account and its balance are left untouched.
func removeDeveloperRewardsFromMint(ctx sdk.Context, mintKeeper *mintkeeper.Keeper) error {
	params := mintKeeper.GetParams(ctx)
	old := params.DistributionProportions
	if old.DeveloperRewards.IsZero() {
		return nil
	}

	inflationFactor := osmomath.OneDec().Sub(old.DeveloperRewards)
	if !inflationFactor.IsPositive() {
		return fmt.Errorf("developer rewards proportion %s leaves no inflation to rescale", old.DeveloperRewards)
	}

	staking := old.Staking.Quo(inflationFactor)
	poolIncentives := old.PoolIncentives.Quo(inflationFactor)
	params.DistributionProportions = minttypes.DistributionProportions{
		Staking:          staking,
		PoolIncentives:   poolIncentives,
		DeveloperRewards: osmomath.ZeroDec(),
		CommunityPool:    osmomath.OneDec().Sub(staking).Sub(poolIncentives),
	}
	params.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{}
	if err := params.Validate(); err != nil {
		return fmt.Errorf("rescaled mint params are invalid: %w", err)
	}

	minter := mintKeeper.GetMinter(ctx)
	oldEpochProvisions := minter.EpochProvisions
	minter.EpochProvisions = oldEpochProvisions.Mul(inflationFactor)
	if err := minter.Validate(); err != nil {
		return fmt.Errorf("rescaled minter is invalid: %w", err)
	}

	mintKeeper.SetParams(ctx, params)
	mintKeeper.SetMinter(ctx, minter)

	ctx.Logger().Info("v33 upgrade: removed developer rewards from mint distribution",
		"old_proportions", old.String(), "new_proportions", params.DistributionProportions.String(),
		"old_epoch_provisions", oldEpochProvisions.String(), "new_epoch_provisions", minter.EpochProvisions.String())
	return nil
}
