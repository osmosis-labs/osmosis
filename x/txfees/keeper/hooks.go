package keeper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-metrics"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v30/x/poolmanager/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	txfeestypes "github.com/osmosis-labs/osmosis/v30/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

var zeroDec = osmomath.ZeroDec()

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

	nonNativefeeTokenCollectorAddress := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeTxFeeCollectorName)

	// Non-native fee token collector for staking rewards get swapped entirely into base denom.
	k.swapNonNativeFeeToDenom(ctx, defaultFeesDenom, nonNativefeeTokenCollectorAddress)

	// Now that the rewards have been swapped, transfer any base denom existing in the non-native tx fee collector to the auth fee token collector (indirectly distributing to stakers)
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativefeeTokenCollectorAddress, defaultFeesDenom))
	err := osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeTxFeeCollectorName, authtypes.FeeCollectorName, baseDenomCoins)
		return err
	})
	if err != nil {
		incTelementryCounter(txfeestypes.TakerFeeFailedNativeRewardUpdateMetricName, baseDenomCoins.String(), err.Error())
	}

	// Send skimmed taker fees to respective fee collectors.
	k.clearTakerFeeShareAccumulators(ctx)

	// Distribute and track the taker fees.
	k.calculateDistributeAndTrackTakerFees(ctx, defaultFeesDenom)

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

// calculateDistributeAndTrackTakerFees calculates the taker fees and distributes them to the community pool, burn address, and stakers.
// The following is the logic for the taker fee distribution:
//
// - OSMO taker fees
//   - For Community Pool: Sent directly to community pool
//   - For Burn: Sent to the null address (osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030) to effectively burn the tokens
//   - For Stakers: Sent directly to auth module account, which distributes it to stakers
//
// - Non native taker fees
//   - For Community Pool: Sent to `non_native_fee_collector_community_pool` module account, swapped to `CommunityPoolDenomToSwapNonWhitelistedAssetsTo`, then sent to community pool
//   - For Stakers: Sent to `non_native_fee_collector_stakers` module account, swapped to OSMO, then sent to auth module account, which distributes it to stakers
//   - The sub-module accounts here are used so that, if a swap fails, the tokens that fail to swap are not grouped back into the wrong taker fee category in the next epoch
func (k Keeper) calculateDistributeAndTrackTakerFees(ctx sdk.Context, defaultFeesDenom string) {
	// First deal with the native tokens in the taker fee collector.
	takerFeeModuleAccount := k.accountKeeper.GetModuleAddress(txfeestypes.TakerFeeCollectorName)
	osmoFromTakerFeeModuleAccount := k.bankKeeper.GetBalance(ctx, takerFeeModuleAccount, defaultFeesDenom)

	poolManagerParams := k.poolManager.GetParams(ctx)
	takerFeeParams := poolManagerParams.TakerFeeParams
	osmoTakerFeeDistribution := takerFeeParams.OsmoTakerFeeDistribution

	// Community Pool:
	if osmoTakerFeeDistribution.CommunityPool.GT(zeroDec) && osmoFromTakerFeeModuleAccount.Amount.GT(osmomath.ZeroInt()) {
		// Osmo community pool funds are a direct send to the community pool.
		osmoTakerFeeToCommunityPoolDec := osmoFromTakerFeeModuleAccount.Amount.ToLegacyDec().Mul(osmoTakerFeeDistribution.CommunityPool)
		osmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(defaultFeesDenom, osmoTakerFeeToCommunityPoolDec.TruncateInt())
		applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
			err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(osmoTakerFeeToCommunityPoolCoin), takerFeeModuleAccount)
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, osmoTakerFeeToCommunityPoolCoin.Denom, osmoTakerFeeToCommunityPoolCoin.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for community pool by denom", "error", err)
			}
			return err
		}, txfeestypes.TakerFeeFailedCommunityPoolUpdateMetricName, osmoTakerFeeToCommunityPoolCoin)
	}

	// Burn:
	if osmoTakerFeeDistribution.Burn.GT(zeroDec) && osmoFromTakerFeeModuleAccount.Amount.GT(osmomath.ZeroInt()) {
		// Calculate burn amount and send to null address to effectively burn the tokens
		osmoTakerFeeToBurnDec := osmoFromTakerFeeModuleAccount.Amount.ToLegacyDec().Mul(osmoTakerFeeDistribution.Burn)
		osmoTakerFeeToBurnCoin := sdk.NewCoin(defaultFeesDenom, osmoTakerFeeToBurnDec.TruncateInt())

		applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
			err := k.bankKeeper.SendCoins(ctx, takerFeeModuleAccount, txfeestypes.DefaultNullAddress, sdk.NewCoins(osmoTakerFeeToBurnCoin))
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForBurnByDenom(ctx, osmoTakerFeeToBurnCoin.Denom, osmoTakerFeeToBurnCoin.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for burn by denom", "error", trackerErr)
			}
			return err
		}, txfeestypes.TakerFeeFailedBurnUpdateMetricName, osmoTakerFeeToBurnCoin)
	}

	// Staking Rewards:
	if osmoTakerFeeDistribution.StakingRewards.GT(zeroDec) && osmoFromTakerFeeModuleAccount.Amount.GT(osmomath.ZeroInt()) {
		// Osmo staking rewards funds are a direct send to the auth fee token collector (indirectly distributing to stakers)
		osmoTakerFeeToStakingRewardsDec := osmoFromTakerFeeModuleAccount.Amount.ToLegacyDec().Mul(osmoTakerFeeDistribution.StakingRewards)
		osmoTakerFeeToStakingRewardsCoin := sdk.NewCoin(defaultFeesDenom, osmoTakerFeeToStakingRewardsDec.TruncateInt())
		applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, takerFeeModuleAccount, authtypes.FeeCollectorName, sdk.NewCoins(osmoTakerFeeToStakingRewardsCoin))
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForStakersByDenom(ctx, osmoTakerFeeToStakingRewardsCoin.Denom, osmoTakerFeeToStakingRewardsCoin.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for stakers by denom", "error", err)
			}
			return err
		}, txfeestypes.TakerFeeFailedNativeRewardUpdateMetricName, osmoTakerFeeToStakingRewardsCoin)
	}

	// Now, deal with the non-native tokens in the taker fee collector.
	takerFeeModuleAccountCoins := k.bankKeeper.GetAllBalances(ctx, takerFeeModuleAccount)
	nonOsmoTakerFeeDistribution := takerFeeParams.NonOsmoTakerFeeDistribution
	communityPoolDenomWhitelist := poolManagerParams.TakerFeeParams.CommunityPoolDenomWhitelist

	nonOsmoForCommunityPool := sdk.NewCoins()
	nonOsmoForBurn := sdk.NewCoins()

	// Loop through all remaining tokens in the taker fee module account.
	for _, takerFeeCoin := range takerFeeModuleAccountCoins {
		// Store original amount to calculate percentages from the total, not from remaining amounts
		originalAmount := takerFeeCoin.Amount

		// Community Pool:
		if nonOsmoTakerFeeDistribution.CommunityPool.GT(zeroDec) && originalAmount.GT(osmomath.ZeroInt()) {
			denomIsWhitelisted := isDenomWhitelisted(takerFeeCoin.Denom, communityPoolDenomWhitelist)
			// If the non osmo denom is a whitelisted quote asset, we directly send to the community pool
			if denomIsWhitelisted {
				nonOsmoTakerFeeToCommunityPoolDec := originalAmount.ToLegacyDec().Mul(nonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(takerFeeCoin.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt())
				applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
					err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(nonOsmoTakerFeeToCommunityPoolCoin), takerFeeModuleAccount)
					if err == nil {
						takerFeeCoin.Amount = takerFeeCoin.Amount.Sub(nonOsmoTakerFeeToCommunityPoolCoin.Amount)
					}
					trackerErr := k.poolManager.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, nonOsmoTakerFeeToCommunityPoolCoin.Denom, nonOsmoTakerFeeToCommunityPoolCoin.Amount)
					if trackerErr != nil {
						ctx.Logger().Error("Error updating taker fee tracker for community pool by denom", "error", err)
					}
					return err
				}, txfeestypes.TakerFeeFailedCommunityPoolUpdateMetricName, nonOsmoTakerFeeToCommunityPoolCoin)
			} else {
				// If the non osmo denom is not a whitelisted asset, we track the assets here and later swap everything to the community pool denom.
				nonOsmoTakerFeeToCommunityPoolDec := originalAmount.ToLegacyDec().Mul(nonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(takerFeeCoin.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt())
				nonOsmoForCommunityPool = nonOsmoForCommunityPool.Add(nonOsmoTakerFeeToCommunityPoolCoin)
				takerFeeCoin.Amount = takerFeeCoin.Amount.Sub(nonOsmoTakerFeeToCommunityPoolCoin.Amount)
			}
		}

		// Burn: Calculate from original amount, not remaining amount
		if nonOsmoTakerFeeDistribution.Burn.GT(zeroDec) && originalAmount.GT(osmomath.ZeroInt()) {
			// For burn, we don't care about whitelist - all non-OSMO tokens designated for burn should be swapped to OSMO and burned
			nonOsmoTakerFeeToBurnDec := originalAmount.ToLegacyDec().Mul(nonOsmoTakerFeeDistribution.Burn)
			nonOsmoTakerFeeToBurnCoin := sdk.NewCoin(takerFeeCoin.Denom, nonOsmoTakerFeeToBurnDec.TruncateInt())
			nonOsmoForBurn = nonOsmoForBurn.Add(nonOsmoTakerFeeToBurnCoin)
			takerFeeCoin.Amount = takerFeeCoin.Amount.Sub(nonOsmoTakerFeeToBurnCoin.Amount)
		}

		// We don't need to calculate the staking rewards for non native taker fees, since it ends up being whatever is left over in the taker fee module account!
	}

	// Send the non-native, non-whitelisted taker fees slated for the community pool to the taker fee community pool module account.
	// We do this in the event that the swap fails, we can still track the amount of non-native, non-whitelisted taker fees that were intended for the community pool.
	applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, takerFeeModuleAccount, txfeestypes.TakerFeeCommunityPoolName, nonOsmoForCommunityPool)
		return err
	}, txfeestypes.TakerFeeFailedCommunityPoolUpdateMetricName, nonOsmoForCommunityPool)

	// Swap the non-native, non-whitelisted taker fees slated for community pool into the denom specified in the pool manager params.
	takerFeeCommunityPoolModuleAccount := k.accountKeeper.GetModuleAddress(txfeestypes.TakerFeeCommunityPoolName)
	denomToSwapTo := poolManagerParams.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo
	totalCoinOut := k.swapNonNativeFeeToDenom(ctx, denomToSwapTo, takerFeeCommunityPoolModuleAccount)
	// Now that the non whitelisted assets have been swapped, fund the community pool with the denom we swapped to.
	if totalCoinOut.Amount.GT(osmomath.ZeroInt()) {
		applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
			err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(totalCoinOut), takerFeeCommunityPoolModuleAccount)
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, totalCoinOut.Denom, totalCoinOut.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for community pool by denom", "error", err)
			}
			return err
		}, txfeestypes.TakerFeeFailedCommunityPoolUpdateMetricName, totalCoinOut)
	}

	// Send the non-native taker fees slated for burn to the taker fee burn module account.
	// We do this in the event that the swap fails, we can still track the amount of non-native taker fees that were intended for burn.
	applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, takerFeeModuleAccount, txfeestypes.TakerFeeBurnName, nonOsmoForBurn)
		return err
	}, txfeestypes.TakerFeeFailedBurnUpdateMetricName, nonOsmoForBurn)

	// Swap the non-native taker fees slated for burn into OSMO.
	takerFeeBurnModuleAccount := k.accountKeeper.GetModuleAddress(txfeestypes.TakerFeeBurnName)
	totalCoinOutForBurn := k.swapNonNativeFeeToDenom(ctx, defaultFeesDenom, takerFeeBurnModuleAccount)
	// Now that the non-native assets have been swapped to OSMO, burn them by sending to the null address.
	if totalCoinOutForBurn.Amount.GT(osmomath.ZeroInt()) {
		applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
			err := k.bankKeeper.SendCoins(ctx, takerFeeBurnModuleAccount, txfeestypes.DefaultNullAddress, sdk.NewCoins(totalCoinOutForBurn))
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForBurnByDenom(ctx, totalCoinOutForBurn.Denom, totalCoinOutForBurn.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for burn by denom", "error", err)
			}
			return err
		}, txfeestypes.TakerFeeFailedBurnUpdateMetricName, totalCoinOutForBurn)
	}

	// Send the non-native taker fees slated for stakers to the taker fee staking module account.
	// We do this in the event that the swap fails, we can still track the amount of non-native taker fees that were intended for stakers.
	var remainingTakerFeeModuleAccBal sdk.Coins
	applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
		remainingTakerFeeModuleAccBal = k.bankKeeper.GetAllBalances(ctx, takerFeeModuleAccount)
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, takerFeeModuleAccount, txfeestypes.TakerFeeStakersName, remainingTakerFeeModuleAccBal)
		return err
	}, txfeestypes.TakerFeeFailedNativeRewardUpdateMetricName, remainingTakerFeeModuleAccBal)

	// Swap the taker fees slated for staking rewards into the base denom.
	takerFeeStakersModuleAccount := k.accountKeeper.GetModuleAddress(txfeestypes.TakerFeeStakersName)
	totalCoinOut = k.swapNonNativeFeeToDenom(ctx, defaultFeesDenom, takerFeeStakersModuleAccount)
	if totalCoinOut.Amount.GT(osmomath.ZeroInt()) {
		// Now that the assets have been swapped, transfer any base denom existing in the taker fee module account to the auth fee collector module account (indirectly distributing to stakers)
		applyFuncIfNoErrorAndLog(ctx, func(cacheCtx sdk.Context) error {
			err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.TakerFeeStakersName, authtypes.FeeCollectorName, sdk.NewCoins(totalCoinOut))
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForStakersByDenom(ctx, totalCoinOut.Denom, totalCoinOut.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for stakers by denom", "error", err)
			}
			return err
		}, txfeestypes.TakerFeeFailedNativeRewardUpdateMetricName, totalCoinOut)
	}
}

// swapNonNativeFeeToDenom swaps coins into the denomToSwapTo from the given fee collector address.
// If an error in swap occurs for a given denom, it will be silently skipped.
// CONTRACT: a pool must exist between each denom in the balance and denomToSwapTo. If doesn't exist. Silently skip swap.
// CONTRACT: protorev must be configured to have a pool for the given denom pair. Otherwise, the denom will be skipped.
func (k Keeper) swapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress) sdk.Coin {
	coinsToSwap := k.bankKeeper.GetAllBalances(ctx, feeCollectorAddress)
	totalCoinOut := sdk.NewCoin(denomToSwapTo, osmomath.ZeroInt())
	coinsNotSwapped := []string{}

	for _, coin := range coinsToSwap {
		if coin.Denom == denomToSwapTo {
			continue
		}

		// Skip coins with zero amount to avoid "token amount must be positive" error
		if coin.Amount.IsZero() {
			continue
		}

		var route []poolmanagertypes.SwapAmountInRoute
		var routeType string // For telemetry tracking

		// First, try to find a direct single hop route
		// Search for the denom pair route via the protorev store.
		// Since OSMO is one of the protorev denoms, many of the routes will exist in this store.
		poolId, err := k.protorevKeeper.GetPoolForDenomPairNoOrder(ctx, denomToSwapTo, coin.Denom)
		if err == nil {
			// Direct route found - build single hop route
			route = []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        poolId,
					TokenOutDenom: denomToSwapTo,
				},
			}
			routeType = "single_hop"
		} else {
			// No direct route found, try 2-hop routes using intermediary denoms as intermediaries
			params := k.GetParams(ctx)
			intermediaryDenoms := params.FeeSwapIntermediaryDenomList

			// Try each intermediary denom until we find a valid 2-hop route
			routeFound := false
			for _, intermediaryDenom := range intermediaryDenoms {
				if intermediaryDenom == coin.Denom || intermediaryDenom == denomToSwapTo {
					continue // Skip if same as input or output denom
				}

				twoHopRoute, routeErr := k.build2HopsRoute(ctx, coin.Denom, intermediaryDenom, denomToSwapTo)
				if routeErr == nil {
					route = twoHopRoute
					routeType = fmt.Sprintf("two_hop_via_%s", intermediaryDenom)
					routeFound = true
					break
				}
			}

			// If no 2-hop route found, skip this coin
			if !routeFound {
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
						Value: fmt.Sprintf("no single hop route: %v, no 2-hop routes found via intermediary denoms", err),
					},
				})

				// The route either doesn't exist or is disabled in protorev.
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
			amtOutInt, err := k.poolManager.RouteExactAmountInNoTakerFee(cacheCtx, feeCollectorAddress, route, coin, minAmountOut)
			if err != nil {
				// Build route description for logging
				routeDesc := ""
				if len(route) == 1 {
					routeDesc = fmt.Sprintf("pool %d", route[0].PoolId)
				} else {
					poolIds := make([]string, len(route))
					for i, r := range route {
						poolIds[i] = strconv.FormatUint(r.PoolId, 10)
					}
					routeDesc = fmt.Sprintf("pools %s", strings.Join(poolIds, "->"))
				}
				coinsNotSwapped = append(coinsNotSwapped, fmt.Sprintf("%s via %s", coin.String(), routeDesc))
			} else {
				totalCoinOut = totalCoinOut.Add(sdk.NewCoin(denomToSwapTo, amtOutInt))
			}
			return err
		})
		if err != nil {
			// Build route info for telemetry
			routePoolIds := ""
			if len(route) > 0 {
				poolIds := make([]string, len(route))
				for i, r := range route {
					poolIds[i] = strconv.FormatUint(r.PoolId, 10)
				}
				routePoolIds = strings.Join(poolIds, "->")
			}

			telemetry.IncrCounterWithLabels([]string{txfeestypes.TakerFeeSwapFailedMetricName}, 1, []metrics.Label{
				{
					Name:  "coin_in",
					Value: coin.String(),
				},
				{
					Name:  "route_pools",
					Value: routePoolIds,
				},
				{
					Name:  "route_type",
					Value: routeType,
				},
				{
					Name:  "err",
					Value: err.Error(),
				},
			})
		}
	}
	if len(coinsNotSwapped) > 0 {
		ctx.Logger().Info(fmt.Sprintf("The following non-native tokens were not swapped (see debug logs for further details): %s", coinsNotSwapped))
	}

	return totalCoinOut
}

// build2HopsRoute builds a 2-hops swap route given an intermediary denom and target denom.
// It first finds a pool from the input coin to the intermediary denom, then from intermediary to target.
// Returns the complete route or an error if any pools are missing.
func (k Keeper) build2HopsRoute(ctx sdk.Context, inputDenom, intermediaryDenom, denomToSwapTo string) ([]poolmanagertypes.SwapAmountInRoute, error) {
	// Find pool for first hop: inputDenom -> intermediaryDenom
	poolId1, err := k.protorevKeeper.GetPoolForDenomPairNoOrder(ctx, inputDenom, intermediaryDenom)
	if err != nil {
		return nil, fmt.Errorf("no pool found for first hop %s -> %s: %w", inputDenom, intermediaryDenom, err)
	}

	// Find pool for second hop: intermediaryDenom -> denomToSwapTo
	poolId2, err := k.protorevKeeper.GetPoolForDenomPairNoOrder(ctx, intermediaryDenom, denomToSwapTo)
	if err != nil {
		return nil, fmt.Errorf("no pool found for second hop %s -> %s: %w", intermediaryDenom, denomToSwapTo, err)
	}

	// Build the 2-hops route
	route := []poolmanagertypes.SwapAmountInRoute{
		{
			PoolId:        poolId1,
			TokenOutDenom: intermediaryDenom,
		},
		{
			PoolId:        poolId2,
			TokenOutDenom: denomToSwapTo,
		},
	}

	return route, nil
}

// clearTakerFeeShareAccumulators retrieves all taker fee share accumulators and sends the coins to the respective addresses.
// This is used to clear the taker fee share accumulators at the end of each epoch, prior to distributing the rest of the taker fees.
func (k Keeper) clearTakerFeeShareAccumulators(ctx sdk.Context) {
	takerFeeSkimAccumulators, err := k.poolManager.GetAllTakerFeeShareAccumulators(ctx)
	if err != nil {
		ctx.Logger().Error("Error getting all taker fee share accumulators", "error", err)
		return
	}
	for _, takerFeeSkimAccumulator := range takerFeeSkimAccumulators {
		takerFeeShareAgreement, found := k.poolManager.GetTakerFeeShareAgreementFromDenomNoCache(ctx, takerFeeSkimAccumulator.Denom)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("Error getting taker fee share from denom: %s", takerFeeSkimAccumulator.Denom))
			continue
		}
		skimAddress := sdk.MustAccAddressFromBech32(takerFeeShareAgreement.SkimAddress)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, txfeestypes.TakerFeeCollectorName, skimAddress, takerFeeSkimAccumulator.SkimmedTakerFees)
		if err != nil {
			ctx.Logger().Error("Error sending coins from module to account", "error", err)
			continue
		}
		// If no errors occurred, delete every denom accumulator for the specified taker fee share denom.
		k.poolManager.DeleteAllTakerFeeShareAccumulatorsForTakerFeeShareDenom(ctx, takerFeeSkimAccumulator.Denom)
	}
}

// isDenomWhitelisted checks if the denom provided exists in the list of community pool denom whitelist.
// If it does, it returns true, otherwise false.
func isDenomWhitelisted(denom string, communityPoolDenomWhitelist []string) bool {
	for _, communityPoolDenom := range communityPoolDenomWhitelist {
		if denom == communityPoolDenom {
			return true
		}
	}
	return false
}

func applyFuncIfNoErrorAndLog[S fmt.Stringer](ctx sdk.Context, f func(sdk.Context) error, metricName string, coins S) {
	err := osmoutils.ApplyFuncIfNoError(ctx, f)
	if err != nil {
		incTelementryCounter(metricName, coins.String(), err.Error())
	}
}

// incTelementryCounter is a helper function to increment a telemetry counter with the given label.
func incTelementryCounter(labelName, coins, err string) {
	telemetry.IncrCounterWithLabels([]string{labelName}, 1, []metrics.Label{
		{
			Name:  "coins",
			Value: coins,
		},
		{
			Name:  "err",
			Value: err,
		},
	})
}
