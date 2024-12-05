package service_test

import (
	"errors"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/service"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v28/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v28/ingest/common/pooltracker"
)

var (
	// emptyWriteListeners is a map of store keys to write listeners.
	// The write listeners are irrelevant for the tests of the sqs service
	// since the service does not use them directly other than storing and returning
	// via getter. As a result, we wire empty write listeners for the tests.
	emptyWriteListeners = make(map[storetypes.StoreKey][]commondomain.WriteListener)

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

type SQSServiceTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestSQSServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SQSServiceTestSuite))
}

// This test validates that the service can recover from an error or panic
// when processing block data.
// It checks that an internal flag is set to true if an error or panic occurs
// and that the pool tracker is reset.
// If no error or panic occurs, the flag should be set to false while pool tracker still reset.
func (s *SQSServiceTestSuite) TestProcessBlockRecoverError() {
	testCases := []struct {
		name                   string
		mockTransformError     error
		mockNilGRPCClientPanic bool

		expectedError error
	}{
		{
			name: "happy path",
		},

		{
			name:               "mock error in processing",
			mockTransformError: mockError,

			expectedError: mockError,
		},
		{
			name:                   "mock panic in processing due to nil grpc client",
			mockNilGRPCClientPanic: true,

			expectedError: errors.New("runtime error: invalid memory address or nil pointer dereference"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

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

			// Initialize the mocks that are not relevant to the test
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}

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

			poolTransformerMock := &mocks.PoolsTransformerMock{
				PoolReturn: transformedPools,
				ErrReturn:  tc.mockTransformError,
			}

			// Trigger a specific error or panic by setting the grpc client to nil
			grpcClientMocks := []domain.SQSGRPClient{
				&mocks.GRPCClientMock{},
				&mocks.GRPCClientMock{},
			}

			for i, _ := range grpcClientMocks {
				if tc.mockNilGRPCClientPanic {
					grpcClientMocks[i] = nil
				}
			}

			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			blockUpdatesProcessUtilsMock := &mocks.BlockUpdateProcessUtilsMock{}

			for _, grpcClientMock := range grpcClientMocks {
				// System under test.
				sqsStreamingService := service.New(blockUpdatesProcessUtilsMock, poolExtractorMock, poolTransformerMock, poolTracker, grpcClientMock, blockProcessStrategyManager, nodeStatusCheckerMock)
				err = sqsStreamingService.ProcessBlockRecoverError(s.Ctx)

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

			}
		})
	}
}
