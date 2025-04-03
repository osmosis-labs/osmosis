package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) error {
	return nil
}

func (k Keeper) AfterEpochStartBeginBlock(ctx sdk.Context) {
	// cref [#830](https://github.com/osmosis-labs/osmosis/issues/830),
	// the supplied epoch number is wrong at time of commit. hence we get from the info.
	curEpoch := k.ek.GetEpochInfo(ctx, k.GetEpochIdentifier(ctx)).CurrentEpoch

	accs := k.GetAllIntermediaryAccounts(ctx)
	// Move delegation rewards to perpetual gauge
	ctx.Logger().Info("Move delegation rewards to gauges")
	k.MoveSuperfluidDelegationRewardToGauges(ctx, accs)

	ctx.Logger().Info("Distribute Superfluid gauges")
	//nolint:errcheck
	osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		k.distributeSuperfluidGauges(cacheCtx)
		return nil
	})

	// Update all LP tokens multipliers for the upcoming epoch.
	// This affects staking reward distribution until the next epochs rewards.
	// Exclusive of current epoch's rewards, inclusive of next epoch's rewards.
	ctx.Logger().Info("Update all osmo equivalency multipliers")
	for _, asset := range k.GetAllSuperfluidAssets(ctx) {
		err := k.UpdateOsmoEquivalentMultipliers(ctx, asset, curEpoch)
		if err != nil {
			// UPDATE: balancer pools are expected to be skipped only on error due to being
			// already well tested in production.
			//
			// CL pools are surrounded by ApplyFuncIfNoError, so they are silently skipped on error or panic.
			return
		}
	}

	// Refresh intermediary accounts' delegation amounts,
	// making staking rewards follow the updated multiplier numbers.
	ctx.Logger().Info("Refresh all superfluid delegation amounts")
	k.RefreshIntermediaryDelegationAmounts(ctx, accs)
}

func (k Keeper) MoveSuperfluidDelegationRewardToGauges(ctx sdk.Context, accs []types.SuperfluidIntermediaryAccount) {
	bondDenom, err := k.sk.BondDenom(ctx)
	if err != nil {
		panic(err)
	}
	for _, acc := range accs {
		addr := acc.GetAccAddress()
		valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
		if err != nil {
			panic(err)
		}

		// To avoid unexpected issues on WithdrawDelegationRewards and AddToGaugeRewards
		// we use cacheCtx and apply the changes later
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			_, err := k.ck.WithdrawDelegationRewards(cacheCtx, addr, valAddr)
			if errors.Is(err, distributiontypes.ErrEmptyDelegationDistInfo) {
				ctx.Logger().Debug("no delegations for this (pool, validator) pair, skipping...")
				// TODO: Remove this account from IntermediaryAccounts that we iterate over
				return nil
			} else if err != nil {
				return err
			}

			// Send delegation rewards to gauges
			// Note! We only send the bond denom (osmo), to avoid attack vectors where people
			// send many different denoms to the intermediary account, and make a resource exhaustion attack on end block.
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
		// we filter superfluid gauges by using the distributeTo denom in the gauge.
		// If the denom in the gauge is a synthetic denom, we append the gauge to the gauge list to distribute to.
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
		bondDenom, err := k.sk.BondDenom(ctx)
		if err != nil {
			return err
		}
		osmoPoolAsset := pool.GetTotalPoolLiquidity(ctx).AmountOf(bondDenom)
		if osmoPoolAsset.IsZero() {
			err := fmt.Errorf("pool %d has zero OSMO amount", poolId)
			// Pool has unexpectedly removed Osmo from its assets.
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		multiplier := k.calculateOsmoBackingPerShare(pool, osmoPoolAsset)
		k.SetOsmoEquivalentMultiplier(ctx, newEpochNumber, asset.Denom, multiplier)
	} else if asset.AssetType == types.SuperfluidAssetTypeConcentratedShare {
		// https://github.com/osmosis-labs/osmosis/issues/6229
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			return k.updateConcentratedOsmoEquivalentMultiplier(cacheCtx, asset, newEpochNumber)
		})
	} else if asset.AssetType == types.SuperfluidAssetTypeNative {
		// TODO: Consider deleting superfluid asset type native
		k.Logger(ctx).Error("unsupported superfluid asset type")
		return errors.New("SuperfluidAssetTypeNative is unsupported")
	}
	return nil
}

// updateConcentratedOsmoEquivalentMultiplier runs the logic for updating the OSMO equivalent multiplier for a concentrated liquidity pool.
func (k Keeper) updateConcentratedOsmoEquivalentMultiplier(ctx sdk.Context, asset types.SuperfluidAsset, newEpochNumber int64) error {
	// LP_token_Osmo_equivalent = OSMO_amount_on_pool / LP_token_supply
	poolId := cltypes.MustGetPoolIdFromShareDenom(asset.Denom)
	pool, err := k.clk.GetConcentratedPoolById(ctx, poolId)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		// Pool has unexpectedly removed Osmo from its assets.
		k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
		return err
	}

	// get underlying assets from all liquidity in a full range position
	// note: this is not the same as the total liquidity in the pool, as this includes positions not in the full range
	fullRangeLiquidity, err := k.clk.GetFullRangeLiquidityInPool(ctx, poolId)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return fmt.Errorf("failed to retrieve full range liquidity from pool (%d): %w", poolId, err)
	}

	position := model.Position{
		LowerTick: cltypes.MinInitializedTick,
		UpperTick: cltypes.MaxTick,
		Liquidity: fullRangeLiquidity,
	}
	// Note that the returned amounts are rounded up. This should be fine as they both are used for calculating the multiplier.
	asset0, asset1, err := cl.CalculateUnderlyingAssetsFromPosition(ctx, position, pool)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
		return err
	}
	assets := sdk.NewCoins(asset0, asset1)

	// get OSMO amount from underlying assets
	bondDenom, err := k.sk.BondDenom(ctx)
	if err != nil {
		return err
	}
	osmoPoolAsset := assets.AmountOf(bondDenom)
	if osmoPoolAsset.IsZero() {
		// Pool has unexpectedly removed OSMO from its assets.
		err := errors.New("pool has unexpectedly removed OSMO as one of its underlying assets")
		k.Logger(ctx).Error(err.Error())
		k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
		return err
	}

	// calculate multiplier and set it
	multiplier := osmoPoolAsset.ToLegacyDec().Quo(fullRangeLiquidity)
	k.SetOsmoEquivalentMultiplier(ctx, newEpochNumber, asset.Denom, multiplier)

	return nil
}
