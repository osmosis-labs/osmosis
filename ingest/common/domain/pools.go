package commondomain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolExtractor defines the interface for extracting pools.
type PoolExtractor interface {
	// ExtractAll extracts all the pools available within the height associated
	// with the context.
	ExtractAll(ctx sdk.Context) (BlockPools, map[uint64]PoolCreation, error)
	// ExtractChanged extracts the pools that were changed in the block height associated
	// with the context.
	ExtractChanged(ctx sdk.Context) (BlockPools, error)
	// ExtractrCreated extracts the pools that were created in the block height associated
	// with the context.
	ExtractCreated(ctx sdk.Context) (BlockPools, map[uint64]PoolCreation, error)
}
