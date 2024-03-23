package service

import (
	"context"
	"sync"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/app/keepers"
	"github.com/osmosis-labs/osmosis/v23/ingest"
)

var _ baseapp.StreamingService = (*sqsStreamingService)(nil)

type sqsStreamingService struct {
	keepers        *keepers.AppKeepers
	writeListeners map[storetypes.StoreKey][]storetypes.WriteListener
	sqsIngester    ingest.Ingester
	poolTracker    PoolTracker

	isColdStart bool
}

func New(keepers *keepers.AppKeepers, writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, sqsIngester ingest.Ingester, poolTracker PoolTracker) baseapp.StreamingService {
	return &sqsStreamingService{
		keepers:        keepers,
		writeListeners: writeListeners,
		sqsIngester:    sqsIngester,
		poolTracker:    poolTracker,

		isColdStart: true,
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

// ListenEndBlock implements baseapp.StreamingService.
func (s *sqsStreamingService) ListenEndBlock(ctx context.Context, req types.RequestEndBlock, res types.ResponseEndBlock) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			// Panics are silently logged and ignored.
			sdkCtx.Logger().Error("panic while processing block during ingest", "err", r)
		}
	}()

	// If cold start, we use SQS ingestert to process the intire block.
	if s.isColdStart {

		if err := s.sqsIngester.ProcessBlock(sdkCtx); err != nil {
			return err
		}

		// Succesfully processed the block, no longer cold start.
		s.isColdStart = false

		return nil
	}

	// If not cold start, we only process the pools that were changed this block.

	ceontratedPools := s.poolTracker.GetConcentratedPools()

	// Reset pool tracking for this block.
	defer s.poolTracker.Reset()

	return s.sqsIngester.ProcessChangedBlockData(sdkCtx, ceontratedPools)
}

// Listeners implements baseapp.StreamingService.
func (s *sqsStreamingService) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener {
	return s.writeListeners
}

// Stream implements baseapp.StreamingService.
func (s *sqsStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}
