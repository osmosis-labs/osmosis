package pooltracker

import (
	"sync"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v28/x/poolmanager/types"
)

// poolBlockUpdateTracker is a struct that tracks the pools that were updated in a block.
type poolBlockUpdateTracker struct {
	concentratedPools             sync.Map
	concentratedPoolIDTickChange  sync.Map
	cfmmPools                     sync.Map
	cosmwasmPools                 sync.Map
	cosmwasmPoolsAddressToPoolMap sync.Map
	// Tracks the pool IDs that were created in the block.
	createdPoolIDs sync.Map
}

// NewMemory creates a new memory pool tracker.
func NewMemory() domain.BlockPoolUpdateTracker {
	return &poolBlockUpdateTracker{
		concentratedPools:             sync.Map{},
		concentratedPoolIDTickChange:  sync.Map{},
		cfmmPools:                     sync.Map{},
		cosmwasmPools:                 sync.Map{},
		cosmwasmPoolsAddressToPoolMap: sync.Map{},
		createdPoolIDs:                sync.Map{},
	}
}

// TrackConcentrated implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackConcentrated(pool poolmanagertypes.PoolI) {
	pt.concentratedPools.Store(pool.GetId(), pool)
}

// TrackCFMM implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCFMM(pool poolmanagertypes.PoolI) {
	pt.cfmmPools.Store(pool.GetId(), pool)
}

// TrackCosmWasm implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCosmWasm(pool poolmanagertypes.PoolI) {
	pt.cosmwasmPools.Store(pool.GetId(), pool)
}

// TrackCosmWasmPoolsAddressToPoolMap implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCosmWasmPoolsAddressToPoolMap(pool poolmanagertypes.PoolI) {
	pt.cosmwasmPoolsAddressToPoolMap.Store(pool.GetAddress().String(), pool)
}

// TrackConcentratedPoolIDTickChange implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackConcentratedPoolIDTickChange(poolID uint64) {
	pt.concentratedPoolIDTickChange.Store(poolID, struct{}{})
}

// TrackCreatedPoolID implements domain.BlockPoolUpdateTracker.
func (pt *poolBlockUpdateTracker) TrackCreatedPoolID(poolCreation commondomain.PoolCreation) {
	pt.createdPoolIDs.Store(poolCreation.PoolId, poolCreation)
}

// GetConcentratedPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetConcentratedPools() []poolmanagertypes.PoolI {
	return poolMapToSlice(&pt.concentratedPools)
}

// GetConcentratedPoolIDTickChange implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetConcentratedPoolIDTickChange() map[uint64]struct{} {
	concentratedPoolIDTickChange := make(map[uint64]struct{})
	pt.concentratedPoolIDTickChange.Range(func(key, value interface{}) bool {
		k, ok := key.(uint64)
		if !ok {
			return true
		}

		v, ok := value.(struct{})
		if !ok {
			return true
		}

		concentratedPoolIDTickChange[k] = v

		return true
	})
	return concentratedPoolIDTickChange
}

// GetCFMMPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCFMMPools() []poolmanagertypes.PoolI {
	return poolMapToSlice(&pt.cfmmPools)
}

// GetCosmWasmPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCosmWasmPools() []poolmanagertypes.PoolI {
	return poolMapToSlice(&pt.cosmwasmPools)
}

// GetCosmWasmPoolsAddressToIDMap implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCosmWasmPoolsAddressToIDMap() map[string]poolmanagertypes.PoolI {
	cosmwasmPoolsAddressToPoolMap := make(map[string]poolmanagertypes.PoolI)
	pt.cosmwasmPoolsAddressToPoolMap.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if !ok {
			return true
		}

		v, ok := value.(poolmanagertypes.PoolI)
		if !ok {
			return true
		}

		cosmwasmPoolsAddressToPoolMap[k] = v

		return true
	})
	return cosmwasmPoolsAddressToPoolMap
}

// GetCreatedPoolIDs implements domain.BlockPoolUpdateTracker.
func (pt *poolBlockUpdateTracker) GetCreatedPoolIDs() map[uint64]commondomain.PoolCreation {
	createdPoolIDs := make(map[uint64]commondomain.PoolCreation)
	pt.createdPoolIDs.Range(func(key, value interface{}) bool {
		k, ok := key.(uint64)
		if !ok {
			return true
		}

		v, ok := value.(commondomain.PoolCreation)
		if !ok {
			return true
		}

		createdPoolIDs[k] = v

		return true
	})

	return createdPoolIDs
}

// Reset implements PoolTracker.
func (pt *poolBlockUpdateTracker) Reset() {
	pt.concentratedPools = sync.Map{}
	pt.cfmmPools = sync.Map{}
	pt.cosmwasmPools = sync.Map{}
	pt.concentratedPoolIDTickChange = sync.Map{}
	pt.createdPoolIDs = sync.Map{}
}

// poolMapToSlice converts a map of pools to a slice of pools.
func poolMapToSlice(m *sync.Map) []poolmanagertypes.PoolI {
	var result []poolmanagertypes.PoolI
	m.Range(func(_, value interface{}) bool {
		v, ok := value.(poolmanagertypes.PoolI)
		if !ok {
			return true
		}
		result = append(result, v)

		return true
	})

	return result
}
