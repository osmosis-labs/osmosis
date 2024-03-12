package keeper

import (
	"strconv"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	txfeestypes "github.com/osmosis-labs/osmosis/v23/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// at the end of each epoch, swap all non-OSMO fees into the desired denom and send either to fee collector or community pool.
// Staking fee collector for staking rewards.
// - All non-native rewards that have a pool with liquidity and a link set in protorev get swapped to native denom
// - All resulting native tokens get sent to the fee collector.
// - Any non-native tokens that did not have associated pool stay in the balance of staking fee collector.
// Community pool fee collector.
// - All non-native rewards that have a pool with liquidity and a link set in protorev get swapped to a denom configured by parameter.
// - All resulting parameter denom tokens get sent to the community pool.
// - Any non-native tokens that did not have associated pool stay in the balance of community pool fee collector.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	defaultFeesDenom, _ := k.GetBaseDenom(ctx)

	nonNativeStakingCollectorAddress := k.accountKeeper.GetModuleAddress(txfeestypes.FeeCollectorForStakingRewardsName)

	// Non-native fee collector for staking rewards get swapped entirely into base denom.
	k.swapNonNativeFeeToDenom(ctx, defaultFeesDenom, nonNativeStakingCollectorAddress)

	// Now that the rewards have been swapped, transfer any base denom existing in the non-native fee collector to the fee collector (indirectly distributing to stakers)
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeStakingCollectorAddress, defaultFeesDenom))
	err := osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.FeeCollectorForStakingRewardsName, txfeestypes.FeeCollectorName, baseDenomCoins)
		return err
	})
	if err != nil {
		telemetry.IncrCounterWithLabels([]string{txfeestypes.TakerFeeFailedNativeRewardUpdateMetricName}, 1, []metrics.Label{
			{
				Name:  "coins",
				Value: baseDenomCoins.String(),
			},
			{
				Name:  "err",
				Value: err.Error(),
			},
		})
	}

	// Non-native fee collector for community pool get swapped entirely into denom specified in the pool manager params.
	poolManagerParams := k.poolManager.GetParams(ctx)
	denomToSwapTo := poolManagerParams.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo
	// Only non-whitelisted assets should exist here since we do direct community pool funds when calculating the taker fee if
	// the input is a whitelisted asset.
	nonNativeCommunityPoolCollectorAddress := k.accountKeeper.GetModuleAddress(txfeestypes.FeeCollectorForCommunityPoolName)
	k.swapNonNativeFeeToDenom(ctx, denomToSwapTo, nonNativeCommunityPoolCollectorAddress)

	// Now that the non whitelisted assets have been swapped, fund the community pool with the denom we swapped to.
	denomToSwapToCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeCommunityPoolCollectorAddress, denomToSwapTo))
	err = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.distributionKeeper.FundCommunityPool(ctx, denomToSwapToCoins, nonNativeCommunityPoolCollectorAddress)
		return err
	})
	if err != nil {
		telemetry.IncrCounterWithLabels([]string{txfeestypes.TakerFeeFailedCommunityPoolUpdateMetricName}, 1, []metrics.Label{
			{
				Name:  "coins",
				Value: denomToSwapToCoins.String(),
			},
			{
				Name:  "err",
				Value: err.Error(),
			},
		})
	}

	return nil
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return txfeestypes.ModuleName
}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

// swapNonNativeFeeToDenom swaps the given non-native fees into the given denom from the given fee collector address.
// If an error in swap occurs for a given denom, it will be silently skipped.
// CONTRACT: a pool must exist between each denom in the balance and denomToSwapTo. If doesn't exist. Silently skip swap.
// CONTRACT: protorev must be configured to have a pool for the given denom pair. Otherwise, the denom will be skipped.
<<<<<<< HEAD
func (k Keeper) swapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress) {
	feeCollectorBalance := k.bankKeeper.GetAllBalances(ctx, feeCollectorAddress)

	for _, coin := range feeCollectorBalance {
=======
func (k Keeper) swapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress) sdk.Coin {
	coinsToSwap := k.bankKeeper.GetAllBalances(ctx, feeCollectorAddress)
	totalCoinOut := sdk.NewCoin(denomToSwapTo, osmomath.ZeroInt())
	coinsNotSwapped := []string{}

	for _, coin := range coinsToSwap {
>>>>>>> 9269f6fa (chore: reduce more epoch log noise (#7660))
		if coin.Denom == denomToSwapTo {
			continue
		}

		// Search for the denom pair route via the protorev store.
		// Since OSMO is one of the protorev denoms, many of the routes will exist in this store.
		// There will be times when this store does not know about a route, but this is acceptable
		// since this will likely be a very small value of a relatively unknown token. If this begins
		// to accrue more value, we can always manually register the route and it will get swapped in
		// the next epoch.
		poolId, err := k.protorevKeeper.GetPoolForDenomPairNoOrder(ctx, denomToSwapTo, coin.Denom)
		if err != nil {
			if err != nil {
				telemetry.IncrCounterWithLabels([]string{txfeestypes.TakerFeeNoSkipRouteMetricName}, 1, []metrics.Label{
					{
						Name:  "base_denom",
						Value: denomToSwapTo,
					},
					{
						Name:  "match_denom",
						Value: coin.Denom,
					},
					{
						Name:  "err",
						Value: err.Error(),
					},
				})

				// The pool route either doesn't exist or is disabled in protorev.
				// It will just accrue in the non-native fee collector account.
				// Skip this denom and move on to the next one.
				continue
			}
		}

		// Do the swap of this fee token denom to base denom.
		err = osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
			// We allow full slippage. There's not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minAmountOut := osmomath.ZeroInt()

			// We swap without charging a taker fee / sending to the non native fee collector, since these are funds that
			// are accruing from the taker fee itself.
<<<<<<< HEAD
			_, err := k.poolManager.SwapExactAmountInNoTakerFee(cacheCtx, feeCollectorAddress, poolId, coin, denomToSwapTo, minAmountOut)
=======
			amtOutInt, err := k.poolManager.SwapExactAmountInNoTakerFee(cacheCtx, feeCollectorAddress, poolId, coin, denomToSwapTo, minAmountOut)
			if err != nil {
				coinsNotSwapped = append(coinsNotSwapped, fmt.Sprintf("%s via pool %v", coin.String(), poolId))
			} else {
				totalCoinOut = totalCoinOut.Add(sdk.NewCoin(denomToSwapTo, amtOutInt))
			}
>>>>>>> 9269f6fa (chore: reduce more epoch log noise (#7660))
			return err
		})
		if err != nil {
			telemetry.IncrCounterWithLabels([]string{txfeestypes.TakerFeeSwapFailedMetricName}, 1, []metrics.Label{
				{
					Name:  "coin_in",
					Value: coin.String(),
				},
				{
					Name:  "pool_id",
					Value: strconv.FormatUint(poolId, 10),
				},
				{
					Name:  "err",
					Value: err.Error(),
				},
			})
		}
		if len(coinsNotSwapped) > 0 {
			ctx.Logger().Info("The following non-native tokens were not swapped (see debug logs for further details): %s", coinsNotSwapped)
		}
	}
}
