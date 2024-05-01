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

	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

var _ baseapp.StreamingService = (*sqsStreamingService)(nil)

// sqsStreamingService is a streaming service that processes block data and ingests it into SQS.
// It does so by either processing the entire block data or only the pools that were changed in the block.
// The service uses a pool tracker to keep track of the pools that were changed in the block.
type sqsStreamingService struct {
	writeListeners map[storetypes.StoreKey][]storetypes.WriteListener
	sqsIngester    domain.Ingester
	poolTracker    domain.BlockPoolUpdateTracker

	nodeStatusChecker domain.NodeStatusChecker

	shouldProcessAllBlockData bool
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, sqsIngester domain.Ingester, poolTracker domain.BlockPoolUpdateTracker, nodeStatusChecker domain.NodeStatusChecker) baseapp.StreamingService {
	return &sqsStreamingService{
		writeListeners:    writeListeners,
		sqsIngester:       sqsIngester,
		poolTracker:       poolTracker,
		nodeStatusChecker: nodeStatusChecker,

		shouldProcessAllBlockData: true,
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
// It controls an internal flag shouldProcessAllBlockData to determine if the block data should be processed in full.
// It resets the pool tracker after processing the block data.
// It sets shouldProcessAllBlockData to true if a panic occurs while processing the block data.
// It sets shouldProcessAllBlockData to true if an error occurs while processing the block data.
// Always returns nil to avoid making this consensus breaking.
// WARNING: this method emits sdk events for testability. Ensure that the caller discards the events.
func (s *sqsStreamingService) processBlockRecoverError(ctx sdk.Context) (err error) {
	defer func() {
		// Reset pool tracking for this block.
		s.poolTracker.Reset()

		if r := recover(); r != nil {
			// Due to panic, we set shouldProcessAllBlockData to true to reprocess the entire block.
			// Be careful when changing this behavior.
			s.shouldProcessAllBlockData = true

			// Emit telemetry for the panic.
			emitFailureTelemetry(ctx, r, domain.SQSProcessBlockPanicMetricName)

			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// Process the block data.
	if err := s.processBlock(ctx); err != nil {
		// Due to error, we set shouldProcessAllBlockData to true to reprocess the entire block.
		// Be careful when changing this behavior.
		s.shouldProcessAllBlockData = true

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

// processBlock processes the block data.
//
// -It processes full block data in the following cases:
// - Cold start. We read the entire block data from the chain to push it into the sink.
// - An error occurred while processing the block data in the previous block. To avoid data loss,
// we reprocess the entire block data.
//
// It processes only the pools that were changed in the block in the following cases:
// - The node is not in cold start and the previous block was processed successfully.
//
// An internal flag shouldProcessAllBlockData is used to determine if the block data should be processed in full.
//
// This method is a no-op in the following two cases:
// - The node is syncing.
// - Fails to determine if the node is syncing.
// The method calls a node's status endpoint to determine if the node is syncing.
//
// Returns error if the block data processing fails.
func (s *sqsStreamingService) processBlock(ctx sdk.Context) error {
	// If cold start, we use SQS ingester to process the entire block.
	if s.shouldProcessAllBlockData {
		// Detect syncing
		isNodesyncing, err := s.nodeStatusChecker.IsNodeSyncing(ctx)
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

		// Process the entire block if the node is caught up
		cwPools, err := s.sqsIngester.ProcessAllBlockData(ctx)
		if err != nil {
			return err
		}

		// Generate the initial cwPool address to pool mapping
		for _, pool := range cwPools {
			s.poolTracker.TrackCosmWasmPoolsAddressToPoolMap(pool)
		}

		// Successfully processed the block, no longer need to process full block data.
		s.shouldProcessAllBlockData = false

		return nil
	}

	// If not cold start, we only process the pools that were changed this block.
	concentratedPools := s.poolTracker.GetConcentratedPools()
	concentratedPoolIDTickChange := s.poolTracker.GetConcentratedPoolIDTickChange()
	cfmmPools := s.poolTracker.GetCFMMPools()
	cosmWasmPools := s.poolTracker.GetCosmWasmPools()

	changedBlockPools := domain.BlockPools{
		ConcentratedPools:            concentratedPools,
		ConcentratedPoolIDTickChange: concentratedPoolIDTickChange,
		CosmWasmPools:                cosmWasmPools,
		CFMMPools:                    cfmmPools,
	}

	return s.sqsIngester.ProcessChangedBlockData(ctx, changedBlockPools)
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
