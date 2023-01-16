package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type PoolCreationListener interface {
	// AfterPoolCreated is called after CreatePool
	AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
}

type PoolCreationListeners []PoolCreationListener

func (l PoolCreationListeners) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range l {
		l[i].AfterPoolCreated(ctx, sender, poolId)
	}
}

// Creates hooks for the Gamm Module.
func NewPoolCreationListeners(listeners ...PoolCreationListener) PoolCreationListeners {
	return listeners
}
