package commondomain

import (
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// BlockPools contains the pools to be ingested in a block.
type BlockPools struct {
	// ConcentratedPools are the concentrated pools to be ingested.
	ConcentratedPools []poolmanagertypes.PoolI
	// ConcentratedPoolIDTickChange is the map of pool ID to tick change for concentrated pools.
	// We use these pool IDs to append concentrated pools with all ticks at the end of the block.
	ConcentratedPoolIDTickChange map[uint64]struct{}
	// CosmWasmPools are the CosmWasm pools to be ingested.
	CosmWasmPools []poolmanagertypes.PoolI
	// CFMMPools are the CFMM pools to be ingested.
	CFMMPools []poolmanagertypes.PoolI
}

func (bp BlockPools) GetAll() []poolmanagertypes.PoolI {
	allPools := make([]poolmanagertypes.PoolI, 0, len(bp.ConcentratedPools)+len(bp.CosmWasmPools)+len(bp.CFMMPools))
	allPools = append(allPools, bp.ConcentratedPools...)
	allPools = append(allPools, bp.CosmWasmPools...)
	allPools = append(allPools, bp.CFMMPools...)
	return allPools
}
