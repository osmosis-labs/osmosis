package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) { }

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	addrNonNativeFee := k.accountKeeper.GetModuleAddress(txfeestypes.NonNativeFeeCollectorName)
	nonNativeFeeAccountBalances := k.bankKeeper.GetAllBalances(ctx, addrNonNativeFee)
	baseDenom, _ := k.GetBaseDenom(ctx)

	for _, coin := range nonNativeFeeAccountBalances {
		if coin.Denom == baseDenom {
			continue
		} else {
		
			feetoken, _ := k.GetFeeToken(ctx, coin.Denom)

			k.gammKeeper.SwapExactAmountIn(ctx, addrNonNativeFee, feetoken.PoolID, coin, baseDenom, sdk.ZeroInt())
		}
	}

	nonNativeFeeAccountBalances = k.bankKeeper.GetAllBalances(ctx, addrNonNativeFee)
	
	k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.NonNativeFeeCollectorName, txfeestypes.FeeCollectorName, nonNativeFeeAccountBalances)
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
