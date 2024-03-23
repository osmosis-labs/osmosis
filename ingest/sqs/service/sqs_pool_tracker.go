package service

import "github.com/osmosis-labs/osmosis/v23/x/concentrated-liquidity/model"

type PoolTracker interface {
	TrackConcentrated(pool model.Pool) error

	GetConcentratedPools() []model.Pool

	Reset()
}

type poolTracker struct {
	concentratedPools []model.Pool
}

func NewPoolTracker() PoolTracker {
	return &poolTracker{
		concentratedPools: []model.Pool{},
	}
}

func (pt *poolTracker) TrackConcentrated(pool model.Pool) error {
	pt.concentratedPools = append(pt.concentratedPools, pool)
	return nil
}

// GetConcentratedPools implements PoolTracker.
func (pt *poolTracker) GetConcentratedPools() []model.Pool {
	return pt.concentratedPools
}

func (pt *poolTracker) Reset() {
	pt.concentratedPools = []model.Pool{}
}
