package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	addrNonNativeFee := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeFeeCollectorName)
	nonNativeFeeAccountBalances := k.bankKeeper.GetAllBalances(ctx, addrNonNativeFee)
	baseDenom, _ := k.GetBaseDenom(ctx)

	for _, coin := range nonNativeFeeAccountBalances {
		if coin.Denom == baseDenom {
			continue
		} else {

			feetoken, err := k.GetFeeToken(ctx, coin.Denom)
			if err != nil {
				return err
			}

			// We allow full slippage. Theres not really an effective way to bound slippage until TWAP's land,
			// but even then the point is a bit moot.
			// The only thing that could be done is a costly griefing attack to reduce the amount of osmo given as tx fees.
			// However the idea of the txfees FeeToken gating is that the pool is sufficiently liquid for that base token.
			minTokenOut := sdk.ZeroInt()
			_, err = k.gammKeeper.SwapExactAmountIn(ctx, addrNonNativeFee, feetoken.PoolID, coin, baseDenom, minTokenOut)
			if err != nil {
				return err
			}
		}
	}

	// Get all of the txfee payout denom in the module account
	nonNativeFeeAccountBaseDenomBalance := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, addrNonNativeFee, baseDenom))

	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeFeeCollectorName, txfeestypes.FeeCollectorName, nonNativeFeeAccountBaseDenomBalance)
	if err != nil {
		return err
	}

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

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
