package blockprocessor

import (
	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
)

func NewBlockUpdatesIndexerBlockProcessStrategy(blockUpdateProcessUtils commondomain.BlockUpdateProcessUtilsI, client domain.Publisher, poolExtractor commondomain.PoolExtractor, poolPairPublisher domain.PairPublisher) *blockUpdatesIndexerBlockProcessStrategy {
	return &blockUpdatesIndexerBlockProcessStrategy{
		blockUpdateProcessUtils: blockUpdateProcessUtils,
		client:                  client,
		poolExtractor:           poolExtractor,
		poolPairPublisher:       poolPairPublisher,
	}
}
