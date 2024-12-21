package pooltracker

import (
	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"
)

// poolBlockUpdateTracker is a struct that tracks the pools that were updated in a block.
type poolBlockUpdateTracker struct {
	concentratedPools             map[uint64]poolmanagertypes.PoolI
	concentratedPoolIDTickChange  map[uint64]struct{}
	cfmmPools                     map[uint64]poolmanagertypes.PoolI
	cosmwasmPools                 map[uint64]poolmanagertypes.PoolI
	cosmwasmPoolsAddressToPoolMap map[string]poolmanagertypes.PoolI
	// Tracks the pool IDs that were created in the block.
	createdPoolIDs map[uint64]commondomain.PoolCreation
}

// NewMemory creates a new memory pool tracker.
func NewMemory() domain.BlockPoolUpdateTracker {
	return &poolBlockUpdateTracker{
		concentratedPools:             map[uint64]poolmanagertypes.PoolI{},
		concentratedPoolIDTickChange:  map[uint64]struct{}{},
		cfmmPools:                     map[uint64]poolmanagertypes.PoolI{},
		cosmwasmPools:                 map[uint64]poolmanagertypes.PoolI{},
		cosmwasmPoolsAddressToPoolMap: map[string]poolmanagertypes.PoolI{},
		createdPoolIDs:                map[uint64]commondomain.PoolCreation{},
	}
}

// TrackConcentrated implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackConcentrated(pool poolmanagertypes.PoolI) {
	pt.concentratedPools[pool.GetId()] = pool
}

// TrackCFMM implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCFMM(pool poolmanagertypes.PoolI) {
	pt.cfmmPools[pool.GetId()] = pool
}

// TrackCosmWasm implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCosmWasm(pool poolmanagertypes.PoolI) {
	pt.cosmwasmPools[pool.GetId()] = pool
}

// TrackCosmWasmPoolsAddressToPoolMap implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCosmWasmPoolsAddressToPoolMap(pool poolmanagertypes.PoolI) {
	pt.cosmwasmPoolsAddressToPoolMap[pool.GetAddress().String()] = pool
}

// TrackConcentratedPoolIDTickChange implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackConcentratedPoolIDTickChange(poolID uint64) {
	pt.concentratedPoolIDTickChange[poolID] = struct{}{}
}

// TrackCreatedPoolID implements domain.BlockPoolUpdateTracker.
func (pt *poolBlockUpdateTracker) TrackCreatedPoolID(poolCreation commondomain.PoolCreation) {
	pt.createdPoolIDs[poolCreation.PoolId] = poolCreation
}

// GetConcentratedPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetConcentratedPools() []poolmanagertypes.PoolI {
	return poolMapToSlice(pt.concentratedPools)
}

// GetConcentratedPoolIDTickChange implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetConcentratedPoolIDTickChange() map[uint64]struct{} {
	return pt.concentratedPoolIDTickChange
}

// GetCFMMPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCFMMPools() []poolmanagertypes.PoolI {
	return poolMapToSlice(pt.cfmmPools)
}

// GetCosmWasmPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCosmWasmPools() []poolmanagertypes.PoolI {
	return poolMapToSlice(pt.cosmwasmPools)
}

// GetCosmWasmPoolsAddressToIDMap implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCosmWasmPoolsAddressToIDMap() map[string]poolmanagertypes.PoolI {
	return pt.cosmwasmPoolsAddressToPoolMap
}

// GetCreatedPoolIDs implements domain.BlockPoolUpdateTracker.
func (pt *poolBlockUpdateTracker) GetCreatedPoolIDs() map[uint64]commondomain.PoolCreation {
	return pt.createdPoolIDs
}

// Reset implements PoolTracker.
func (pt *poolBlockUpdateTracker) Reset() {
	pt.concentratedPools = map[uint64]poolmanagertypes.PoolI{}
	pt.cfmmPools = map[uint64]poolmanagertypes.PoolI{}
	pt.cosmwasmPools = map[uint64]poolmanagertypes.PoolI{}
	pt.concentratedPoolIDTickChange = map[uint64]struct{}{}
	pt.createdPoolIDs = map[uint64]commondomain.PoolCreation{}
}

// poolMapToSlice converts a map of pools to a slice of pools.
func poolMapToSlice(m map[uint64]poolmanagertypes.PoolI) []poolmanagertypes.PoolI {
	result := make([]poolmanagertypes.PoolI, 0, len(m))
	for _, pool := range m {
		result = append(result, pool)
	}
	return result
}
