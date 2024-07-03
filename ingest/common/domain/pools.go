package commondomain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolExtracter defines the interface for extracting pools.
type PoolExtracter interface {
	// ExtractAll extracts all the pools available within the height associated
	// with the context.
	ExtractAll(ctx sdk.Context) (BlockPools, error)
	// ExtractChanged extracts the pools that were changed in the block height associated
	// with the context.
	ExtractChanged(ctx sdk.Context) (BlockPools, error)
}
