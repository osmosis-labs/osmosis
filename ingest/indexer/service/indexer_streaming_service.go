package service

import (
	"context"
	"encoding/hex"
	"strings"
	"sync"

<<<<<<< HEAD
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
=======
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
>>>>>>> c994c001 (Revised txn ingester for iterating txns in main branch (#8495))
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/blockprocessor"
)

var _ baseapp.StreamingService = (*indexerStreamingService)(nil)

// ind is a streaming service that processes block data and ingests it into the indexer
type indexerStreamingService struct {
	writeListeners map[storetypes.StoreKey][]storetypes.WriteListener

	// manages tracking of whether the node is code started
	// manages tracking of whether all the data should be processed or only the changed in the block
	blockProcessStrategyManager commondomain.BlockProcessStrategyManager

	client domain.Publisher

	keepers domain.Keepers

	// extracts the pools from chain state
	poolExtractor commondomain.PoolExtractor

	txDecoder sdk.TxDecoder

	logger log.Logger
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
<<<<<<< HEAD
func New(writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, poolExtractor commondomain.PoolExtractor, keepers domain.Keepers) baseapp.StreamingService {
=======
func New(blockUpdatesProcessUtils commondomain.BlockUpdateProcessUtilsI, blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, storeKeyMap map[string]storetypes.StoreKey, poolExtractor commondomain.PoolExtractor, keepers domain.Keepers, txDecoder sdk.TxDecoder, logger log.Logger) storetypes.ABCIListener {
>>>>>>> c994c001 (Revised txn ingester for iterating txns in main branch (#8495))
	return &indexerStreamingService{
		blockProcessStrategyManager: blockProcessStrategyManager,

		writeListeners: writeListeners,

		poolExtractor: poolExtractor,

		client: client,

		keepers: keepers,
<<<<<<< HEAD
=======

		blockUpdatesProcessUtils: blockUpdatesProcessUtils,

		txDecoder: txDecoder,

		logger: logger,
>>>>>>> c994c001 (Revised txn ingester for iterating txns in main branch (#8495))
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
	// Publish the transaction data
	err := s.publishTxn(ctx, res)
	if err != nil {
		return err
	}
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

<<<<<<< HEAD
// publishTxn publishes the transaction data to the indexer.
func (s *indexerStreamingService) publishTxn(ctx context.Context, res types.ResponseDeliverTx) error {
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

// ListenEndBlock implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenEndBlock(ctx context.Context, req types.RequestEndBlock, res types.ResponseEndBlock) error {
	// Publish the block data
	err := s.publishBlock(ctx, req)
	if err != nil {
		return err
	}

=======
// publishTxn iterates through the transactions in the block and publishes them to the indexer backend.
func (s *indexerStreamingService) publishTxn(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	txns := req.GetTxs()
	for txnIndex, txByteArr := range txns {
		// Decode the transaction
		tx, err := s.txDecoder(txByteArr)
		if err != nil {
			return err
		}
		// Calculate the transaction hash
		txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(txByteArr)))

		// Gas data
		gasWanted := res.TxResults[txnIndex].GasWanted
		gasUsed := res.TxResults[txnIndex].GasUsed

		// Fee data
		feeTx, _ := tx.(sdk.FeeTx)
		fee := feeTx.GetFee()

		// Message type
		// TO BE VERIFIED - This may not be the correct way to obtain message type
		txMessages := tx.GetMsgs()
		msgType := txMessages[0].String()

		// Include these events only:
		// - token_swapped
		// - pool_joined
		// - pool_exited
		// - create_position
		// - withdraw_position
		events := res.GetEvents()
		var includedEvents []domain.EventWrapper
		for i, event := range events {
			eventType := event.Type
			if eventType == "token_swapped" || eventType == "pool_joined" || eventType == "pool_exited" || eventType == "create_position" || eventType == "withdraw_position" {
				includedEvents = append(includedEvents, domain.EventWrapper{Index: i, Event: event})
			}
		}

		// Publish the transaction
		txn := domain.Transaction{
			Height:             uint64(sdkCtx.BlockHeight()),
			BlockTime:          sdkCtx.BlockTime().UTC(),
			GasWanted:          uint64(gasWanted),
			GasUsed:            uint64(gasUsed),
			Fees:               fee,
			MessageType:        msgType,
			TransactionHash:    txHash,
			TransactionIndexId: txnIndex,
			Events:             includedEvents,
		}
		err = s.client.PublishTransaction(sdkCtx, txn)
		if err != nil {
			// if there is an error in publishing the transaction, return the error
			return err
		}
	}
	return nil
}

// ListenFinalizeBlock updates the streaming service with the latest FinalizeBlock messages
func (s *indexerStreamingService) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	// Publish the block data
	var err error
	err = s.publishBlock(ctx, req)
	if err != nil {
		s.logger.Error("Error publishing block data by indexer", err)
		return err
	}
	// Iterate through the transactions in the block and publish them
	err = s.publishTxn(ctx, req, res)
	if err != nil {
		s.logger.Error("Error publishing transaction data by indexer", err)
		return err
	}
	return nil
}

// ListenCommit updates the steaming service with the latest Commit messages and state changes
func (s *indexerStreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
>>>>>>> c994c001 (Revised txn ingester for iterating txns in main branch (#8495))
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create block processor
	blockProcessor := blockprocessor.NewBlockProcessor(s.blockProcessStrategyManager, s.client, s.poolExtractor, s.keepers)

	// Process block.
	if err := blockProcessor.ProcessBlock(sdkCtx); err != nil {
		return err
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
