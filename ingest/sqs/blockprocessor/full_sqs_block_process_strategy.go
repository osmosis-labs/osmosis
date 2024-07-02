package blockprocessor

import (
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

type fullIndexerBlockProcessStrategy struct {
	sqsGRPCClient domain.SQSGRPClient

	poolExtracter    commondomain.PoolExtracter
	poolsTransformer domain.PoolsTransformer

	nodeStatusChecker domain.NodeStatusChecker
}

var _ commondomain.BlockProcessor = &fullIndexerBlockProcessStrategy{}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *fullIndexerBlockProcessStrategy) ProcessBlock(ctx sdk.Context) (err error) {
	// Detect syncing
	isNodesyncing, err := f.nodeStatusChecker.IsNodeSyncing(ctx)
	if err != nil {
		telemetry.IncrCounterWithLabels([]string{domain.SQSNodeSyncCheckErrorMetricName}, 1, []metrics.Label{
			{Name: "err", Value: err.Error()},
			{Name: "height", Value: fmt.Sprintf("%d", ctx.BlockHeight())},
		})
		return fmt.Errorf("failed to check if node is syncing: %w", err)
	}
	if isNodesyncing {
		return fmt.Errorf("node is syncing, skipping block processing")
	}

	pools, err := f.poolExtracter.ExtractAll(ctx)
	if err != nil {
		return err
	}

	// Publish the pools
	err = transformAndLoad(ctx, f.poolsTransformer, f.sqsGRPCClient, pools)
	if err != nil {
		return err
	}

	return nil
}
