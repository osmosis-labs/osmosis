package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {

	addrFoo := k.accountKeeper.GetModuleAddress(txfeestypes.FooCollectorName)

	fooAccountBalances := k.bankKeeper.GetAllBalances(ctx, addrFoo)

	baseDenom, _ := k.GetBaseDenom(ctx)

	for _, coin := range fooAccountBalances {

		if coin.Denom == baseDenom {
			continue
		} else {
		
			feetoken, _ := k.GetFeeToken(ctx, coin.Denom)

			k.gammKeeper.SwapExactAmountIn(ctx, addrFoo, feetoken.PoolID, coin, baseDenom, sdk.ZeroInt())
		}
	}
	
	k.bankKeeper.SendCoinsFromModuleToModule(ctx, txfeestypes.FooCollectorName, txfeestypes.FeeCollectorName, fooAccountBalances)

	// Should events be emitted at the end here?
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
// Don't do anything pre epoch start
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
