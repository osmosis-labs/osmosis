package blockprocessor

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-metrics"

	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v28/ingest/sqs/domain"
)

type fullSQSBlockProcessStrategy struct {
	sqsGRPCClient domain.SQSGRPClient

	poolExtractor    commondomain.PoolExtractor
	poolsTransformer domain.PoolsTransformer

	nodeStatusChecker domain.NodeStatusChecker

	transformAndLoadFunc transformAndLoadFunc
}

// IsFullBlockProcessor implements commondomain.BlockProcessor.
func (f *fullSQSBlockProcessStrategy) IsFullBlockProcessor() bool {
	return true
}

var _ commondomain.BlockProcessor = &fullSQSBlockProcessStrategy{}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *fullSQSBlockProcessStrategy) ProcessBlock(ctx sdk.Context) (err error) {
	// Detect syncing
	isNodesyncing, err := f.nodeStatusChecker.IsNodeSyncing(ctx)
	if err != nil {
		telemetry.IncrCounterWithLabels([]string{domain.SQSNodeSyncCheckErrorMetricName}, 1, []metrics.Label{
			{Name: "err", Value: err.Error()},
			{Name: "height", Value: fmt.Sprintf("%d", ctx.BlockHeight())},
		})
		return &commondomain.NodeSyncCheckError{Err: err}
	}
	if isNodesyncing {
		return commondomain.ErrNodeIsSyncing
	}

	pools, _, err := f.poolExtractor.ExtractAll(ctx)
	if err != nil {
		return err
	}

	// Publish the pools
	err = f.transformAndLoadFunc(ctx, f.poolsTransformer, f.sqsGRPCClient, pools)
	if err != nil {
		return err
	}

	return nil
}
