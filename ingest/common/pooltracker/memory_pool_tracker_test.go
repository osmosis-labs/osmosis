package pooltracker_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v29/app/apptesting"

	"github.com/osmosis-labs/osmosis/v29/ingest/common/pooltracker"
)

type PoolTrackerTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestPoolTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(PoolTrackerTestSuite))
}

// This is a sanity-check test ensuring that all pools are tracked and returned correctly.
func (s *PoolTrackerTestSuite) TestPoolTracker_Track() {
	s.Setup()

	allPools := s.PrepareAllSupportedPools()

	poolTracker := pooltracker.NewMemory()

	////////// Concetrated

	// Get the concentrated pool
	concentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, allPools.ConcentratedPoolID)
	s.Require().NoError(err)

	// Track the concentrated pool
	poolTracker.TrackConcentrated(concentratedPool)

	// Track the same pool to ensure no duplicates
	poolTracker.TrackConcentrated(concentratedPool)

	// Get the concentrated pool
	concentratedPools := poolTracker.GetConcentratedPools()
	s.Require().Len(concentratedPools, 1)

	////////// CFMM

	// Get balancer pools
	balancerPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPools.BalancerPoolID)
	s.Require().NoError(err)

	// Track the balancer pool
	poolTracker.TrackCFMM(balancerPool)

	// Track the same pool to ensure no duplicates
	poolTracker.TrackCFMM(balancerPool)

	// Get stableswap pool
	stableswapPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPools.StableSwapPoolID)
	s.Require().NoError(err)

	// Track the stableswap pool
	poolTracker.TrackCFMM(stableswapPool)

	// Track the same pool to ensure no duplicates
	poolTracker.TrackCFMM(stableswapPool)

	// Get the CFMM pools
	cfmmPools := poolTracker.GetCFMMPools()
	s.Require().Len(cfmmPools, 2)

	////////// CosmWasm

	// Get the CosmWasm pool
	cosmWasmPool, err := s.App.PoolManagerKeeper.GetPool(s.Ctx, allPools.CosmWasmPoolID)
	s.Require().NoError(err)

	// Track the CosmWasm pool
	poolTracker.TrackCosmWasm(cosmWasmPool)

	// Track the same pool to ensure no duplicates
	poolTracker.TrackCosmWasm(cosmWasmPool)

	// Get the CosmWasm pools
	cosmWasmPools := poolTracker.GetCosmWasmPools()
	s.Require().Len(cosmWasmPools, 1)

	// Track concentrated tick change
	poolTracker.TrackConcentratedPoolIDTickChange(allPools.ConcentratedPoolID)

	// Track the same tick change to ensure no duplicates
	poolTracker.TrackConcentratedPoolIDTickChange(allPools.ConcentratedPoolID)

	// Get concentrated tick change
	concentratedPoolIDTickChange := poolTracker.GetConcentratedPoolIDTickChange()
	s.Require().Len(concentratedPoolIDTickChange, 1)

	// Reset the pool tracker
	poolTracker.Reset()

	// All fields should be empty
	concentratedPools = poolTracker.GetConcentratedPools()
	s.Require().Len(concentratedPools, 0)

	cfmmPools = poolTracker.GetCFMMPools()
	s.Require().Len(cfmmPools, 0)

	cosmWasmPools = poolTracker.GetCosmWasmPools()
	s.Require().Len(cosmWasmPools, 0)

	concentratedPoolIDTickChange = poolTracker.GetConcentratedPoolIDTickChange()
	s.Require().Len(concentratedPoolIDTickChange, 0)
}
