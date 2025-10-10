package blockprocessor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	indexermocks "github.com/osmosis-labs/osmosis/v31/ingest/indexer/domain/mocks"
	"github.com/osmosis-labs/osmosis/v31/ingest/indexer/service/blockprocessor"
)

type PairPublisherTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

// PairPublisherTestSuite tests the pair publisher.
// The pair publisher is responsible for iterating over all pools and their denoms combinations,
// with optional creation details provided in the parameter, it also append creation details to the pair struct.
// Since the test suite initializes all supported pools (concentrated, cfmm, cosmwasm) via s.App.PrepareAllSupportedPools(),
// the created pool IDs are 1-5, and the number of expected calls to the publisher is 12 coz:
// pool id 1 has 2 denoms, pool id 2 has 4 denoms, pool id 3 has 3 denoms, pool id 4 has 2 denoms, pool id 5 has 2 denoms.
// Scenarios tested:
// - Happy path without pool creation
// - publish pool pairs with two-denom pool creations
// - publish pool pairs with four-denom pool creations
func TestPairPublisherTestSuite(t *testing.T) {
	suite.Run(t, new(PairPublisherTestSuite))
}

func (s *PairPublisherTestSuite) TestPublishPoolPairs() {
	tests := []struct {
		name string
		// map of pool id to pool creation
		createdPoolIDs map[uint64]commondomain.PoolCreation
		// expected number of calls to the PublishPoolPair method
		expectedNumPublishPairCalls int
		// expected number of calls to the PublishPoolPair method with pool creation details
		expectedNumPublishPairCallsWithCreation int
	}{
		{
			name:                                    "happy path without pool creation",
			createdPoolIDs:                          map[uint64]commondomain.PoolCreation{},
			expectedNumPublishPairCalls:             12,
			expectedNumPublishPairCallsWithCreation: 0,
		},
		{
			name: "publish pool pairs with multiple two-denom pool creations",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				1: {
					PoolId:      1,
					BlockHeight: 12345,
					BlockTime:   time.Now(),
					TxnHash:     "txhash1",
				},
				4: {
					PoolId:      4,
					BlockHeight: 12346,
					BlockTime:   time.Now(),
					TxnHash:     "txhash2",
				},
			},
			expectedNumPublishPairCalls:             12,
			expectedNumPublishPairCallsWithCreation: 2,
		},
		{
			name: "publish pool pairs with four-denom pool creations",
			createdPoolIDs: map[uint64]commondomain.PoolCreation{
				2: {
					PoolId:      2,
					BlockHeight: 12345,
					BlockTime:   time.Now(),
					TxnHash:     "txhash1",
				},
			},
			expectedNumPublishPairCalls:             12,
			expectedNumPublishPairCallsWithCreation: 6,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			// pool id 1 (2 denoms) created below
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			// pool id 2 (4 denoms), 3 (3 denoms) created below
			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			// pool id 4 (2 denoms), 5 (2 denoms) created below
			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)
			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			// Mock out publisher
			publisherMock := &indexermocks.PublisherMock{}

			pairPublisher := blockprocessor.NewPairPublisher(publisherMock, s.App.PoolManagerKeeper)

			err = pairPublisher.PublishPoolPairs(s.Ctx, blockPools.GetAll(), test.createdPoolIDs)
			s.Require().NoError(err)
			s.Require().Equal(test.expectedNumPublishPairCalls, publisherMock.NumPublishPairCalls)
			s.Require().Equal(test.expectedNumPublishPairCallsWithCreation, publisherMock.NumPublishPairCallsWithCreation)

		})
	}
}
