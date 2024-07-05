package service

import (
	"context"
	"sync"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
)

var _ storetypes.ABCIListener = (*indexerStreamingService)(nil)

// ind is a streaming service that processes block data and ingests it into the indexer
type indexerStreamingService struct {
	// manages tracking of whether the node is code started
	coldStartManager indexerdomain.ColdStartManager

	client indexerdomain.Publisher

	keepers indexerdomain.Keepers

	blockUpdatesProcessUtils commondomain.BlockUpdateProcessUtilsI
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(blockUpdatesProcessUtils commondomain.BlockUpdateProcessUtilsI, coldStartManager indexerdomain.ColdStartManager, client indexerdomain.Publisher, storeKeyMap map[string]storetypes.StoreKey, keepers indexerdomain.Keepers) storetypes.ABCIListener {
	return &indexerStreamingService{

		coldStartManager: coldStartManager,

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
	block := indexerdomain.Block{
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// If did not ingest initial data yet, ingest it now
	if !s.coldStartManager.HasIngestedInitialData() {
		var err error

		// Ingest the initial data
		s.keepers.BankKeeper.IterateTotalSupply(sdkCtx, func(coin sdk.Coin) bool {
			// Check if the denom should be filtered out and skip it if so
			if indexerdomain.ShouldFilterDenom(coin.Denom) {
				return false
			}

			// Publish the token supply
			err = s.client.PublishTokenSupply(sdkCtx, indexerdomain.TokenSupply{
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
				err = s.client.PublishTokenSupplyOffset(sdkCtx, indexerdomain.TokenSupplyOffset{
					Denom:        coin.Denom,
					SupplyOffset: supplyOffset,
				})
			}

			return false
		})

		// Mark that the initial data has been ingested
		s.coldStartManager.MarkInitialDataIngested()
	} else {
		s.blockUpdatesProcessUtils.SetChangeSet(changeSet)
		// Avoid
		if err := s.blockUpdatesProcessUtils.ProcessBlockChangeSet(); err != nil {
			sdkCtx.Logger().Error("failed to process block change set in indexer ListenCommit", "error", err)

			// Return error to stop processing blocks expecting manual intervention.
			return err
		}
	}

	return nil
}

// Stream implements baseapp.StreamingService.
func (s *indexerStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}
