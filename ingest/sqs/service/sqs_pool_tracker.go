package service

import (
	"github.com/osmosis-labs/osmosis/v24/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v24/x/poolmanager/types"
)

// poolBlockUpdateTracker is a struct that tracks the pools that were updated in a block.
type poolBlockUpdateTracker struct {
	concentratedPools            []poolmanagertypes.PoolI
	concentratedPoolIDTickChange map[uint64]struct{}
	cfmmPools                    []poolmanagertypes.PoolI
	cosmwasmPools                []poolmanagertypes.PoolI
}

// NewPoolTracker creates a new poolBlockUpdateTracker.
func NewPoolTracker() domain.BlockPoolUpdateTracker {
	return &poolBlockUpdateTracker{
		concentratedPools:            []poolmanagertypes.PoolI{},
		concentratedPoolIDTickChange: map[uint64]struct{}{},
		cfmmPools:                    []poolmanagertypes.PoolI{},
		cosmwasmPools:                []poolmanagertypes.PoolI{},
	}
}

// TrackConcentrated implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackConcentrated(pool poolmanagertypes.PoolI) {
	pt.concentratedPools = append(pt.concentratedPools, pool)
}

// TrackCFMM implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCFMM(pool poolmanagertypes.PoolI) {
	pt.cfmmPools = append(pt.cfmmPools, pool)
}

// TrackCosmWasm implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackCosmWasm(pool poolmanagertypes.PoolI) {
	pt.cosmwasmPools = append(pt.cosmwasmPools, pool)
}

// TrackConcentratedPoolIDTickChange implements PoolTracker.
func (pt *poolBlockUpdateTracker) TrackConcentratedPoolIDTickChange(poolID uint64) {
	pt.concentratedPoolIDTickChange[poolID] = struct{}{}
}

// GetConcentratedPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetConcentratedPools() []poolmanagertypes.PoolI {
	return pt.concentratedPools
}

// GetConcentratedPoolIDTickChange implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetConcentratedPoolIDTickChange() map[uint64]struct{} {
	return pt.concentratedPoolIDTickChange
}

// GetCFMMPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCFMMPools() []poolmanagertypes.PoolI {
	return pt.cfmmPools
}

// GetCosmWasmPools implements PoolTracker.
func (pt *poolBlockUpdateTracker) GetCosmWasmPools() []poolmanagertypes.PoolI {
	return pt.cosmwasmPools
}

// Reset implements PoolTracker.
func (pt *poolBlockUpdateTracker) Reset() {
	pt.concentratedPools = []poolmanagertypes.PoolI{}
	pt.cfmmPools = []poolmanagertypes.PoolI{}
	pt.cosmwasmPools = []poolmanagertypes.PoolI{}
	pt.concentratedPoolIDTickChange = map[uint64]struct{}{}
}
