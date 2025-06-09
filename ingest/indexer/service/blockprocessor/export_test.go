package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v30/ingest/common/domain"
	commonservice "github.com/osmosis-labs/osmosis/v30/ingest/common/service"
	indexerdomain "github.com/osmosis-labs/osmosis/v30/ingest/indexer/domain"
)

// Alias to BlockUpdatesIndexerBlockProcessStrategy to allow exporting private functions for testing.
type BlockUpdatesIndexerBlockProcessStrategy = blockUpdatesIndexerBlockProcessStrategy

func NewBlockUpdatesIndexerBlockProcessStrategy(blockUpdateProcessUtils commondomain.BlockUpdateProcessUtilsI, client indexerdomain.Publisher, poolExtractor commondomain.PoolExtractor, poolPairPublisher indexerdomain.PairPublisher) *blockUpdatesIndexerBlockProcessStrategy {
	return &blockUpdatesIndexerBlockProcessStrategy{
		blockUpdateProcessUtils: blockUpdateProcessUtils,
		client:                  client,
		poolExtractor:           poolExtractor,
		poolPairPublisher:       poolPairPublisher,
	}
}

func (s *blockUpdatesIndexerBlockProcessStrategy) PublishCreatedPools(ctx types.Context) error {
	return s.publishCreatedPools(ctx)
}

// Alias to FullIndexerBlockProcessStrategy to allow exporting private functions for testing.
type FullIndexerBlockProcessStrategy = fullIndexerBlockProcessStrategy

func NewFullIndexerBlockProcessStrategy(client indexerdomain.Publisher, keepers indexerdomain.Keepers, poolExtractor commondomain.PoolExtractor, poolPairPublisher indexerdomain.PairPublisher, nodeStatusChecker commonservice.NodeStatusChecker) *fullIndexerBlockProcessStrategy {
	return &fullIndexerBlockProcessStrategy{
		client:            client,
		keepers:           keepers,
		poolExtractor:     poolExtractor,
		poolPairPublisher: poolPairPublisher,
		nodeStatusChecker: nodeStatusChecker,
	}
}

func (s *fullIndexerBlockProcessStrategy) PublishAllSupplies(ctx types.Context) {
	s.publishAllSupplies(ctx)
}

func (s *fullIndexerBlockProcessStrategy) ProcessPools(ctx types.Context) error {
	return s.processPools(ctx)
}
