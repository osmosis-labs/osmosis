package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v9/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v9/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v9/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v9/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v9/x/superfluid/types"
)

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) {
}

func (k Keeper) AfterEpochStartBeginBlock(ctx sdk.Context) {
	// cref [#830](https://github.com/osmosis-labs/osmosis/issues/830),
	// the supplied epoch number is wrong at time of commit. hence we get from the info.
	curEpoch := k.ek.GetEpochInfo(ctx, k.GetEpochIdentifier(ctx)).CurrentEpoch

	// Move delegation rewards to perpetual gauge
	ctx.Logger().Info("Move delegation rewards to gauges")
	k.MoveSuperfluidDelegationRewardToGauges(ctx)

	ctx.Logger().Info("Distribute Superfluid gauges")
	k.distributeSuperfluidGauges(ctx)

	// Update all LP tokens multipliers for the upcoming epoch.
	// This affects staking reward distribution until the next epochs rewards.
	// Exclusive of current epoch's rewards, inclusive of next epoch's rewards.
	ctx.Logger().Info("Update all osmo equivalency multipliers")
	for _, asset := range k.GetAllSuperfluidAssets(ctx) {
		err := k.UpdateOsmoEquivalentMultipliers(ctx, asset, curEpoch)
		if err != nil {
			// TODO: Revisit what we do here. (halt all distr, only skip this asset)
			// Since at MVP of feature, we only have one pool of superfluid staking,
			// we can punt this question.
			// each of the errors feels like significant misconfig
			return
		}
	}

	// Refresh intermediary accounts' delegation amounts,
	// making staking rewards follow the updated multiplier numbers.
	ctx.Logger().Info("Refresh all superfluid delegation amounts")
	k.RefreshIntermediaryDelegationAmounts(ctx)
}

func (k Keeper) MoveSuperfluidDelegationRewardToGauges(ctx sdk.Context) {
	accs := k.GetAllIntermediaryAccounts(ctx)
	for _, acc := range accs {
		addr := acc.GetAccAddress()
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		// To avoid unexpected issues on WithdrawDelegationRewards and AddToGaugeRewards
		// we use cacheCtx and apply the changes later
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			_, err := k.dk.WithdrawDelegationRewards(cacheCtx, addr, valAddr)
			return err
		})

		// Send delegation rewards to gauges
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// Note! We only send the bond denom (osmo), to avoid attack vectors where people
			// send many different denoms to the intermediary account, and make a resource exhaustion attack on end block.
			bondDenom := k.sk.BondDenom(cacheCtx)
			balance := k.bk.GetBalance(cacheCtx, addr, bondDenom)
			if balance.IsZero() {
				return nil
			}
			return k.ik.AddToGaugeRewards(cacheCtx, addr, sdk.Coins{balance}, acc.GaugeId)
		})
	}
}

func (k Keeper) distributeSuperfluidGauges(ctx sdk.Context) {
	gauges := k.ik.GetActiveGauges(ctx)

	// only distribute to active gauges that are for perpetual synthetic denoms
	distrGauges := []incentivestypes.Gauge{}
	for _, gauge := range gauges {
		isSynthetic := lockuptypes.IsSyntheticDenom(gauge.DistributeTo.Denom)
		if isSynthetic && gauge.IsPerpetual {
			distrGauges = append(distrGauges, gauge)
		}
	}
	_, err := k.ik.Distribute(ctx, distrGauges)
	if err != nil {
		panic(err)
	}
}

func (k Keeper) UpdateOsmoEquivalentMultipliers(ctx sdk.Context, asset types.SuperfluidAsset, newEpochNumber int64) error {
	if asset.AssetType == types.SuperfluidAssetTypeLPShare {
		// LP_token_Osmo_equivalent = OSMO_amount_on_pool / LP_token_supply
		poolId := gammtypes.MustGetPoolIdFromShareDenom(asset.Denom)
		pool, err := k.gk.GetPoolAndPoke(ctx, poolId)
		if err != nil {
			// Pool has been unexpectedly deleted
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		// get OSMO amount
		bondDenom := k.sk.BondDenom(ctx)
		osmoPoolAsset := pool.GetTotalPoolLiquidity(ctx).AmountOf(bondDenom)
		if osmoPoolAsset.IsZero() {
			// Pool has unexpectedly removed Osmo from its assets.
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		multiplier := k.calculateOsmoBackingPerShare(pool, osmoPoolAsset)
		k.SetOsmoEquivalentMultiplier(ctx, newEpochNumber, asset.Denom, multiplier)
	} else if asset.AssetType == types.SuperfluidAssetTypeNative {
		// TODO: Consider deleting superfluid asset type native
		k.Logger(ctx).Error("unsupported superfluid asset type")
		return errors.New("SuperfluidAssetTypeNative is unsupported")
	}
	return nil
}
