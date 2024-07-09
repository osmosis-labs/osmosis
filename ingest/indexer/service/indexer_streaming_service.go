package service

import (
	"context"
	"sync"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/blockprocessor"
)

var _ storetypes.ABCIListener = (*indexerStreamingService)(nil)

// ind is a streaming service that processes block data and ingests it into the indexer
type indexerStreamingService struct {
	// manages tracking of whether all the data should be processed or only the changed in the block
	blockProcessStrategyManager commondomain.BlockProcessStrategyManager

	client domain.Publisher

	keepers domain.Keepers

	blockUpdatesProcessUtils commondomain.BlockUpdateProcessUtilsI

	// extracts the pools from chain state
	poolExtractor commondomain.PoolExtractor
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(blockUpdatesProcessUtils commondomain.BlockUpdateProcessUtilsI, blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, storeKeyMap map[string]storetypes.StoreKey, poolExtractor commondomain.PoolExtractor, keepers domain.Keepers) storetypes.ABCIListener {
	return &indexerStreamingService{
		blockProcessStrategyManager: blockProcessStrategyManager,

		poolExtractor: poolExtractor,

		client: client,

		keepers: keepers,

		blockUpdatesProcessUtils: blockUpdatesProcessUtils,
	}
}

// Close implements baseapp.StreamingService.
func (s *indexerStreamingService) Close() error {
	return nil
}

// publishBlock publishes the block data to the indexer backend.
func (s *indexerStreamingService) publishBlock(ctx context.Context, req abci.RequestFinalizeBlock) error {
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

// ListenFinalizeBlock updates the streaming service with the latest FinalizeBlock messages
func (s *indexerStreamingService) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
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
	return nil
}

// publishTxn publishes the transaction data to the indexer backend.
// TO DO: Tested if res.GetEvents() is the correct way to get the events in the SDK used in 'main'
func (s *indexerStreamingService) publishTxn(ctx context.Context, res abci.ResponseFinalizeBlock) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	events := res.GetEvents()
	if len(events) == 0 {
		return nil
	}
	txn := domain.Transaction{
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create block processor
	blockProcessor := blockprocessor.NewBlockProcessor(s.blockProcessStrategyManager, s.client, s.poolExtractor, s.keepers)

	// Process block.
	if err := blockProcessor.ProcessBlock(sdkCtx); err != nil {
		return err
	}

	return nil
}

// Stream implements baseapp.StreamingService.
func (s *indexerStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}
