package poolextractor_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/common/poolextractor"
	"github.com/osmosis-labs/osmosis/v28/ingest/common/pooltracker"
)

type PoolExtractorTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestPoolExtractorTestSuite(t *testing.T) {
	suite.Run(t, new(PoolExtractorTestSuite))
}

// TestExtractor tests that the appropriate pools are extracted
// when calling ExtractAll and ExtractChanged methods of the extractor.
func (s *PoolExtractorTestSuite) TestExtractor() {

	s.Setup()

	// Initialized chain pools
	chainPools := s.PrepareAllSupportedPools()

	// Get all chain pools from state for asserting later
	allChainPools, err := s.App.PoolManagerKeeper.AllPools(s.Ctx)
	s.Require().NoError(err)

	// Initialize a position on the concentrated pool
	concentratedPoolWithPosition := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(apptesting.ETH, apptesting.USDC)

	keepers := commondomain.PoolExtractorKeepers{
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

	// Inject a new pool creation and track it
	poolTracker.TrackCreatedPoolID(commondomain.PoolCreation{
		PoolId:      concentratedPool.GetId(),
		BlockHeight: 1000,
		BlockTime:   s.Ctx.BlockTime(),
		TxnHash:     "txnhash",
	})

	// Initialize the extractor
	extractor := poolextractor.New(keepers, poolTracker)

	// System under test #1
	blockPools, createdPoolIDs, err := extractor.ExtractAll(s.Ctx)
	s.Require().NoError(err)

	// Validate all pools are extracted
	allPools := blockPools.GetAll()
	// + 1 for an extra concentrated pool.
	s.Require().Equal(len(allChainPools)+1, len(allPools))

	// System under test #2
	// Extract the pools again but now only changed
	blockPools, err = extractor.ExtractChanged(s.Ctx)
	s.Require().NoError(err)

	// Validate only the changed pools are extracted
	changedPools := blockPools.GetAll()
	s.Require().Equal(2, len(changedPools))

	// Validate that the newly created pool is extracted
	// Since only one newly created pool is injected during the test in the above code earlier,
	// the length of the createdPoolIDs should be 1.
	// the length of the pools.GetAll() should be equal to the length of the createdPoolIDs.
	pools, createdPoolIDs, err := extractor.ExtractCreated(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(createdPoolIDs), len(pools.GetAll()))
	s.Require().Equal(1, len(createdPoolIDs))

	// Validate that the tick change is detected
	s.Require().Len(blockPools.ConcentratedPoolIDTickChange, 2)
	_, ok := blockPools.ConcentratedPoolIDTickChange[concentratedPoolWithPosition.GetId()]
	s.Require().True(ok)
}
