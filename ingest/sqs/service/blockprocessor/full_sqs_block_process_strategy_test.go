package blockprocessor_test

import (
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v25/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/service/blockprocessor"
)

// TestProcessBlock tests the PublishChangedPools method by
// mocking the dependencies and asserting the expected behavior.
func (s *SQSBlockProcessorTestSuite) TestProcessBlock_FullBlockProcessStrategy() {
	tests := []struct {
		name string

		exractorBlockPools        commondomain.BlockPools
		extractorAllDataError     error
		transformAndLoadMockError error
		isSyncingMockValue        bool
		isSynchingMockError       error

		expectedError error
	}{
		{
			name: "happy path",

			exractorBlockPools:        emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        false,

			expectedError: nil,
		},
		{
			name: "transformAndLoadFunc error",

			exractorBlockPools:        emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: defaultError,
			isSyncingMockValue:        false,

			expectedError: defaultError,
		},
		{
			name: "extractor error",

			exractorBlockPools:        emptyBlockPools,
			extractorAllDataError:     defaultError,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        false,

			expectedError: defaultError,
		},
		{
			name: "error: syncing",

			exractorBlockPools:        emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        true,

			expectedError: domain.ErrNodeIsSyncing,
		},
		{
			name: "error: syncing",

			exractorBlockPools:        emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        true,

			expectedError: domain.ErrNodeIsSyncing,
		},

		{
			name: "error: checking node sync status",

			exractorBlockPools:        emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        false,
			isSynchingMockError:       defaultError,

			expectedError: &domain.NodeSyncCheckError{
				Err: defaultError,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			s.Setup()

			// Mock out pool extractor
			poolsExtracter := &commonmocks.PoolsExtractorMock{
				BlockPools:        tt.exractorBlockPools,
				AllBlockDataError: tt.extractorAllDataError,
			}

			// Initialized transformAndLoadFunc mock
			transformAndLoadMock := blockprocessor.TransformAndLoadFuncMock{
				Error: tt.transformAndLoadMockError,
			}

			nodeStatusCheckerMock := &mocks.NodeStatusCheckerMock{
				IsSyncing:          tt.isSyncingMockValue,
				IsNodeSyncingError: tt.isSynchingMockError,
			}

			// System under test
			newBlockProcessor := blockprocessor.NewFullBlockSQSBlockProcessStrategy(uninitialzedGRPClient, uninitializedTransformer, poolsExtracter, nodeStatusCheckerMock, transformAndLoadMock.TransformAndLoad)

			// Sanity check
			s.Require().True(newBlockProcessor.IsFullBlockProcessor())

			// Check if the block processor is a full block processor
			actualErr := newBlockProcessor.ProcessBlock(s.Ctx)
			s.Require().Equal(tt.expectedError, actualErr)

			// Validate the transformAndLoadFunc mock
			expectPreTransformError := tt.extractorAllDataError != nil || tt.isSynchingMockError != nil || tt.isSyncingMockValue
			s.validateTransformAndLoadFuncMock(expectPreTransformError, tt.exractorBlockPools, transformAndLoadMock, uninitializedTransformer, uninitialzedGRPClient, tt.exractorBlockPools)
		})
	}
}
