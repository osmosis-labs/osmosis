package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
)

type Hooks struct {
	k Keeper
}

var (
	_ minttypes.MintHooks = Hooks{}
)

// Hooks creates new pool incentives hooks.
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// AfterDistributeMintedCoin coins after minter module allocate assets to incentives module.
func (h Hooks) AfterDistributeMintedCoin(ctx sdk.Context) {

	// WARNING: The order of how modules interact with the default distribution module matters if the distribution module is used in a similar way to:
	// 1. mint module or 2. custom mint modules that uses the auth module’s feeCollector to mint new tokens that uses the default distribution module.
	// Currently, the mint module mints inflation amount to the feeCollector module account,
	// and on the next BeginBlock the distribution module uses all the available balance of the feeCollector to process the distribution.
	// Therefore, for the pool-incentives module to only take the AllocationRatio from the inflated amount, it should be run before the distribution module’s BeginBlock.
	// Also, the pool-incentives module first takes the AllocationRatio from the total inflated amount, and the remainder is used by the distribution.
	// So the amount is relative to each other. For example, if the AllocationRatio is 0.2(20%),
	// the distribution uses the remaining 80% to calculate–which means if the community pool is set to receive 10% of newly minted OSMO, community pool is 8% of the total inflation.
	err := h.k.AllocateAsset(ctx)
	if err != nil {
		panic(err)
	}
}
