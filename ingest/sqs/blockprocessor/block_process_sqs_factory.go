package blockprocessor

import (
	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

// NewBlockProcessor creates a new block process strategy.
func NewBlockProcessor(blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.SQSGRPClient, poolExtracter commondomain.PoolExtracter, poolsTransformer domain.PoolsTransformer, nodeStatusChecker domain.NodeStatusChecker) commondomain.BlockProcessor {
	// If true, ingest all the data.
	if blockProcessStrategyManager.ShouldPushAllData() {

		blockProcessStrategyManager.MarkInitialDataIngested()

		return &fullIndexerBlockProcessStrategy{
			sqsGRPCClient:     client,
			poolExtracter:     poolExtracter,
			poolsTransformer:  poolsTransformer,
			nodeStatusChecker: nodeStatusChecker,
		}
	}

	return &blockUpdatesSQSBlockProcessStrategy{
		sqsGRPCCLient:    client,
		poolExtracter:    poolExtracter,
		poolsTransformer: poolsTransformer,
	}
}
