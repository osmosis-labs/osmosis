package blockprocessor_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v28/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/service/blockprocessor"
)

type SQSBlockProcessorTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestSQSServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SQSBlockProcessorTestSuite))
}

func (suite *SQSBlockProcessorTestSuite) TestNewBlockProcessor() {

	tests := []struct {
		name string

		mockInitialDataIngested bool

		expectedIsFullBlockProcessor bool
	}{
		{
			name:                         "full block processor",
			mockInitialDataIngested:      true,
			expectedIsFullBlockProcessor: false,
		},
		{
			name:                         "updates block processor",
			mockInitialDataIngested:      false,
			expectedIsFullBlockProcessor: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {

			// Initialize mock inputs that do not affect the test
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}
			poolsExtracter := &commonmocks.PoolsExtractorMock{}
			poolsTransformer := &mocks.PoolsTransformerMock{}
			grpcClientMock := &mocks.GRPCClientMock{}
			blockUpdatesProcessUtilsMock := &mocks.BlockUpdateProcessUtilsMock{}

			// Initialize the block strategy manager
			blockStrategyManager := commondomain.NewBlockProcessStrategyManager()
			if tt.mockInitialDataIngested {
				blockStrategyManager.MarkInitialDataIngested()
			}

			// System under test
			newBlockProcessor := blockprocessor.NewBlockProcessor(blockStrategyManager, grpcClientMock, poolsExtracter, poolsTransformer, nodeStatusCheckerMock, blockUpdatesProcessUtilsMock)

			// Check if the block processor is a full block processor
			isFullBlockProcessor := newBlockProcessor.IsFullBlockProcessor()
			suite.Require().Equal(tt.expectedIsFullBlockProcessor, isFullBlockProcessor)
		})
	}
}
