package service

import (
	"context"
	"sync"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

var _ baseapp.StreamingService = (*indexerStreamingService)(nil)

// ind is a streaming service that processes block data and ingests it into the indexer
type indexerStreamingService struct {
	writeListeners map[storetypes.StoreKey][]storetypes.WriteListener

	// manages tracking of whether the node is code started
	coldStartManager domain.ColdStartManager

	client domain.Publisher

	keepers domain.Keepers
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, coldStartManager domain.ColdStartManager, client domain.Publisher, keepers domain.Keepers) baseapp.StreamingService {
	return &indexerStreamingService{

		writeListeners: writeListeners,

		coldStartManager: coldStartManager,

		client: client,

		keepers: keepers,
	}
}

// Close implements baseapp.StreamingService.
func (s *indexerStreamingService) Close() error {
	return nil
}

// ListenBeginBlock implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenBeginBlock(ctx context.Context, req types.RequestBeginBlock, res types.ResponseBeginBlock) error {
	return nil
}

// ListenCommit implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenCommit(ctx context.Context, res types.ResponseCommit) error {
	return nil
}

// ListenDeliverTx implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenDeliverTx(ctx context.Context, req types.RequestDeliverTx, res types.ResponseDeliverTx) error {
	return nil
}

// publishBlock publishes the block data to the indexer.
func (s *indexerStreamingService) publishBlock(ctx context.Context, req types.RequestEndBlock) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := (uint64)(req.GetHeight())
	timeEndBlock := sdkCtx.BlockTime().UTC()
	chainId := sdkCtx.ChainID()
	gasConsumed := sdkCtx.GasMeter().GasConsumed()
	block := domain.Block{
		ChainId:     chainId,
		Height:      height,
		BlockTime:   timeEndBlock,
		GasConsumed: gasConsumed,
	}
	return s.client.PublishBlock(sdkCtx, block)
}

// ListenEndBlock implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenEndBlock(ctx context.Context, req types.RequestEndBlock, res types.ResponseEndBlock) error {
	// Publish the block data
	var err error
	err = s.publishBlock(ctx, req)
	if err != nil {
		return err
	}
	// Publish the transaction data
	err = s.publishTxn(ctx, res)
	if err != nil {
		return err
	}

<<<<<<< HEAD
=======
// publishTxn publishes the transaction data to the indexer backend.
// TO DO: Tested if res.GetEvents() is the correct way to get the events in the SDK used in 'main'
func (s *indexerStreamingService) publishTxn(ctx context.Context, res abci.ResponseFinalizeBlock) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	events := res.GetEvents()
	if len(events) == 0 {
		return nil
	}
	txn := indexerdomain.Transaction{
		Height:    uint64(sdkCtx.BlockHeight()),
		BlockTime: sdkCtx.BlockTime().UTC(),
		Events:    make([]interface{}, len(events)),
	}
	for i, event := range events {
		txn.Events[i] = event
	}
	return s.client.PublishTransaction(sdkCtx, txn)
}

// ListenCommit updates the steaming service with the latest Commit messages and state changes
func (s *indexerStreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
>>>>>>> 8f766609 (feat: added txn ingester for indexer (#8473))
	// If did not ingest initial data yet, ingest it now
	if !s.coldStartManager.HasIngestedInitialData() {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		var err error

		// Ingest the initial data
		s.keepers.BankKeeper.IterateTotalSupply(sdkCtx, func(coin sdk.Coin) bool {
			// Check if the denom should be filtered out and skip it if so
			if domain.ShouldFilterDenom(coin.Denom) {
				return false
			}

			// Publish the token supply
			err = s.client.PublishTokenSupply(sdkCtx, domain.TokenSupply{
				Denom:  coin.Denom,
				Supply: coin.Amount,
			})

			// Skip any error silently but log it.
			if err != nil {
				// TODO: alert
				sdkCtx.Logger().Error("failed to publish token supply", "error", err)
			}

			supplyOffset := s.keepers.BankKeeper.GetSupplyOffset(sdkCtx, coin.Denom)

			// If supply offset is non-zero, publish it.
			if !supplyOffset.IsZero() {
				// Publish the token supply offset
				err = s.client.PublishTokenSupplyOffset(sdkCtx, domain.TokenSupplyOffset{
					Denom:        coin.Denom,
					SupplyOffset: supplyOffset,
				})
			}

			return false
		})

		// Mark that the initial data has been ingested
		s.coldStartManager.MarkInitialDataIngested()
	}

	return nil
}

// Listeners implements baseapp.StreamingService.
func (s *indexerStreamingService) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener {
	return s.writeListeners
}

// Stream implements baseapp.StreamingService.
func (s *indexerStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}
