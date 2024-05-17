package service_test

import (
	"errors"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/service"
)

var (
	// emptyWriteListeners is a map of store keys to write listeners.
	// The write listeners are irrelevant for the tests of the sqs service
	// since the service does not use them directly other than storing and returning
	// via getter. As a result, we wire empty write listeners for the tests.
	emptyWriteListeners = make(map[storetypes.StoreKey][]domain.WriteListener)
	emptyStoreKeyMap    = make(map[string]storetypes.StoreKey)

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

// This is a sanity-check test ensuring that all pools are tracked and returned correctly.
func (s *SQSServiceTestSuite) TestProcessBlock() {
	testCases := []struct {
		name string
		// Flag that we preset to the service to process all block data if true.
		shouldProcessAllBlockData bool

		// Test parameters
		// flag indicating if ProcessAllBlockData should return an error.
		doesProcessAllBlockDataError bool
		// flag indicating if ProcessChangedBlockData should return an error.
		doesProcessChangedBlockDataError bool

		// flag indicating if the node is syncing.
		isSyncing bool
		// flag indicating if IsNodeSyncing should return an error.
		doesIsNodeSyncingError bool

		expectedError error
	}{
		{

			name:                      "node is syncing while processing all block data - returns error",
			shouldProcessAllBlockData: true,
			isSyncing:                 true,

			expectedError: domain.ErrNodeIsSyncing,
		},
		{
			// The reason this does not error is so that if node falls behind a few blocks, we do not
			// want to retrigger the processing of all block data and instead let the node catch up.
			name:      "node is syncing while processing changed block data - unaffected",
			isSyncing: true,
		},
		{
			name:                      "processing all block data",
			shouldProcessAllBlockData: true,
		},
		{
			name: "processing changed block data",
		},
		{
			name:                      "processing all block data and error occurs",
			shouldProcessAllBlockData: true,

			doesProcessAllBlockDataError: true,

			expectedError: mockError,
		},
		{
			name:                             "processing changed block data and error occurs",
			doesProcessChangedBlockDataError: true,

			expectedError: mockError,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// Prepare each pool for testing
			allPools := s.PrepareAllSupportedPools()

			sqsIngesterMock := &mocks.SQSIngesterMock{
				AllBlockDataError:     mockErrorFromFlag(tc.doesProcessAllBlockDataError),
				ChangedBlockDataError: mockErrorFromFlag(tc.doesProcessChangedBlockDataError),
			}
			nodeStatusCheckerMock := &mocks.NodeStatusCheckerMock{
				IsSyncing:          tc.isSyncing,
				IsNodeSyncingError: mockErrorFromFlag(tc.doesIsNodeSyncingError),
			}

			poolTracker := service.NewPoolTracker()
			// Get balancer pool from chain and track it, making the pool tracker
			// have one pool.
			balancerPool, err := s.App.GAMMKeeper.GetCFMMPool(s.Ctx, allPools.BalancerPoolID)
			s.Require().NoError(err)
			poolTracker.TrackCFMM(balancerPool)
			s.Require().Equal(len(poolTracker.GetCFMMPools()), 1)

			sqsStreamingServiceI := service.New(emptyWriteListeners, emptyStoreKeyMap, sqsIngesterMock, poolTracker, nodeStatusCheckerMock)

			// cast the interface to the concrete type for testing unexported concrete method.
			sqsStreamingService, ok := sqsStreamingServiceI.(*service.SQSStreamingService)
			s.Require().True(ok)

			// Set the flag to process all block data.
			sqsStreamingService.SetShouldProcessAllBlockData(tc.shouldProcessAllBlockData)

			// System under test
			err = sqsStreamingService.ProcessBlock(s.Ctx)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				return
			} else {
				s.Require().NoError(err)
			}

			// We expect only one process method to be called based on the flag.
			s.Require().Equal(tc.shouldProcessAllBlockData, sqsIngesterMock.IsProcessAllBlockDataCalled)
			s.Require().Equal(!tc.shouldProcessAllBlockData, sqsIngesterMock.IsProcessAllChangedDataCalled)

			// if processing changed block data, we can also assert
			// that the pools were tracked correctly.
			if !tc.shouldProcessAllBlockData {
				s.Require().True(sqsIngesterMock.IsProcessAllChangedDataCalled)

				// This is how we configure the test universally where we make the pool tracker only track one pool.
				// As a result, this is what we expect to see propagated to the mock.
				s.Require().Equal(len(poolTracker.GetCFMMPools()), len(sqsIngesterMock.LastChangedPoolsObserved.CFMMPools))
				s.Require().Empty(sqsIngesterMock.LastChangedPoolsObserved.ConcentratedPools)
				s.Require().Empty(sqsIngesterMock.LastChangedPoolsObserved.CosmWasmPools)
				s.Require().Empty(sqsIngesterMock.LastChangedPoolsObserved.ConcentratedPoolIDTickChange)
			}

			// In all cases, we expect the shouldProcessAllBlockData flag to be set to false
			// after the block is processed.
			s.Require().False(sqsStreamingService.GetShouldProcessAllBlockData())
		})
	}
}

// This test validates that the service can recover from an error or panic
// when processing block data.
// It checks that an internal flag is set to true if an error or panic occurs
// and that the pool tracker is reset.
// If no error or panic occurs, the flag should be set to false while pool tracker still reset.
func (s *SQSServiceTestSuite) TestProcessBlockRecoverError() {
	testCases := []struct {
		name string

		// Test parameters
		// flag indicating if ProcessAllBlockData should return an error.
		doesProcessAllBlockDataError bool
		processAllBlockDataPanicMsg  string

		expectedError error
	}{
		{
			name: "happy path",
		},
		{
			name:                         "error occurs - sets shouldProcessAllBlockData flag to true",
			doesProcessAllBlockDataError: true,

			expectedError: mockError,
		},
		{
			name:                        "panic occurs - sets shouldProcessAllBlockData flag to true",
			processAllBlockDataPanicMsg: mockError.Error(),
			expectedError:               mockError,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// Prepare each pool for testing
			allPools := s.PrepareAllSupportedPools()

			sqsIngesterMock := &mocks.SQSIngesterMock{
				AllBlockDataError:           mockErrorFromFlag(tc.doesProcessAllBlockDataError),
				ProcessAllBlockDataPanicMsg: tc.processAllBlockDataPanicMsg,
			}
			nodeStatusCheckerMock := &mocks.NodeStatusCheckerMock{}

			poolTracker := service.NewPoolTracker()
			// Get balancer pool from chain and track it, making the pool tracker
			// have one pool.
			balancerPool, err := s.App.GAMMKeeper.GetCFMMPool(s.Ctx, allPools.BalancerPoolID)
			s.Require().NoError(err)
			poolTracker.TrackCFMM(balancerPool)
			s.Require().Equal(len(poolTracker.GetCFMMPools()), 1)

			sqsStreamingServiceI := service.New(emptyWriteListeners, emptyStoreKeyMap, sqsIngesterMock, poolTracker, nodeStatusCheckerMock)

			// cast the interface to the concrete type for testing unexported concrete method.
			sqsStreamingService, ok := sqsStreamingServiceI.(*service.SQSStreamingService)
			s.Require().True(ok)

			// Set the flag to always process all block data for simplicity.
			sqsStreamingService.SetShouldProcessAllBlockData(true)

			err = sqsStreamingService.ProcessBlockRecoverError(s.Ctx)

			// We expect the pool tracker to always be reset
			s.Require().Empty(poolTracker.GetCFMMPools())
			s.Require().Empty(poolTracker.GetConcentratedPools())
			s.Require().Empty(poolTracker.GetCosmWasmPools())
			s.Require().Empty(poolTracker.GetConcentratedPoolIDTickChange())

			// We expect the shouldProcessAllBlockData flag to be set to true due to error/panic
			s.Require().Equal(tc.expectedError != nil, sqsStreamingService.GetShouldProcessAllBlockData())

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())

				return
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
