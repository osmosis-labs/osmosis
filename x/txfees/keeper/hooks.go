package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	txfeestypes "github.com/osmosis-labs/osmosis/v23/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

var (
	zeroDec = osmomath.ZeroDec()
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

	nonNativefeeTokenCollectorAddress := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeTxFeeCollectorName)

	// Non-native fee token collector for staking rewards get swapped entirely into base denom.
	k.swapNonNativeFeeToDenom(ctx, defaultFeesDenom, nonNativefeeTokenCollectorAddress, k.bankKeeper.GetAllBalances(ctx, nonNativefeeTokenCollectorAddress))

	// Now that the rewards have been swapped, transfer any base denom existing in the non-native tx fee collector to the auth fee token collector (indirectly distributing to stakers)
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativefeeTokenCollectorAddress, defaultFeesDenom))
	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeTxFeeCollectorName, authtypes.FeeCollectorName, baseDenomCoins)
		return err
	})

	// Distribute and track the taker fees.
	k.calculateDistributeAndTrackTakerFees(ctx, defaultFeesDenom)

	return nil
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

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

func (k Keeper) calculateDistributeAndTrackTakerFees(ctx sdk.Context, defaultFeesDenom string) {
	// First deal with the native tokens in the taker fee collector.
	takerFeeModuleAccountName := txfeestypes.TakerFeeCollectorName
	takerFeeModuleAccount := k.accountKeeper.GetModuleAddress(takerFeeModuleAccountName)
	osmoFromFeeModuleAccount := k.bankKeeper.GetBalance(ctx, takerFeeModuleAccount, defaultFeesDenom)
	poolManagerParams := k.poolManager.GetParams(ctx)
	takerFeeParams := poolManagerParams.TakerFeeParams
	osmoTakerFeeDistribution := takerFeeParams.OsmoTakerFeeDistribution

	// Community Pool:
	if osmoTakerFeeDistribution.CommunityPool.GT(zeroDec) && osmoFromFeeModuleAccount.Amount.GT(osmomath.ZeroInt()) {
		// Osmo community pool funds are a direct send to the community pool.
		osmoTakerFeeToCommunityPoolDec := osmoFromFeeModuleAccount.Amount.ToLegacyDec().Mul(osmoTakerFeeDistribution.CommunityPool)
		osmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(defaultFeesDenom, osmoTakerFeeToCommunityPoolDec.TruncateInt())
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(osmoTakerFeeToCommunityPoolCoin), takerFeeModuleAccount)
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, osmoTakerFeeToCommunityPoolCoin.Denom, osmoTakerFeeToCommunityPoolCoin.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for community pool by denom", "error", err)
			}
			return err
		})
		osmoFromFeeModuleAccount = osmoFromFeeModuleAccount.Sub(osmoTakerFeeToCommunityPoolCoin)
	}

	// Staking Rewards:
	if osmoTakerFeeDistribution.StakingRewards.GT(zeroDec) && osmoFromFeeModuleAccount.Amount.GT(osmomath.ZeroInt()) {
		// Osmo staking rewards funds are a direct send to the auth fee token collector (indirectly distributing to stakers)
		osmoTakerFeeToStakingRewardsCoin := sdk.NewCoin(defaultFeesDenom, osmoFromFeeModuleAccount.Amount)
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, takerFeeModuleAccount, authtypes.FeeCollectorName, sdk.NewCoins(osmoTakerFeeToStakingRewardsCoin))
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForStakersByDenom(ctx, osmoTakerFeeToStakingRewardsCoin.Denom, osmoTakerFeeToStakingRewardsCoin.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for stakers by denom", "error", err)
			}
			return err
		})
	}

	// Now, deal with the non-native tokens in the taker fee collector.
	takerFeeModuleAccountCoins := k.bankKeeper.GetAllBalances(ctx, takerFeeModuleAccount)
	nonOsmoTakerFeeDistribution := takerFeeParams.NonOsmoTakerFeeDistribution
	authorizedQuoteDenoms := poolManagerParams.AuthorizedQuoteDenoms

	nonOsmoForStaking := sdk.NewCoins()
	nonOsmoForCommunityPool := sdk.NewCoins()

	// Loop through all remaining tokens in the taker fee module account.
	for _, takerFeeCoin := range takerFeeModuleAccountCoins {
		// Community Pool:
		if nonOsmoTakerFeeDistribution.CommunityPool.GT(zeroDec) && takerFeeCoin.Amount.GT(osmomath.ZeroInt()) {
			denomIsWhitelisted := isDenomWhitelisted(takerFeeCoin.Denom, authorizedQuoteDenoms)
			// If the non osmo denom is a whitelisted quote asset, we directly send to the community pool
			if denomIsWhitelisted {
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeCoin.Amount.ToLegacyDec().Mul(nonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(takerFeeCoin.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt())
				_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
					err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(nonOsmoTakerFeeToCommunityPoolCoin), takerFeeModuleAccount)
					if err == nil {
						takerFeeCoin.Amount = takerFeeCoin.Amount.Sub(nonOsmoTakerFeeToCommunityPoolCoin.Amount)
					}
					trackerErr := k.poolManager.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, nonOsmoTakerFeeToCommunityPoolCoin.Denom, nonOsmoTakerFeeToCommunityPoolCoin.Amount)
					if trackerErr != nil {
						ctx.Logger().Error("Error updating taker fee tracker for community pool by denom", "error", err)
					}
					return err
				})
			} else {
				// If the non osmo denom is not a whitelisted asset, we track the assets here and later swap everything to the community pool denom.
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeCoin.Amount.ToLegacyDec().Mul(nonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoin := sdk.NewCoin(takerFeeCoin.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt())
				nonOsmoForCommunityPool = nonOsmoForCommunityPool.Add(nonOsmoTakerFeeToCommunityPoolCoin)
				takerFeeCoin.Amount = takerFeeCoin.Amount.Sub(nonOsmoTakerFeeToCommunityPoolCoin.Amount)
			}
		}

		// Staking Rewards:
		if nonOsmoTakerFeeDistribution.StakingRewards.GT(zeroDec) && takerFeeCoin.Amount.GT(osmomath.ZeroInt()) {
			// Track the non osmo assets designated for staking rewards here and later swap everything to the base denom.
			nonOsmoTakerFeeToStakingRewardsCoin := sdk.NewCoin(takerFeeCoin.Denom, takerFeeCoin.Amount)
			nonOsmoForStaking = nonOsmoForStaking.Add(nonOsmoTakerFeeToStakingRewardsCoin)
		}
	}

	// Swap the non-native, non-whitelisted taker fees slated for community pool into the denom specified in the pool manager params.
	denomToSwapTo := poolManagerParams.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo
	totalCoinOut := k.swapNonNativeFeeToDenom(ctx, denomToSwapTo, takerFeeModuleAccount, nonOsmoForCommunityPool)
	// Now that the non whitelisted assets have been swapped, fund the community pool with the denom we swapped to.
	if totalCoinOut.Amount.GT(osmomath.ZeroInt()) {
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			err := k.distributionKeeper.FundCommunityPool(ctx, sdk.NewCoins(totalCoinOut), takerFeeModuleAccount)
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx, totalCoinOut.Denom, totalCoinOut.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for community pool by denom", "error", err)
			}
			return err
		})
	}

	// Swap the taker fees slated for staking rewards into the base denom.
	totalCoinOut = k.swapNonNativeFeeToDenom(ctx, defaultFeesDenom, takerFeeModuleAccount, nonOsmoForStaking)
	if totalCoinOut.Amount.GT(osmomath.ZeroInt()) {
		// Now that the assets have been swapped, transfer any base denom existing in the taker fee module account to the auth fee collector module account (indirectly distributing to stakers)
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.TakerFeeCollectorName, authtypes.FeeCollectorName, sdk.NewCoins(totalCoinOut))
			trackerErr := k.poolManager.UpdateTakerFeeTrackerForStakersByDenom(ctx, totalCoinOut.Denom, totalCoinOut.Amount)
			if trackerErr != nil {
				ctx.Logger().Error("Error updating taker fee tracker for stakers by denom", "error", err)
			}
			return err
		})
	}
}

// swapNonNativeFeeToDenom swaps coins into the denomToSwapTo from the given fee collector address.
// If an error in swap occurs for a given denom, it will be silently skipped.
// CONTRACT: a pool must exist between each denom in the balance and denomToSwapTo. If doesn't exist. Silently skip swap.
// CONTRACT: protorev must be configured to have a pool for the given denom pair. Otherwise, the denom will be skipped.
func (k Keeper) swapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress, coinsToSwap sdk.Coins) sdk.Coin {
	totalCoinOut := sdk.NewCoin(denomToSwapTo, osmomath.ZeroInt())
	for _, coin := range coinsToSwap {
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
			// The pool route either doesn't exist or is disabled in protorev.
			// It will just accrue in the non-native fee collector account.
			// Skip this denom and move on to the next one.
			continue
		}

		// Do the swap of this fee token denom to base denom.
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// We allow full slippage. There's not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minAmountOut := osmomath.ZeroInt()

			// We swap without charging a taker fee / sending to the non native fee collector, since these are funds that
			// are accruing from the taker fee itself.
			amtOutInt, err := k.poolManager.SwapExactAmountInNoTakerFee(cacheCtx, feeCollectorAddress, poolId, coin, denomToSwapTo, minAmountOut)
			if err == nil {
				totalCoinOut = totalCoinOut.Add(sdk.NewCoin(denomToSwapTo, amtOutInt))
			}
			return err
		})
	}
	return totalCoinOut
}

// isDenomWhitelisted checks if the denom provided exists in the list of authorized quote denoms.
// If it does, it returns true, otherwise false.
func isDenomWhitelisted(denom string, authorizedQuoteDenoms []string) bool {
	for _, authorizedQuoteDenom := range authorizedQuoteDenoms {
		if denom == authorizedQuoteDenom {
			return true
		}
	}
	return false
}
