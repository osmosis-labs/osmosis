package service

import (
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// poolBlockUpdateTracker is a struct that tracks the pools that were updated in a block.
type poolBlockUpdateTracker struct {
	concentratedPools             map[uint64]poolmanagertypes.PoolI
	concentratedPoolIDTickChange  map[uint64]struct{}
	cfmmPools                     map[uint64]poolmanagertypes.PoolI
	cosmwasmPools                 map[uint64]poolmanagertypes.PoolI
	cosmwasmPoolsAddressToPoolMap map[string]poolmanagertypes.PoolI
}

// NewPoolTracker creates a new poolBlockUpdateTracker.
func NewPoolTracker() domain.BlockPoolUpdateTracker {
	return &poolBlockUpdateTracker{
		concentratedPools:             map[uint64]poolmanagertypes.PoolI{},
		concentratedPoolIDTickChange:  map[uint64]struct{}{},
		cfmmPools:                     map[uint64]poolmanagertypes.PoolI{},
		cosmwasmPools:                 map[uint64]poolmanagertypes.PoolI{},
		cosmwasmPoolsAddressToPoolMap: map[string]poolmanagertypes.PoolI{},
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

// Reset implements PoolTracker.
func (pt *poolBlockUpdateTracker) Reset() {
	pt.concentratedPools = map[uint64]poolmanagertypes.PoolI{}
	pt.cfmmPools = map[uint64]poolmanagertypes.PoolI{}
	pt.cosmwasmPools = map[uint64]poolmanagertypes.PoolI{}
	pt.concentratedPoolIDTickChange = map[uint64]struct{}{}
}

// poolMapToSlice converts a map of pools to a slice of pools.
func poolMapToSlice(m map[uint64]poolmanagertypes.PoolI) []poolmanagertypes.PoolI {
	result := make([]poolmanagertypes.PoolI, 0, len(m))
	for _, pool := range m {
		result = append(result, pool)
	}
	return result
}
