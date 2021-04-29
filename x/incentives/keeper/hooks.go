package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
)

var _ lockuptypes.LockupHooks = Hooks{}

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) OnTokenLocked(ctx sdk.Context, _ sdk.AccAddress, _ uint64, amount sdk.Coins, _ time.Duration, _ time.Time) {
	h.k.IncreaseTotalLocked(ctx, amount)
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, _ sdk.AccAddress, _ uint64, amount sdk.Coins, _ time.Duration, _ time.Time) {
	h.k.DecreaseTotalLocked(ctx, amount)
}
