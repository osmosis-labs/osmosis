package service

import (
	"context"
	"encoding/hex"
	"strings"
	"sync"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/blockprocessor"

	sqsdomain "github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
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

	poolTracker sqsdomain.BlockPoolUpdateTracker

	txDecoder sdk.TxDecoder

	logger log.Logger
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(blockUpdatesProcessUtils commondomain.BlockUpdateProcessUtilsI, blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, storeKeyMap map[string]storetypes.StoreKey, poolExtractor commondomain.PoolExtractor, poolTracker sqsdomain.BlockPoolUpdateTracker, keepers domain.Keepers, txDecoder sdk.TxDecoder, logger log.Logger) storetypes.ABCIListener {
	return &indexerStreamingService{
		blockProcessStrategyManager: blockProcessStrategyManager,

		poolExtractor: poolExtractor,

		poolTracker: poolTracker,

		client: client,

		keepers: keepers,

		blockUpdatesProcessUtils: blockUpdatesProcessUtils,

		txDecoder: txDecoder,

		logger: logger,
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

		// Obtain the token in before spread
		// Related thread: https://osmosis-network.slack.com/archives/C060VCJAVBL/p1721316448977489
		var tokenInBeforeSpread string
		for _, txMsgGeneric := range txMessages {
			if txMsg, ok := txMsgGeneric.(*poolmanagertypes.MsgSwapExactAmountIn); ok {
				tokenInBeforeSpread = txMsg.GetTokenIn().String()
				break
			}
			if txMsg, ok := txMsgGeneric.(*poolmanagertypes.MsgSplitRouteSwapExactAmountIn); ok {
				tokenInDenom := txMsg.GetTokenInDenom()
				routes := txMsg.GetRoutes()
				for _, route := range routes {
					tokenInAmount := route.TokenInAmount
					tokenInBeforeSpread = tokenInAmount.String() + tokenInDenom
					break
				}
			}
			if txMsg, ok := txMsgGeneric.(*gammtypes.MsgSwapExactAmountIn); ok {
				tokenInBeforeSpread = txMsg.GetTokenIn().String()
				break
			}
		}

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
			Height:              uint64(sdkCtx.BlockHeight()),
			BlockTime:           sdkCtx.BlockTime().UTC(),
			GasWanted:           uint64(gasWanted),
			GasUsed:             uint64(gasUsed),
			Fees:                fee,
			MessageType:         msgType,
			TransactionHash:     txHash,
			TransactionIndexId:  txnIndex,
			TokenInBeforeSpread: tokenInBeforeSpread,
			Events:              includedEvents,
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
	// Log the status only for the first block
	// Avoid subsequent blocks to avoid spamming the logs
	if s.blockProcessStrategyManager.ShouldPushAllData() {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting indexer ingest ListenFinalizeBlock", "height", sdkCtx.BlockHeight())

		defer func() {
			sdkCtx.Logger().Info("Finished indexer ingest ListenFinalizeBlock", "height", sdkCtx.BlockHeight())
		}()
	}

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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Log the status only for the first block
	// Avoid subsequent blocks to avoid spamming the logs
	if s.blockProcessStrategyManager.ShouldPushAllData() {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.Logger().Info("Starting indexer ingest ListenCommit", "height", sdkCtx.BlockHeight())

		defer func() {
			sdkCtx.Logger().Info("Finished indexer ingest ListenCommit", "height", sdkCtx.BlockHeight())
		}()
	}

	defer func() {
		// Reset the pool tracker after processing the block.
		s.poolTracker.Reset()
	}()

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
