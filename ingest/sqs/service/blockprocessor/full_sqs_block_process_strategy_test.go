package blockprocessor_test

import (
	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v26/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v26/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v26/ingest/sqs/service/blockprocessor"
)

// TestProcessBlock tests the PublishChangedPools method by
// mocking the dependencies and asserting the expected behavior.
func (s *SQSBlockProcessorTestSuite) TestProcessBlock_FullBlockProcessStrategy() {
	tests := []struct {
		name string

		extractorBlockPools       commondomain.BlockPools
		extractorAllDataError     error
		transformAndLoadMockError error
		isSyncingMockValue        bool
		isSyncingMockError        error

		expectedError error
	}{
		{
			name: "happy path",

			extractorBlockPools:       emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        false,

			expectedError: nil,
		},
		{
			name: "transformAndLoadFunc error",

			extractorBlockPools:       emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: defaultError,
			isSyncingMockValue:        false,

			expectedError: defaultError,
		},
		{
			name: "extractor error",

			extractorBlockPools:       emptyBlockPools,
			extractorAllDataError:     defaultError,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        false,

			expectedError: defaultError,
		},
		{
			name: "error: syncing",

			extractorBlockPools:       emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        true,

			expectedError: domain.ErrNodeIsSyncing,
		},

		{
			name: "error: checking node sync status",

			extractorBlockPools:       emptyBlockPools,
			extractorAllDataError:     nil,
			transformAndLoadMockError: nil,
			isSyncingMockValue:        false,
			isSyncingMockError:        defaultError,

			expectedError: &domain.NodeSyncCheckError{
				Err: defaultError,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Mock out pool extractor
			poolsExtracter := &commonmocks.PoolsExtractorMock{
				BlockPools:        tt.extractorBlockPools,
				AllBlockDataError: tt.extractorAllDataError,
			}

			// Initialized transformAndLoadFunc mock
			transformAndLoadMock := blockprocessor.TransformAndLoadFuncMock{
				Error: tt.transformAndLoadMockError,
			}

			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{
				IsSyncing:          tt.isSyncingMockValue,
				IsNodeSyncingError: tt.isSyncingMockError,
			}

			// System under test
			newBlockProcessor := blockprocessor.NewFullBlockSQSBlockProcessStrategy(uninitialzedGRPClient, uninitializedTransformer, poolsExtracter, nodeStatusCheckerMock, transformAndLoadMock.TransformAndLoad)

			// Sanity check
			s.Require().True(newBlockProcessor.IsFullBlockProcessor())

			// Check if the block processor is a full block processor
			actualErr := newBlockProcessor.ProcessBlock(s.Ctx)
			s.Require().Equal(tt.expectedError, actualErr)

			// Validate the transformAndLoadFunc mock
			expectPreTransformError := tt.extractorAllDataError != nil || tt.isSyncingMockError != nil || tt.isSyncingMockValue
			s.validateTransformAndLoadFuncMock(expectPreTransformError, tt.extractorBlockPools, transformAndLoadMock, uninitializedTransformer, uninitialzedGRPClient)
		})
	}
}
