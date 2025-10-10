package blockprocessor

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v31/ingest/sqs/domain"
)

type (
	BlockUpdatesSQSBlockProcessStrategy = blockUpdatesSQSBlockProcessStrategy
	FullBlockSQSBlockProcessStrategy    = fullSQSBlockProcessStrategy
	TransformAndLoadFunc                = transformAndLoadFunc
)

func NewBlockUpdatesSQSBlockProcessStrategy(blockUpdateProcessUtils commondomain.BlockUpdateProcessUtilsI, sqsGRPCClient domain.SQSGRPClient, poolsTransformer domain.PoolsTransformer, poolExtractor commondomain.PoolExtractor, transformAndLoadFunc transformAndLoadFunc) *BlockUpdatesSQSBlockProcessStrategy {
	return &blockUpdatesSQSBlockProcessStrategy{
		sqsGRPCClient: sqsGRPCClient,

		poolsTransformer: poolsTransformer,
		poolExtractor:    poolExtractor,

		transformAndLoadFunc: transformAndLoadFunc,

		blockUpdateProcessUtils: blockUpdateProcessUtils,
	}
}

func NewFullBlockSQSBlockProcessStrategy(sqsGRPCCLient domain.SQSGRPClient, poolsTransformer domain.PoolsTransformer, poolExtractor commondomain.PoolExtractor, nodeStatusChecker domain.NodeStatusChecker, transformAndLoadFunc transformAndLoadFunc) *FullBlockSQSBlockProcessStrategy {
	return &fullSQSBlockProcessStrategy{
		sqsGRPCClient: sqsGRPCCLient,

		poolsTransformer: poolsTransformer,
		poolExtractor:    poolExtractor,

		nodeStatusChecker: nodeStatusChecker,

		transformAndLoadFunc: transformAndLoadFunc,
	}
}

type TransformAndLoadFuncMock struct {
	CalledWithTransformer domain.PoolsTransformer
	CalledWithSQSClient   domain.SQSGRPClient
	CalledWithPools       commondomain.BlockPools

	Error error
}

func (m *TransformAndLoadFuncMock) TransformAndLoad(ctx sdk.Context, poolsTrasnformer domain.PoolsTransformer, sqsGRPCClient domain.SQSGRPClient, pools commondomain.BlockPools) error {
	m.CalledWithSQSClient = sqsGRPCClient
	m.CalledWithTransformer = poolsTrasnformer
	m.CalledWithPools = pools

	return m.Error
}
