package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"
)

// BlockPoolUpdateTracker is an interface for tracking
// the pools that were updated in a block.
// It persists the pools using "Track" methods in its internal state.
// It tracks the latest pool update, discarding the previous updates.
// Only on Reset, the internal state is cleared.
type BlockPoolUpdateTracker interface {
	// TrackConcentrated tracks the concentrated pool.
	TrackConcentrated(pool poolmanagertypes.PoolI)

	// TrackConcentratedPoolIDTickChange tracks the concentrated pool ID tick change.
	// Due to internal implementation, it is non-trivial to apply tick changes.
	// As a result, we track the pool ID tick change and read the pool with all of its ticks
	// if at least one tick change was applied within the block.
	TrackConcentratedPoolIDTickChange(poolID uint64)

	// TrackCFMM tracks the CFMM pool.
	TrackCFMM(pool poolmanagertypes.PoolI)

	// TrackCosmWasm tracks the CosmWasm pool.
	TrackCosmWasm(pool poolmanagertypes.PoolI)

	// TrackCosmWasmPoolsAddressToPoolMap tracks the CosmWasm pools address to the pool object map.
	TrackCosmWasmPoolsAddressToPoolMap(pool poolmanagertypes.PoolI)

	// TrackCreatedPoolID tracks whenever a new pool is created.
	// CONTRACT: the caller calls this method only once per pool creation as observed
	// by poolmanagertypes.TypeEvtPoolCreated
	TrackCreatedPoolID(commondomain.PoolCreation)

	// GetConcentratedPools returns the tracked concentrated pools.
	GetConcentratedPools() []poolmanagertypes.PoolI

	// GetConcentratedPoolIDTickChange returns the tracked concentrated pool ID tick change.
	GetConcentratedPoolIDTickChange() map[uint64]struct{}

	// GetCFMMPools returns the tracked CFMM pools.
	GetCFMMPools() []poolmanagertypes.PoolI

	// GetCosmWasmPools returns the tracked CosmWasm pools.
	GetCosmWasmPools() []poolmanagertypes.PoolI

	// GetCosmWasmPoolsAddressToIDMap returns the tracked CosmWasm pools address to pool object map.
	GetCosmWasmPoolsAddressToIDMap() map[string]poolmanagertypes.PoolI

	// GetCreatedPoolIDs returns the tracked pool IDs that were created in the block.
	GetCreatedPoolIDs() map[uint64]commondomain.PoolCreation

	// Reset clears the internal state.
	Reset()
}

// NodeStatusChecker is an interface for checking the node status.
type NodeStatusChecker interface {
	// IsNodeSyncing checks if the node is syncing.
	// Returns true if the node is syncing, false otherwise.
	// Returns error if the node syncing status cannot be determined.
	IsNodeSyncing(ctx sdk.Context) (bool, error)
}
