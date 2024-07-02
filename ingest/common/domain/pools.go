package commondomain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PoolExtracter interface {
	ExtractAll(ctx sdk.Context) (BlockPools, error)
	ExtractChanged(ctx sdk.Context) (BlockPools, error)
}
