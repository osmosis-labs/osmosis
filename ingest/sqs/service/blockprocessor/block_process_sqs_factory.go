package blockprocessor

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	commonservice "github.com/osmosis-labs/osmosis/v26/ingest/common/service"
	"github.com/osmosis-labs/osmosis/v26/ingest/sqs/domain"
)

type transformAndLoadFunc func(ctx sdk.Context, poolsTrasnformer domain.PoolsTransformer, sqsGRPCClient domain.SQSGRPClient, pools commondomain.BlockPools) error

// NewBlockProcessor creates a new block process strategy.
// If block process strategy manager should push all data, then it will return a full indexer block process strategy.
// Otherwise, it will return a block updates SQS block process strategy.
func NewBlockProcessor(blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.SQSGRPClient, poolExtractor commondomain.PoolExtractor, poolsTransformer domain.PoolsTransformer, nodeStatusChecker commonservice.NodeStatusChecker, blockUpdateProcessUtils commondomain.BlockUpdateProcessUtilsI) commondomain.BlockProcessor {
	// If true, ingest all the data.
	if blockProcessStrategyManager.ShouldPushAllData() {
		return &fullSQSBlockProcessStrategy{
			sqsGRPCClient:     client,
			poolExtractor:     poolExtractor,
			poolsTransformer:  poolsTransformer,
			nodeStatusChecker: nodeStatusChecker,

			transformAndLoadFunc: transformAndLoad,
		}
	}

	return &blockUpdatesSQSBlockProcessStrategy{
		sqsGRPCClient:    client,
		poolExtractor:    poolExtractor,
		poolsTransformer: poolsTransformer,

		transformAndLoadFunc: transformAndLoad,

		blockUpdateProcessUtils: blockUpdateProcessUtils,
	}
}
