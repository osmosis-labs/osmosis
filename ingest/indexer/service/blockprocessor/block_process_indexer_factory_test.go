package blockprocessor_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v31/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v31/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v31/ingest/indexer/domain/mocks"
	"github.com/osmosis-labs/osmosis/v31/ingest/indexer/service/blockprocessor"
)

type IndexerBlockProcessorTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestIndexerBlockProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(IndexerBlockProcessorTestSuite))
}

func (suite *IndexerBlockProcessorTestSuite) TestNewBlockProcessor() {

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
			poolsExtracter := &commonmocks.PoolsExtractorMock{}
			publisherMock := &mocks.PublisherMock{}

			// Initialize the block strategy manager
			blockStrategyManager := commondomain.NewBlockProcessStrategyManager()
			if tt.mockInitialDataIngested {
				blockStrategyManager.MarkInitialDataIngested()
			}

			// Initialize the node status checker mock
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}

			// System under test
			newBlockProcessor := blockprocessor.NewBlockProcessor(blockStrategyManager, publisherMock, poolsExtracter, domain.Keepers{}, nodeStatusCheckerMock, nil)

			// Check if the block processor is a full block processor
			isFullBlockProcessor := newBlockProcessor.IsFullBlockProcessor()
			suite.Require().Equal(tt.expectedIsFullBlockProcessor, isFullBlockProcessor)
		})
	}
}
