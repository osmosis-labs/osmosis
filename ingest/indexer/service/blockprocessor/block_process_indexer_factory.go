package blockprocessor

import (
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

// NewBlockProcessor creates a new block process strategy.
func NewBlockProcessor(blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, poolExtractor commondomain.PoolExtractor, keepers domain.Keepers) commondomain.BlockProcessor {
	// Initialize the pool pair publisher
	poolPairPublisher := NewPairPublisher(client, keepers.PoolManagerKeeper)

	// If true, ingest all the data.
	if blockProcessStrategyManager.ShouldPushAllData() {
		// TODO: move this to a higher level component
		blockProcessStrategyManager.MarkInitialDataIngested()

		return &fullIndexerBlockProcessStrategy{
			client:            client,
			keepers:           keepers,
			poolExtractor:     poolExtractor,
			poolPairPublisher: poolPairPublisher,
		}
	}

	return &blockUpdatesIndexerBlockProcessStrategy{
		client:            client,
		poolExtractor:     poolExtractor,
		poolPairPublisher: poolPairPublisher,
	}
}
