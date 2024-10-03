package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	commonservice "github.com/osmosis-labs/osmosis/v26/ingest/common/service"
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

type BlockUpdatesIndexerBlockProcessStrategy = blockUpdatesIndexerBlockProcessStrategy

func (s *blockUpdatesIndexerBlockProcessStrategy) PublishCreatedPools(ctx types.Context) error {
	return s.publishCreatedPools(ctx)
}

func NewFullIndexerBlockProcessStrategy(client domain.Publisher, keepers domain.Keepers, poolExtractor commondomain.PoolExtractor, poolPairPublisher domain.PairPublisher, nodeStatusChecker commonservice.NodeStatusChecker) *fullIndexerBlockProcessStrategy {
	return &fullIndexerBlockProcessStrategy{
		client:            client,
		keepers:           keepers,
		poolExtractor:     poolExtractor,
		poolPairPublisher: poolPairPublisher,
		nodeStatusChecker: nodeStatusChecker,
	}
}

type FullIndexerBlockProcessStrategy = fullIndexerBlockProcessStrategy

func (s *fullIndexerBlockProcessStrategy) PublishAllSupplies(ctx types.Context) {
	s.publishAllSupplies(ctx)
}

func (s *fullIndexerBlockProcessStrategy) ProcessPools(ctx types.Context) error {
	return s.processPools(ctx)
}
