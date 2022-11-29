package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GammKeeper defines the expected interface needed for swaprouter module
type GammKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (PoolI, error)

	GetNextPoolId(ctx sdk.Context) uint64
}
