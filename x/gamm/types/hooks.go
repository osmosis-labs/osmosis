package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type GammHooks interface {
	AfterPoolCreated(ctx sdk.Context, poolId uint64)
}

// combine multiple gamm hooks, all hook functions are run in array sequence
type MultiGammHooks []GammHooks

func NewMultiStakingHooks(hooks ...GammHooks) MultiGammHooks {
	return hooks
}

func (h MultiGammHooks) AfterPoolCreated(ctx sdk.Context, poolId uint64) {
	for i := range h {
		h[i].AfterPoolCreated(ctx, poolId)
	}
}
