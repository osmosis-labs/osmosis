package blockprocessor_test

import (
	"errors"

	commondomain "github.com/osmosis-labs/osmosis/v29/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v29/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v29/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v29/ingest/sqs/service/blockprocessor"
	"github.com/osmosis-labs/osmosis/v29/x/poolmanager/types"
)

var (
	emptyBlockPools = commondomain.BlockPools{}

	// Create uninitialized mocks to be used where components are irrelevant for testing.
	uninitializedTransformer = &mocks.PoolsTransformerMock{}
	uninitialzedGRPClient    = &mocks.GRPCClientMock{}

	defaultError = errors.New("default error")
)

// TestProcessBlock tests the PublishChangedPools method by
// mocking the dependencies and asserting the expected behavior.
func (s *SQSBlockProcessorTestSuite) TestProcessBlock_UpdatesOnlyStrategy() {

	tests := []struct {
		name string

		exractorBlockPools        commondomain.BlockPools
		extractChangedError       error
		transformAndLoadMockError error
		processChangeSetError     error

		expectedError error
	}{
		{
			name: "happy path",

			exractorBlockPools:        emptyBlockPools,
			extractChangedError:       nil,
			transformAndLoadMockError: nil,

			expectedError: nil,
		},
		{
			name: "transformAndLoadFunc error",

			exractorBlockPools:        emptyBlockPools,
			extractChangedError:       nil,
			transformAndLoadMockError: defaultError,

			expectedError: defaultError,
		},
		{
			name: "extractor error",

			exractorBlockPools:        emptyBlockPools,
			extractChangedError:       defaultError,
			transformAndLoadMockError: nil,

			expectedError: defaultError,
		},
		{
			name: "process change set error",

			exractorBlockPools:        emptyBlockPools,
			extractChangedError:       nil,
			transformAndLoadMockError: nil,
			processChangeSetError:     defaultError,

			expectedError: defaultError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			s.Setup()

			// Mock out pool extractor
			poolsExtracter := &commonmocks.PoolsExtractorMock{
				BlockPools:            tt.exractorBlockPools,
				ChangedBlockDataError: tt.extractChangedError,
			}

			// Initialized transformAndLoadFunc mock
			transformAndLoadMock := blockprocessor.TransformAndLoadFuncMock{
				Error: tt.transformAndLoadMockError,
			}

			// Mock out block updates process utils
			blockUpdatesProcessUtilsMock := &mocks.BlockUpdateProcessUtilsMock{
				ProcessBlockReturn: tt.processChangeSetError,
			}

			// System under test
			newBlockProcessor := blockprocessor.NewBlockUpdatesSQSBlockProcessStrategy(blockUpdatesProcessUtilsMock, uninitialzedGRPClient, uninitializedTransformer, poolsExtracter, transformAndLoadMock.TransformAndLoad)

			// Sanity check
			s.Require().False(newBlockProcessor.IsFullBlockProcessor())

			// Check if the block processor is a full block processor
			actualErr := newBlockProcessor.ProcessBlock(s.Ctx)
			s.Require().Equal(tt.expectedError, actualErr)

			// Validate the transformAndLoadFunc mock
			expectPreTransformError := tt.extractChangedError != nil || tt.processChangeSetError != nil
			s.validateTransformAndLoadFuncMock(expectPreTransformError, tt.exractorBlockPools, transformAndLoadMock, uninitializedTransformer, uninitialzedGRPClient)
		})
	}
}

// validateTransformAndLoadFuncMock validates the transformAndLoadFunc mock
// based on the expected inputs and outputs.
// expectPreTransformError indicates if any component before the transformAndLoadFunc errored.
func (s *SQSBlockProcessorTestSuite) validateTransformAndLoadFuncMock(expectPreTransformError bool, expectedBlockPools commondomain.BlockPools, transformAndLoadMock blockprocessor.TransformAndLoadFuncMock, expectedTransformer *mocks.PoolsTransformerMock, expectedSQSClient *mocks.GRPCClientMock) {
	// If extractor errored, we do not expect transformer mock to be called
	if expectPreTransformError {
		// Note: nil indicates that the method was not called
		s.Require().Nil(transformAndLoadMock.CalledWithTransformer)
		s.Require().Nil(transformAndLoadMock.CalledWithSQSClient)
		// Note: this structure signifes nil pools
		s.Require().Equal(commondomain.BlockPools{ConcentratedPools: []types.PoolI(nil)}, transformAndLoadMock.CalledWithPools)
		return
	}

	// Assert tranformAndLoadFunc is called with the correct inputs
	s.Require().Equal(uninitializedTransformer, transformAndLoadMock.CalledWithTransformer)
	s.Require().Equal(uninitialzedGRPClient, transformAndLoadMock.CalledWithSQSClient)
	s.Require().Equal(expectedBlockPools, transformAndLoadMock.CalledWithPools)
}
