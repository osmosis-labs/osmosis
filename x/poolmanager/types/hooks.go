package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type PoolManagerHooks interface {
	// AfterPoolCreated is called after CreatePool
	AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
}

var _ PoolManagerHooks = MultiPoolManagerhooks{}

// combine multiple PoolManager hooks, all hook functions are run in array sequence.
type MultiPoolManagerhooks []PoolManagerHooks

// Creates hooks for the PoolManager Module.
func NewMultiPoolManagerhooks(hooks ...PoolManagerHooks) MultiPoolManagerhooks {
	return hooks
}

func (h MultiPoolManagerhooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range h {
		h[i].AfterPoolCreated(ctx, sender, poolId)
	}
}
