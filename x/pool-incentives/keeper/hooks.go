package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
)

type Hooks struct {
	k Keeper
}

var _ gammtypes.GammHooks = Hooks{}

// Create new pool incentives hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// creates a pot for each pool’s lockable duration
func (h Hooks) AfterPoolCreated(ctx sdk.Context, poolId uint64) {
	err := h.k.CreatePoolPots(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

// execute distribution after adding collected fees to fee pool
func (h Hooks) AfterAddCollectedFees(ctx sdk.Context, fees sdk.Coins) {
	// @Sunny, @Tony, @Dev, what comments should we keep after modifying own BeginBlocker to hooks?

	// WARNING: The order of how modules interact with the default distribution module matters if the distribution module is used in a similar way to:
	// 1. mint module or 2. custom mint modules that uses the auth module’s feeCollector to mint new tokens that uses the default distribution module.
	// Currently, the mint module mints inflation amount to the feeCollector module account,
	// and on the next BeginBlock the distribution module uses all the available balance of the feeCollector to process the distribution.
	// Therefore, for the pool-incentives module to only take the AllocationRatio from the inflated amount, it should be run before the distribution module’s BeginBlock.
	// Also, the pool-incentives module first takes the AllocationRatio from the total inflated amount, and the remainder is used by the distribution.
	// So the amount is relative to each other. For example, if the AllocationRatio is 0.2(20%),
	// the distribution uses the remaining 80% to calculate–which means if the community pool is set to receive 10% of newly minted OSMO, community pool is 8% of the total inflation.

	// Calculate the AllocatableAsset using the AllocationRatio and the MintedDenom,
	// then allocate the tokens to the registered pools’ pots.
	// If there is no record, inflation is not drained and the all amounts are used by the distribution module’s next BeginBlock.
	asset := h.k.GetAllocatableAsset(ctx)
	if asset.IsValid() && asset.IsPositive() {
		err := h.k.AllocateAsset(ctx, asset)
		if err != nil {
			panic(err)
		}
	}
}
