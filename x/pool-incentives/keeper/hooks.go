package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	minttypes "github.com/osmosis-labs/osmosis/v27/x/mint/types"
)

type Hooks struct {
	k Keeper
}

var (
	_ gammtypes.GammHooks = Hooks{}
	_ minttypes.MintHooks = Hooks{}
)

// Create new pool incentives hooks.
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// AfterCFMMPoolCreated creates a gauge for each pool’s lockable duration.
func (h Hooks) AfterCFMMPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := h.k.CreateLockablePoolGauges(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

// AfterJoinPool hook is a noop.
func (h Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount osmomath.Int) {
}

// AfterExitPool hook is a noop.
func (h Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount osmomath.Int, exitCoins sdk.Coins) {
}

// AfterCFMMSwap hook is a noop.
func (h Hooks) AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
}

// Distribute coins after minter module allocate assets to pool-incentives module.
func (h Hooks) AfterDistributeMintedCoin(ctx sdk.Context) {
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
	// then allocate the tokens to the registered pools’ gauges.
	// If there is no record, inflation is not drained and the all amounts are used by the distribution module’s next BeginBlock.
	err := h.k.AllocateAsset(ctx)
	if err != nil {
		panic(err)
	}
}

// AfterConcentratedPoolCreated creates a single gauge for the concentrated liquidity pool.
func (h Hooks) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := h.k.CreateConcentratedLiquidityPoolGauge(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

// AfterInitialPoolPositionCreated is a noop.
func (h Hooks) AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
}

// AfterLastPoolPositionRemoved is a noop.
func (h Hooks) AfterLastPoolPositionRemoved(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
}

// AfterConcentratedPoolSwap is a noop.
func (h Hooks) AfterConcentratedPoolSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
}
