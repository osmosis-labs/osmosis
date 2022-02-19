package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.RefreshEpochIdentifier {
		// cref [#830](https://github.com/osmosis-labs/osmosis/issues/830),
		// the supplied epoch number is wrong at time of commit. hence we get from the info.
		endedEpochNumber := k.ek.GetEpochInfo(ctx, epochIdentifier).CurrentEpoch

		// Move delegation rewards to perpetual gauge
		k.MoveSuperfluidDelegationRewardToGauges(ctx)

		// Update all LP tokens TWAP's for the upcoming epoch.
		// This affects staking reward distribution until the next epochs rewards.
		// Exclusive of current epoch's rewards, inclusive of next epoch's rewards.
		for _, asset := range k.GetAllSuperfluidAssets(ctx) {
			err := k.updateEpochTwap(ctx, asset, endedEpochNumber)
			if err != nil {
				// TODO: Revisit what we do here. (halt all distr, only skip this asset)
				// Since at MVP of feature, we only have one pool of superfluid staking,
				// we can punt this question.
				// each of the errors feels like significant misconfig
				return
			}
		}

		// Refresh intermediary accounts' delegation amounts,
		// making staking rewards follow the updated TWAP numbers.
		k.RefreshIntermediaryDelegationAmounts(ctx)
	}
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
		osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			_, err := k.dk.WithdrawDelegationRewards(cacheCtx, addr, valAddr)
			return err
		})

		// Send delegation rewards to gauges
		osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// Note! We only send the bond denom (osmo), to avoid attack vectors where people
			// send many different denoms to the intermediary account, and make a resource exhaustion attack on end block.
			bondDenom := k.sk.BondDenom(cacheCtx)
			balance := k.bk.GetBalance(cacheCtx, addr, bondDenom)
			return k.ik.AddToGaugeRewards(cacheCtx, addr, sdk.Coins{balance}, acc.GaugeId)
		})
	}
}

func (k Keeper) updateEpochTwap(ctx sdk.Context, asset types.SuperfluidAsset, endedEpochNumber int64) error {
	if asset.AssetType == types.SuperfluidAssetTypeLPShare {
		// LP_token_Osmo_equivalent = OSMO_amount_on_pool / LP_token_supply
		poolId := gammtypes.MustGetPoolIdFromShareDenom(asset.Denom)
		pool, err := k.gk.GetPool(ctx, poolId)
		if err != nil {
			// Pool has been unexpectedly deleted
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		// get OSMO amount
		bondDenom := k.sk.BondDenom(ctx)
		osmoPoolAsset, err := pool.GetPoolAsset(bondDenom)
		if err != nil {
			// Pool has unexpectedly removed Osmo from its assets.
			k.Logger(ctx).Error(err.Error())
			k.BeginUnwindSuperfluidAsset(ctx, 0, asset)
			return err
		}

		twap := k.calculateOsmoBackingPerShare(pool, osmoPoolAsset)
		beginningEpochNumber := endedEpochNumber + 1
		k.SetEpochOsmoEquivalentTWAP(ctx, beginningEpochNumber, asset.Denom, twap)
	} else if asset.AssetType == types.SuperfluidAssetTypeNative {
		// TODO: Consider deleting superfluid asset type native
		k.Logger(ctx).Error("unsupported superfluid asset type")
		return errors.New("SuperfluidAssetTypeNative is unspported")
	}
	return nil
}
