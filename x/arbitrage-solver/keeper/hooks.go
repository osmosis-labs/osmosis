package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Hooks struct {
	k Keeper
}

// var _ gammtypes.GammHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterPoolCreated is called after CreatePool
func (h *Hooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {

}
func (h *Hooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {

}
func (h *Hooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {

}
func (h *Hooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	h.k.RecordSwap(ctx, sender, poolId, input, output)
}
