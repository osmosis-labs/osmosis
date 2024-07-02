package blockprocessor

import (
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

// NewBlockProcessor creates a new block process strategy.
func NewBlockProcessor(blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, poolExtracter commondomain.PoolExtracter, keepers domain.Keepers) commondomain.BlockProcessor {
	// If true, ingest all the data.
	if blockProcessStrategyManager.ShouldPushAllData() {

		blockProcessStrategyManager.MarkInitialDataIngested()

		return &fullIndexerBlockProcessStrategy{
			client:        client,
			keepers:       keepers,
			poolExtracter: poolExtracter,
		}
	}

	return &blockUpdatesIndexerBlockProcessStrategy{
		client:        client,
		poolExtracter: poolExtracter,
	}
}
