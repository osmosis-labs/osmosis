package v33

import (
	"context"
	"fmt"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	"github.com/osmosis-labs/osmosis/v32/app/upgrades"
	incentiveskeeper "github.com/osmosis-labs/osmosis/v32/x/incentives/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v32/x/lockup/types"
	mintkeeper "github.com/osmosis-labs/osmosis/v32/x/mint/keeper"
)

// CreateUpgradeHandler returns the v33 upgrade handler. After the module migrations it:
//   - recovers the HUAHUA held by gauges 1954-1959 (see HuahuaStuckGauges),
//   - moves any mint pool incentives proportion to the community pool, since minted coins are no
//     longer allocated to pool incentives,
//   - finishes the stale gauges and sends their undistributed coins to the community pool
//     (see pruneStaleGauges).
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

		movePoolIncentivesProportionToCommunityPool(sdkCtx, keepers.MintKeeper)

		// Like the HUAHUA recovery, pruning is not critical to the chain. It runs after the recovery so
		// that the HUAHUA gauges are never pruned, even if the recovery failed.
		cacheCtx, write = sdkCtx.CacheContext()
		pruned, prunedCoins, err := pruneStaleGauges(cacheCtx, keepers.IncentivesKeeper)
		if err != nil {
			sdkCtx.Logger().Error("v33 upgrade: stale gauge pruning skipped", "error", err)
		} else {
			write()
			sdkCtx.Logger().Info("v33 upgrade: stale gauges pruned", "count", pruned, "community_pool_amount", prunedCoins.String())
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

// movePoolIncentivesProportionToCommunityPool adds the mint pool incentives proportion to the community pool
// proportion and zeroes it, so that the mint parameters satisfy the new validation. It is a no-op on mainnet,
// where the proportion is already zero.
func movePoolIncentivesProportionToCommunityPool(ctx sdk.Context, mintKeeper *mintkeeper.Keeper) {
	params := mintKeeper.GetParams(ctx)
	proportions := params.DistributionProportions
	if proportions.PoolIncentives.IsZero() {
		return
	}

	proportions.CommunityPool = proportions.CommunityPool.Add(proportions.PoolIncentives)
	proportions.PoolIncentives = osmomath.ZeroDec()
	params.DistributionProportions = proportions
	mintKeeper.SetParams(ctx, params)
}

// pruneStaleGauges finishes every stale gauge and sends their undistributed coins to the community pool.
// It returns the number of gauges finished and the coins sent.
//
// A gauge is stale if it is an active, non-perpetual lock gauge that started before StaleGaugeStartCutoff
// and never completed a single distribution. Such a gauge only progresses when locks qualify for it, so
// these gauges have been skipped every epoch for years while still being loaded and checked each epoch.
// On mainnet this matches 14,540 gauges, almost all created in 2022 and funded with LP shares.
func pruneStaleGauges(ctx sdk.Context, incentivesKeeper *incentiveskeeper.Keeper) (int, sdk.Coins, error) {
	excluded := make(map[uint64]struct{}, len(HuahuaStuckGauges))
	for _, stuck := range HuahuaStuckGauges {
		excluded[stuck.GaugeID] = struct{}{}
	}

	staleGaugeIDs := []uint64{}
	for _, gauge := range incentivesKeeper.GetActiveGauges(ctx) {
		if _, ok := excluded[gauge.Id]; ok {
			continue
		}
		if isStaleGauge(gauge.IsPerpetual, gauge.DistributeTo.LockQueryType, gauge.FilledEpochs, gauge.StartTime) {
			staleGaugeIDs = append(staleGaugeIDs, gauge.Id)
		}
	}

	coins, err := incentivesKeeper.ForceFinishGaugesToCommunityPool(ctx, staleGaugeIDs)
	if err != nil {
		return 0, nil, err
	}
	return len(staleGaugeIDs), coins, nil
}

func isStaleGauge(isPerpetual bool, lockQueryType lockuptypes.LockQueryType, filledEpochs uint64, startTime time.Time) bool {
	return !isPerpetual &&
		lockQueryType == lockuptypes.ByDuration &&
		filledEpochs == 0 &&
		startTime.Before(StaleGaugeStartCutoff)
}
