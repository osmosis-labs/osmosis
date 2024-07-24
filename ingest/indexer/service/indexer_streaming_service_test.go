package service_test

import (
	"errors"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	indexerservice "github.com/osmosis-labs/osmosis/v25/ingest/indexer/service"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain/mocks"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/common/pooltracker"
)

var (
	emptyStoreKeyMap = make(map[string]storetypes.StoreKey)

	// mockError is a mock error for testing.
	mockError = errors.New("mock error")

	// mockErrorFromFlag returns a mock error if shouldError is true.
	mockErrorFromFlag = func(shouldError bool) error {
		if shouldError {
			return mockError
		}
		return nil
	}
)

type IndexerServiceTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestIndexerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IndexerServiceTestSuite))
}

func (s *IndexerServiceTestSuite) TestAdjustTokenInAmountBySpreadFactor() {
	testCases := []struct {
		name          string
		expectedError error
	}{
		{
			name: "happy path",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()
		})
	}
}

func (s *IndexerServiceTestSuite) TestAddTokenLiquidity() {
	testCases := []struct {
		name                   string
		mockTransformError     error
		mockNilGRPCClientPanic bool
		expectedError          error
	}{
		{
			name: "happy path",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)

			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)

			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)

			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			transformedPools := []sqsdomain.PoolI{}
			for _, pool := range blockPools.GetAll() {
				// Note: balances are irrelevant for the test so we supply empty balances
				transformedPool := sqsdomain.NewPool(pool, pool.GetSpreadFactor(s.Ctx), sdk.Coins{})
				transformedPools = append(transformedPools, transformedPool)
			}

			poolTracker := pooltracker.NewMemory()
			// Add some pools to the tracker.
			// We simply want to assert that the tracker is reset after the block processing.
			for _, pool := range blockPools.ConcentratedPools {
				poolTracker.TrackConcentrated(pool)
			}

			poolExtractorMock := &mocks.PoolsExtractorMock{
				BlockPools: commondomain.BlockPools{
					ConcentratedPools: concentratedPools,
					CFMMPools:         cfmmPools,
					CosmWasmPools:     cosmWasmPools,
				},
			}

			publisherMock := &mocks.PublisherMock{}

			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			blockUpdatesProcessUtilsMock := &mocks.BlockUpdateProcessUtilsMock{}

			keepers := indexerdomain.Keepers{
				BankKeeper:        s.App.BankKeeper,
				PoolManagerKeeper: s.App.PoolManagerKeeper,
			}

			txDecoder := s.App.GetTxConfig().TxDecoder()
			logger := s.App.Logger()

			// indexerStreamingService := indexerservice.New(blockUpdatesProcessUtilsMock, blockProcessStrategyManager, publisherMock, emptyStoreKeyMap, poolExtractorMock, poolTracker, keepers, txDecoder, logger)
			_ = indexerservice.New(blockUpdatesProcessUtilsMock, blockProcessStrategyManager, publisherMock, emptyStoreKeyMap, poolExtractorMock, poolTracker, keepers, txDecoder, logger)

			// System under test.
			// err = sqsStreamingService.ProcessBlockRecoverError(s.Ctx)

			// We expect the pool tracker to always be reset
			s.Require().Empty(poolTracker.GetCFMMPools())
			s.Require().Empty(poolTracker.GetConcentratedPools())
			s.Require().Empty(poolTracker.GetCosmWasmPools())
			s.Require().Empty(poolTracker.GetConcentratedPoolIDTickChange())

			// Pool Tracker is reset
			// Note: we only initialized it with concentrated pools above
			s.Require().Empty(poolTracker.GetConcentratedPools())

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())

				// Validate that the block processing strategy is set to push all data
				// due to error or panic
				s.Require().True(blockProcessStrategyManager.ShouldPushAllData())
			} else {
				s.Require().NoError(err)

				// Validate that the block processing strategy is now set to only process updates
				s.Require().False(blockProcessStrategyManager.ShouldPushAllData())
			}
		})
	}
}
