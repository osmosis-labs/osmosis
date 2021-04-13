package keeper

import (
	"time"

	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	k.SetUserActionHistory(ctx, sender, 1)
}

func (k Keeper) OnTokenLocked(ctx sdk.Context, sender sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	k.SetUserActionHistory(ctx, sender, 2)
}

func (k Keeper) OnTokenUnlocked(ctx sdk.Context, sender sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	k.SetUserActionHistory(ctx, sender, 3)
}

//_________________________________________________________________________________________

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

var _ gammtypes.GammHooks = Hooks{}
var _ lockuptypes.LockupHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// Implements hooks
func (h Hooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	h.k.AfterPoolCreated(ctx, sender, poolId)
}

// Implements hooks
func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	h.k.OnTokenLocked(ctx, address, lockID, amount, lockDuration, unlockTime)
}

// Implements hooks
func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	h.k.OnTokenUnlocked(ctx, address, lockID, amount, lockDuration, unlockTime)
}
