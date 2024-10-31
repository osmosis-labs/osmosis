package blockprocessor_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v27/ingest/common/domain/mocks"
	indexermocks "github.com/osmosis-labs/osmosis/v27/ingest/indexer/domain/mocks"
	"github.com/osmosis-labs/osmosis/v27/ingest/indexer/service/blockprocessor"
	sqsmocks "github.com/osmosis-labs/osmosis/v27/ingest/sqs/domain/mocks"
)

type BlockUpdateIndexerBlockProcessStrategyTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

// TestBlockUpdateIndexerBlockProcessStrategyTestSuite verifies the block update indexer strategy for processing created pools.
// The test suite initializes all supported pools (concentrated, cfmm, cosmwasm) via s.App.PrepareAllSupportedPools(), creating pool IDs 1-5.
// Test cases inject pool creation data and validate expected behavior of the pair publisher.
//
// Scenarios tested:
// - Happy path: single pool creation: should perform publishing
// - Happy path: multiple pool creation: should perform publishing
// - No pool creation: nothing is published
// - Pool creation data without a match: nothing is published
func TestBlockUpdateIndexerBlockProcessStrategyTestSuite(t *testing.T) {
	suite.Run(t, new(BlockUpdateIndexerBlockProcessStrategyTestSuite))
}

func (s *BlockUpdateIndexerBlockProcessStrategyTestSuite) TestPublishCreatedPools() {
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
				DefaultConcentratedPoolId: NewPoolCreation(DefaultConcentratedPoolId, DefaultConcentratedPoolHeight, DefaultConcentratedPoolTime, DefaultConcentratedPoolTxnHash),
			},
			expectedPublishPoolPairsCalled:   true,
			expectedNumPoolsPublished:        1,
			expectedNumPoolsWithCreationData: 1,
		},
		{
			name: "happy path with multiple pool creation",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				DefaultConcentratedPoolId: NewPoolCreation(DefaultConcentratedPoolId, DefaultConcentratedPoolHeight, DefaultConcentratedPoolTime, DefaultConcentratedPoolTxnHash),
				DefaultCfmmPoolId:         NewPoolCreation(DefaultCfmmPoolId, DefaultCfmmPoolHeight, DefaultCfmmPoolTime, DefaultCfmmPoolTxnHash),
			},
			expectedPublishPoolPairsCalled:   true,
			expectedNumPoolsPublished:        2,
			expectedNumPoolsWithCreationData: 2,
		},
		{
			name:                             "should not publish when there is no pool creation",
			createdPoolIDs:                   map[uint64]commondomain.PoolCreation{},
			expectedPublishPoolPairsCalled:   false,
			expectedNumPoolsPublished:        0,
			expectedNumPoolsWithCreationData: 0,
		},
		{
			name: "should not publish when pool creation data has no match in the pool list",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				NonExistentPoolId: NewPoolCreation(NonExistentPoolId, NonExistentPoolHeight, NonExistentPoolTime, NonExistentPoolTxnHash),
			},
			expectedPublishPoolPairsCalled:   false,
			expectedNumPoolsPublished:        0,
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

			// Mock out block updates process utils
			blockUpdatesProcessUtilsMock := &sqsmocks.BlockUpdateProcessUtilsMock{}

			// Mock out pool extractor
			poolsExtracter := &commonmocks.PoolsExtractorMock{
				BlockPools:     blockPools,
				CreatedPoolIDs: test.createdPoolIDs,
			}

			// Mock out publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Mock out pair publisher
			pairPublisherMock := &indexermocks.MockPairPublisher{}

			bprocess := blockprocessor.NewBlockUpdatesIndexerBlockProcessStrategy(blockUpdatesProcessUtilsMock, publisherMock, poolsExtracter, pairPublisherMock)

			err = bprocess.PublishCreatedPools(s.Ctx)
			s.Require().NoError(err)

			// Check that the pair publisher is called correctly
			s.Require().Equal(test.expectedPublishPoolPairsCalled, pairPublisherMock.PublishPoolPairsCalled)
			if test.expectedPublishPoolPairsCalled {
				// Check that the number of pools published
				s.Require().Equal(test.expectedNumPoolsPublished, pairPublisherMock.NumPoolPairPublished)
				// Check that the pools and created pool IDs are set correctly
				s.Require().Equal(test.createdPoolIDs, pairPublisherMock.CalledWithCreatedPoolIDs)
				// Check that the number of pools with creation data
				s.Require().Equal(test.expectedNumPoolsWithCreationData, pairPublisherMock.NumPoolPairWithCreationData)
			}
		})
	}
}
