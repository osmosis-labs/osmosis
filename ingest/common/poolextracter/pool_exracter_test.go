package poolextracter_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/common/poolextracter"
	"github.com/osmosis-labs/osmosis/v25/ingest/common/pooltracker"
)

type PoolExtracterTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestPoolExtracterTestSuite(t *testing.T) {
	suite.Run(t, new(PoolExtracterTestSuite))
}

// TestExtracter tests that the appropriate pools are extracted
// when calling ExtractAll and ExtractChanged methods of the extracter.
func (s *PoolExtracterTestSuite) TestExtracter() {

	s.Setup()

	// Initialized chain pools
	chainPools := s.PrepareAllSupportedPools()

	// Get all chain pools from state for asserting later
	allChainPools, err := s.App.PoolManagerKeeper.AllPools(s.Ctx)
	s.Require().NoError(err)

	// Initialize a position on the concentrated pool
	concentratedPoolWithPosition := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(apptesting.ETH, apptesting.USDC)

	keepers := commondomain.PoolExtracterKeepers{
		GammKeeper:         s.App.GAMMKeeper,
		CosmWasmPoolKeeper: s.App.CosmwasmPoolKeeper,
		WasmKeeper:         s.App.WasmKeeper,
		ConcentratedKeeper: s.App.ConcentratedLiquidityKeeper,
		PoolManagerKeeper:  s.App.PoolManagerKeeper,
		BankKeeper:         s.App.BankKeeper,
	}

	poolTracker := pooltracker.NewMemory()

	// Track only the concentrated pool
	concentratedPool, err := s.App.ConcentratedLiquidityKeeper.GetPool(s.Ctx, chainPools.ConcentratedPoolID)
	s.Require().NoError(err)
	poolTracker.TrackConcentrated(concentratedPool)

	// Track tick change for a concentraed pool.
	poolTracker.TrackConcentratedPoolIDTickChange(concentratedPoolWithPosition.GetId())

	// Initialize the extracter
	extracter := poolextracter.New(keepers, poolTracker)

	// System under test #1
	blockPools, err := extracter.ExtractAll(s.Ctx)
	s.Require().NoError(err)

	// Validate all pools are exracted
	allPools := blockPools.GetAll()
	// + 1 for an extra concentrated pool.
	s.Require().Equal(len(allChainPools)+1, len(allPools))

	// System under test #2
	// Extract the pools again but now only changed
	blockPools, err = extracter.ExtractChanged(s.Ctx)
	s.Require().NoError(err)

	// Validate only the changed pools are extracted
	changedPools := blockPools.GetAll()
	s.Require().Equal(2, len(changedPools))

	// Validate that the tick change is detected
	s.Require().Len(blockPools.ConcentratedPoolIDTickChange, 2)
	_, ok := blockPools.ConcentratedPoolIDTickChange[concentratedPoolWithPosition.GetId()]
	s.Require().True(ok)
}
