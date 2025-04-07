package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-metrics"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	epochstypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd at the end of each epoch, swap all non-OSMO fees into the desired denom and send either to fee collector or community pool.
// Staking fee collector for staking rewards.
// - All non-native rewards that have a pool with liquidity and a link set in protorev get swapped to native denom
// - All resulting native tokens get sent to the fee collector.
// - Any non-native tokens that did not have associated pool stay in the balance of staking fee collector.
// Community pool fee collector.
// - All non-native rewards that have a pool with liquidity and a link set in protorev get swapped to a denom configured by parameter.
// - All resulting parameter denom tokens get sent to the community pool.
// - Any non-native tokens that did not have associated pool stay in the balance of community pool fee collector.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != k.GetParams(ctx).SwapFeesEpochIdentifier {
		return nil
	}

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

		// Do the swap of this fee token denom to base denom.
		err := osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
			swapResult, err := k.marketKeeper.Swap(cacheCtx, feeCollectorAddress, feeCollectorAddress, coin, denomToSwapTo)
			if err != nil {
				coinsNotSwapped = append(coinsNotSwapped, coin.String())
			} else {
				totalCoinOut = totalCoinOut.Add(sdk.NewCoin(denomToSwapTo, swapResult.SwapCoin.Amount))
			}
			return err
		})
		if err != nil {
			telemetry.IncrCounterWithLabels([]string{txfeestypes.TakerFeeSwapFailedMetricName}, 1, []metrics.Label{
				{
					Name:  "coin_in",
					Value: coin.String(),
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
