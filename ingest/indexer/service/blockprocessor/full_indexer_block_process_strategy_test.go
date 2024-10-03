package blockprocessor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v26/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v26/ingest/common/domain/mocks"
	indexerdomain "github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
	indexermocks "github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain/mocks"
	"github.com/osmosis-labs/osmosis/v26/ingest/indexer/service/blockprocessor"
)

var (
	DefaultConcentratedPoolId      = uint64(1)
	DefaultConcentratedPoolHeight  = int64(12345)
	DefaultConcentratedPoolTime    = time.Now()
	DefaultConcentratedPoolTxnHash = "txhash"
	DefaultCfmmPoolId              = uint64(2)
	DefaultCfmmPoolHeight          = int64(12346)
	DefaultCfmmPoolTime            = time.Now()
	DefaultCfmmPoolTxnHash         = "txhash2"
)

type FullIndexerBlockProcessStrategyTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestFullIndexerBlockProcessStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(FullIndexerBlockProcessStrategyTestSuite))
}

// The purpose of this test is to verify that the PublishAllSupplies method correctly publishes
// the token supplies and offsets based on the primed data from the state.

// Token supplies and offsets are primed in the test, thru PrepareAllSupportedPools():
// - axlusdc 100001000000000
// - bar 40005000000
// - baz 40005000000
// - factory/osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r/alloyed/allusdc 20000000000
// - foo 40005000000
// - gravusdc 100001000000000
// - stake 225000001000000, with offset -225000000000000
// - uosmo 51005000000
//
// Therefore, we expect the following:
// - 8 calls to PublishTokenSupply
// - 1 call to PublishTokenSupplyOffset
func (s *FullIndexerBlockProcessStrategyTestSuite) TestPublishAllSupplies() {
	tests := []struct {
		name                                     string
		expectedNumPublishTokenSupplyCalls       int
		expectedNumPublishTokenSupplyOffsetCalls int
	}{
		{
			name:                                     "happy path with the primed data from state",
			expectedNumPublishTokenSupplyCalls:       8,
			expectedNumPublishTokenSupplyOffsetCalls: 1,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			// pool id 1 created below
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			// pool id 2, 3 created below
			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			// pool id 4, 5 created below
			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)
			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			// Mock out pool extractor
			poolsExtracter := &commonmocks.PoolsExtractorMock{
				BlockPools: blockPools,
			}

			// Mock out publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Mock out pair publisher
			pairPublisherMock := &indexermocks.MockPairPublisher{}

			// Initialize keepers
			keepers := indexerdomain.Keepers{
				PoolManagerKeeper: s.App.PoolManagerKeeper,
				BankKeeper:        s.App.BankKeeper,
			}

			blockProcessor := blockprocessor.NewFullIndexerBlockProcessStrategy(publisherMock, keepers, poolsExtracter, pairPublisherMock)

			blockProcessor.PublishAllSupplies(s.Ctx)
			s.Require().Equal(test.expectedNumPublishTokenSupplyCalls, publisherMock.NumPublishTokenSupplyCalls)
			s.Require().Equal(test.expectedNumPublishTokenSupplyOffsetCalls, publisherMock.NumPublishTokenSupplyOffsetCalls)
		})
	}
}

// The purpose of this test is to verify that the ProcessPools method correctly publishes
// the full set of pool pairs, regardless of whether they have creation data or not.
// See also: block_updates_indexer_block_process_strategy_test::TestPublishCreatedPools,
// The difference is full_indexer_block_process_strategy_test always publishes all pool pairs,
// while block_updates_indexer_block_process_strategy_test only publishes when there is a creation data.
func (s *FullIndexerBlockProcessStrategyTestSuite) TestProcessPools() {
	tests := []struct {
		name                             string
		createdPoolIDs                   map[uint64]commondomain.PoolCreation
		expectedPublishPoolPairsCalled   bool
		expectedNumPoolsPublished        int
		expectedNumPoolsWithCreationData int
	}{
		{
			name: "happy path with one pool creation",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				DefaultConcentratedPoolId: {
					PoolId:      DefaultConcentratedPoolId,
					BlockHeight: DefaultConcentratedPoolHeight,
					BlockTime:   DefaultConcentratedPoolTime,
					TxnHash:     DefaultConcentratedPoolTxnHash,
				},
			},
			expectedPublishPoolPairsCalled:   true,
			expectedNumPoolsPublished:        5,
			expectedNumPoolsWithCreationData: 1,
		},
		{
			name: "happy path with multiple pool creation",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				DefaultConcentratedPoolId: {
					PoolId:      DefaultConcentratedPoolId,
					BlockHeight: DefaultConcentratedPoolHeight,
					BlockTime:   DefaultConcentratedPoolTime,
					TxnHash:     DefaultConcentratedPoolTxnHash,
				},
				DefaultCfmmPoolId: {
					PoolId:      DefaultCfmmPoolId,
					BlockHeight: DefaultCfmmPoolHeight,
					BlockTime:   DefaultCfmmPoolTime,
					TxnHash:     DefaultCfmmPoolTxnHash,
				},
			},
			expectedPublishPoolPairsCalled:   true,
			expectedNumPoolsPublished:        5,
			expectedNumPoolsWithCreationData: 2,
		},
		{
			name:                             "should publish even when there is no pool creation data",
			createdPoolIDs:                   map[uint64]commondomain.PoolCreation{},
			expectedPublishPoolPairsCalled:   true,
			expectedNumPoolsPublished:        5,
			expectedNumPoolsWithCreationData: 0,
		},
		{
			name: "should still publish but without creation data when pool creation data has no match in the pool list",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				999: {
					PoolId:      999,
					BlockHeight: 12345,
					BlockTime:   time.Now(),
					TxnHash:     "txhash",
				},
			},
			expectedPublishPoolPairsCalled:   true,
			expectedNumPoolsPublished:        5,
			expectedNumPoolsWithCreationData: 0,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			// pool id 1 created below
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			// pool id 2, 3 created below
			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			// pool id 4, 5 created below
			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)
			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			// Mock out pool extractor
			poolsExtracter := &commonmocks.PoolsExtractorMock{
				BlockPools:     blockPools,
				CreatedPoolIDs: test.createdPoolIDs,
			}

			// Mock out publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Mock out pair publisher
			pairPublisherMock := &indexermocks.MockPairPublisher{}

			// Initialize keepers
			keepers := indexerdomain.Keepers{
				PoolManagerKeeper: s.App.PoolManagerKeeper,
				BankKeeper:        s.App.BankKeeper,
			}

			blockProcessor := blockprocessor.NewFullIndexerBlockProcessStrategy(publisherMock, keepers, poolsExtracter, pairPublisherMock)

			err = blockProcessor.ProcessPools(s.Ctx)
			s.Require().NoError(err)

			// Check that the pair publisher is called correctly
			s.Require().Equal(test.expectedPublishPoolPairsCalled, pairPublisherMock.PublishPoolPairsCalled)
			if test.expectedPublishPoolPairsCalled {
				// Check that the number of pools published
				s.Require().Equal(test.expectedNumPoolsPublished, pairPublisherMock.NumPoolPairPublished)
				// Check that the pools and created pool IDs are set correctly
				s.Require().Equal(blockPools.GetAll(), pairPublisherMock.CalledWithPools)
				s.Require().Equal(test.createdPoolIDs, pairPublisherMock.CalledWithCreatedPoolIDs)
				// Check that the number of pools with creation data
				s.Require().Equal(test.expectedNumPoolsWithCreationData, pairPublisherMock.NumPoolPairWithCreationData)
			}

		})
	}
}
