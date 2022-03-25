package keeper

import (
	// "fmt"

	// "github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

// at the end of each epoch, swap all non-OSMO fees into OSMO and transfer to fee module account
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {

	// get module address for FooCollectorName
	addrFoo := k.accountKeeper.GetModuleAddress(txfeestypes.FooCollectorName)

	// get balances for all denoms in the module account
	fooAccountBalances := k.bankKeeper.GetAllBalances(ctx, addrFoo)

	// pulls base denom from TxFeesKeeper (should be uOSMO)
	baseDenom, _ := k.GetBaseDenom(ctx)

	// iterate through the resulting array and swap each denom into OSMO using the GAMM module
	for _, coin := range fooAccountBalances {

		if coin.Denom == baseDenom {
			continue
		} else {
		
			// TO DO: figure out how to cast this or get the pool ID for the main OSMO paired pool
			feetoken, _ := k.GetFeeToken(ctx, coin.Denom)

			// swap into OSMO
			// question: is spotPrice really the minimum out for the swap?
			k.gammKeeper.SwapExactAmountIn(ctx, addrFoo, feetoken.PoolID, coin, baseDenom, sdk.ZeroInt())
		}
	}

	// Potential error to test for: do swaps to OSMO consolidate into a single denom in fooAccountBalances? Should be yes but keep an eye out

	// send all OSMO to fee module account to be distributed in the next block
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
