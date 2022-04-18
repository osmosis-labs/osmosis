package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
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
			_, err := k.gammKeeper.SwapExactAmountIn(cacheCtx, nonNativeFeeAddr, feetoken.PoolID, coinBalance, baseDenom, minAmountOut)
			return err
		})
	}

	// Get all of the txfee payout denom in the module account
	baseDenomCoins := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, nonNativeFeeAddr, baseDenom))

	_ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeFeeCollectorName, txfeestypes.FeeCollectorName, baseDenomCoins)
		return err
	})
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

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
