package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/service/blockprocessor"
)

var _ baseapp.StreamingService = (*sqsStreamingService)(nil)

// sqsStreamingService is a streaming service that processes block data and ingests it into SQS.
// It does so by either processing the entire block data or only the pools that were changed in the block.
// The service uses a pool tracker to keep track of the pools that were changed in the block.
type sqsStreamingService struct {
	writeListeners              map[storetypes.StoreKey][]storetypes.WriteListener
	grpcClient                  domain.SQSGRPClient
	poolsExtractor              commondomain.PoolExtractor
	poolsTransformer            domain.PoolsTransformer
	poolTracker                 domain.BlockPoolUpdateTracker
	blockProcessStrategyManager commondomain.BlockProcessStrategyManager

	nodeStatusChecker domain.NodeStatusChecker
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, poolsExtractor commondomain.PoolExtractor, poolsTransformer domain.PoolsTransformer, poolTracker domain.BlockPoolUpdateTracker, grpcClient domain.SQSGRPClient, blockProcessStrategyManager commondomain.BlockProcessStrategyManager, nodeStatusChecker domain.NodeStatusChecker) *sqsStreamingService {
	return &sqsStreamingService{
		writeListeners:              writeListeners,
		poolsExtractor:              poolsExtractor,
		poolsTransformer:            poolsTransformer,
		poolTracker:                 poolTracker,
		nodeStatusChecker:           nodeStatusChecker,
		grpcClient:                  grpcClient,
		blockProcessStrategyManager: blockProcessStrategyManager,
	}
}

// Close implements baseapp.StreamingService.
func (s *sqsStreamingService) Close() error {
	return nil
}

// ListenBeginBlock implements baseapp.StreamingService.
func (s *sqsStreamingService) ListenBeginBlock(ctx context.Context, req types.RequestBeginBlock, res types.ResponseBeginBlock) error {
	return nil
}

// ListenCommit implements baseapp.StreamingService.
func (s *sqsStreamingService) ListenCommit(ctx context.Context, res types.ResponseCommit) error {
	return nil
}

// ListenDeliverTx implements baseapp.StreamingService.
func (s *sqsStreamingService) ListenDeliverTx(ctx context.Context, req types.RequestDeliverTx, res types.ResponseDeliverTx) error {
	return nil
}

func (s *sqsStreamingService) ListenEndBlock(ctx context.Context, req types.RequestEndBlock, res types.ResponseEndBlock) error {
	blockProcessStartTime := time.Now()
	defer func() {
		// Emit telemetry for the duration of processing the block.
		telemetry.MeasureSince(blockProcessStartTime, domain.SQSProcessBlockDurationMetricName)
	}()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Always return nil to avoid making this consensus breaking.
	_ = s.processBlockRecoverError(sdkCtx)
	return nil
}

// processBlockRecoverError processes the block data and ingests it into SQS. Recovers from panics and returns them as errors.
// It utilizes blockProcessStrategyManager to determine if the block data should be processed in full.
// It resets the pool tracker after processing the block data.
// It notifies blockProcessStrategyManager if a panic or an error occurs while processing the block data.
// -It processes full block data in the following cases:
// - Cold start. We read the entire block data from the chain to push it into the sink.
// - An error occurred while processing the block data in the previous block. To avoid data loss,
// we reprocess the entire block data.
//
// It processes only the pools that were changed in the block in the following cases:
// - The node is not in cold start and the previous block was processed successfully.
func (s *sqsStreamingService) processBlockRecoverError(ctx sdk.Context) (err error) {
	defer func() {
		// Reset pool tracking for this block.
		s.poolTracker.Reset()

		if r := recover(); r != nil {
			// Due to panic, we set shouldProcessAllBlockData to true to reprocess the entire block.
			// Be careful when changing this behavior.
			s.blockProcessStrategyManager.MarkErrorObserved()

			// Emit telemetry for the panic.
			emitFailureTelemetry(ctx, r, domain.SQSProcessBlockPanicMetricName)

			err = fmt.Errorf("panic: %v", r)
		}

		if err == nil {
			// If no error or panic occurred, mark the data as ingested
			// so that the next block processes only the pools that were changed.
			s.blockProcessStrategyManager.MarkInitialDataIngested()
		}
	}()

	blockProcessor := blockprocessor.NewBlockProcessor(s.blockProcessStrategyManager, s.grpcClient, s.poolsExtractor, s.poolsTransformer, s.nodeStatusChecker)

	if err := blockProcessor.ProcessBlock(ctx); err != nil {
		// Due to error, we set shouldProcessAllBlockData to true to reprocess the entire block.
		// Be careful when changing this behavior.
		s.blockProcessStrategyManager.MarkErrorObserved()

		// Emit telemetry for the error.
		emitFailureTelemetry(ctx, err, domain.SQSProcessBlockErrorMetricName)

		return err
	}

	return nil
}

// Listeners implements baseapp.StreamingService.
func (s *sqsStreamingService) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener {
	return s.writeListeners
}

// Stream implements baseapp.StreamingService.
func (s *sqsStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}

// emitFailureTelemetry emits telemetry for panics or errors
func emitFailureTelemetry(ctx sdk.Context, r interface{}, metricName string) {
	// Panics are silently logged and ignored.
	ctx.Logger().Error(metricName, "err", r)

	// Emit telemetry for the panic.
	telemetry.IncrCounterWithLabels([]string{metricName}, 1, []metrics.Label{
		{Name: "height", Value: fmt.Sprintf("%d", ctx.BlockHeight())},
		{Name: "msg", Value: fmt.Sprintf("%v", r)},
	})
}
