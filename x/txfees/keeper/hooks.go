package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	txfeestypes "github.com/osmosis-labs/osmosis/v18/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	nonNativeFeeAddr := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeFeeCollectorName)
	baseDenom, _ := k.GetBaseDenom(ctx)
	feeTokens := k.GetFeeTokens(ctx)

	for _, feetoken := range feeTokens {
		if feetoken.Denom == baseDenom {
			continue
		}
		coinBalance := k.bankKeeper.GetBalance(ctx, nonNativeFeeAddr, feetoken.Denom)
		if coinBalance.Amount.IsZero() {
			continue
		}

		// Do the swap of this fee token denom to base denom.
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// We allow full slippage. Theres not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minAmountOut := sdk.ZeroInt()
			_, err := k.poolManager.SwapExactAmountIn(cacheCtx, nonNativeFeeAddr, feetoken.PoolID, coinBalance, baseDenom, minAmountOut)
			return err
		})
	}

	// Get all of the txfee payout denom in the module account
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeFeeAddr, baseDenom))

	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeFeeCollectorName, txfeestypes.FeeCollectorName, baseDenomCoins)
		return err
	})

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
<<<<<<< HEAD
=======

// swapNonNativeFeeToDenom swaps the given non-native fees into the given denom from the given fee collector address.
// If an error in swap occurs for a given denom, it will be silently skipped.
// CONTRACT: a pool must exist between each denom in the balance and denomToSwapTo. If doesn't exist. Silently skip swap.
// CONTRACT: protorev must be configured to have a pool for the given denom pair. Otherwise, the denom will be skipped.
func (k Keeper) swapNonNativeFeeToDenom(ctx sdk.Context, denomToSwapTo string, feeCollectorAddress sdk.AccAddress) {
	feeCollectorBalance := k.bankKeeper.GetAllBalances(ctx, feeCollectorAddress)

	for _, coin := range feeCollectorBalance {
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
				// The pool route either doesn't exist or is disabled in protorev.
				// It will just accrue in the non-native fee collector account.
				// Skip this denom and move on to the next one.
				continue
			}
		}

		// Do the swap of this fee token denom to base denom.
		_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			// We allow full slippage. Theres not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minAmountOut := osmomath.ZeroInt()

			// We swap without charging a taker fee / sending to the non native fee collector, since these are funds that
			// are accruing from the taker fee itself.
			_, err := k.poolManager.SwapExactAmountInNoTakerFee(cacheCtx, feeCollectorAddress, poolId, coin, denomToSwapTo, minAmountOut)
			return err
		})
	}
}
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
