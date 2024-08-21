package service

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/blockprocessor"
	sqsdomain "github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"

	"github.com/osmosis-labs/osmosis/osmomath"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v25/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

var (
	_      baseapp.StreamingService = (*indexerStreamingService)(nil)
	oneDec                          = osmomath.OneDec()
)

// ind is a streaming service that processes block data and ingests it into the indexer
type indexerStreamingService struct {
	writeListeners map[storetypes.StoreKey][]storetypes.WriteListener

	// manages tracking of whether the node is code started
	// manages tracking of whether all the data should be processed or only the changed in the block
	blockProcessStrategyManager commondomain.BlockProcessStrategyManager

	client domain.Publisher

	keepers domain.Keepers

	txDecoder sdk.TxDecoder

	txnIndexId int
	// extracts the pools from chain state
	poolExtractor commondomain.PoolExtractor

	poolTracker sqsdomain.BlockPoolUpdateTracker

	logger log.Logger
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, blockProcessStrategyManager commondomain.BlockProcessStrategyManager, client domain.Publisher, poolExtractor commondomain.PoolExtractor, poolTracker sqsdomain.BlockPoolUpdateTracker, keepers domain.Keepers, txDecoder sdk.TxDecoder, logger log.Logger) *indexerStreamingService {
	return &indexerStreamingService{
		blockProcessStrategyManager: blockProcessStrategyManager,

		writeListeners: writeListeners,

		poolExtractor: poolExtractor,

		poolTracker: poolTracker,

		client: client,

		keepers: keepers,

		txDecoder: txDecoder,

		logger: logger,
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
	// Increment the transaction index after delivering the transaction
	defer func() {
		s.txnIndexId++
	}()

	// Publish the transaction data
	err := s.publishTxn(ctx, req, res)
	if err != nil {
		s.logger.Error("Error publishing transaction data", "error", err)
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

// adjustAmountBySpreadFactor adjusts the amount by the spread factor.
// This is done to adjust the amount of tokens in the token_swapped event by the spread factor,
// as the amount in the event is the amount AFTER the spread factor is applied.
// therefore, we need to adjust the amount by the spread factor to get the amount BEFORE the spread factor is applied.
// NOTE: This applies to CL pools only
func (s *indexerStreamingService) adjustTokenInAmountBySpreadFactor(ctx context.Context, event *types.Event) error {
	if event.Type != gammtypes.TypeEvtTokenSwapped {
		return nil
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var poolIdStr string
	var afterTokensIn string
	var afterTokensInIndex int
	attributes := event.Attributes
	// Find the pool id and tokens in attributes from the token_swapped
	for index, attribute := range attributes {
		if poolIdStr != "" && afterTokensIn != "" {
			break
		}
		if attribute.Key == concentratedliquiditytypes.AttributeKeyPoolId {
			poolIdStr = attribute.Value
		}
		if attribute.Key == concentratedliquiditytypes.AttributeKeyTokensIn {
			afterTokensIn = attribute.Value
			afterTokensInIndex = index
		}
	}
	if poolIdStr == "" || afterTokensIn == "" {
		return errors.New("pool id or tokens in not found")
	}
	poolId, err := strconv.ParseInt(poolIdStr, 10, 64)
	if err != nil {
		return errors.New("failed to parse pool id")
	}
	pool, err := s.keepers.PoolManagerKeeper.GetPool(sdkCtx, uint64(poolId))
	if err != nil {
		return errors.New("failed to get pool")
	}
	// Adjustment required only for CL pools
	if pool.GetType() != poolmanagertypes.Concentrated {
		return nil
	}
	spreadFactor := pool.GetSpreadFactor(sdkCtx)
	coins, err := sdk.ParseCoinsNormalized(afterTokensIn)
	if err != nil {
		return errors.New("failed to parse tokens in")
	}
	tokenInAmount := coins[0].Amount.ToLegacyDec()
	// Adjust the amount by the spread factor, i.e. before = after/(1 - spreadFactor)
	adjustedAmt := tokenInAmount.Quo(oneDec.Sub(spreadFactor))
	attributes[afterTokensInIndex].Value = adjustedAmt.String() + coins[0].Denom
	return nil
}

// publishTxn publishes the transaction data to the indexer.
func (s *indexerStreamingService) publishTxn(ctx context.Context, req types.RequestDeliverTx, res types.ResponseDeliverTx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Decode the transaction
	tx, err := s.txDecoder(req.GetTx())
	if err != nil {
		return err
	}

	// Calculate the transaction hash
	txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(req.GetTx())))

	// Gas data
	gasWanted := res.GasWanted
	gasUsed := res.GasUsed

	// Fee data
	feeTx, _ := tx.(sdk.FeeTx)
	fee := feeTx.GetFee()

	// Message type
	txMessages := tx.GetMsgs()
	msgType := proto.MessageName(txMessages[0])

	// Include these events only:
	// - token_swapped
	// - pool_joined
	// - pool_exited
	// - create_position
	// - withdraw_position
	events := res.GetEvents()
	var includedEvents []domain.EventWrapper
	for i, event := range events {
		clonedEvent := deepCloneEvent(&event)
		err := s.adjustTokenInAmountBySpreadFactor(ctx, clonedEvent)
		if err != nil {
			s.logger.Error("Error adjusting amount by spread factor", "error", err)
			continue
		}
		err = s.addTokenLiquidity(ctx, clonedEvent)
		if err != nil {
			s.logger.Error("Error adding reserves to event", "error", err)
			continue
		}
		eventType := clonedEvent.Type
		if eventType == gammtypes.TypeEvtTokenSwapped || eventType == gammtypes.TypeEvtPoolJoined || eventType == gammtypes.TypeEvtPoolExited || eventType == concentratedliquiditytypes.TypeEvtCreatePosition || eventType == concentratedliquiditytypes.TypeEvtWithdrawPosition {
			includedEvents = append(includedEvents, domain.EventWrapper{Index: i, Event: *clonedEvent})
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
		TransactionIndexId: s.txnIndexId,
		Events:             includedEvents,
	}
	return s.client.PublishTransaction(sdkCtx, txn)
}

// addTokenLiquidity adds the token liquidity to the event.
// It refers to the pooled amount of each asset after a swap event has occurred.
func (s *indexerStreamingService) addTokenLiquidity(ctx context.Context, event *types.Event) error {
	if event.Type != gammtypes.TypeEvtTokenSwapped {
		return nil
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var poolIdStr string
	attributes := event.Attributes
	// Find the pool id from the token_swapped
	for _, attribute := range attributes {
		if attribute.Key == concentratedliquiditytypes.AttributeKeyPoolId {
			poolIdStr = attribute.Value
			break
		}
	}
	if poolIdStr == "" {
		return errors.New("pool id not found")
	}
	poolId, err := strconv.ParseUint(poolIdStr, 10, 64)
	if err != nil {
		return err
	}
	coins, err := s.keepers.PoolManagerKeeper.GetTotalPoolLiquidity(sdkCtx, poolId)
	if err != nil {
		return err
	}
	// Store the liquidity of the token in the attributes map of the event, keyed by "liquidity_" + coin.Denom
	for _, coin := range coins {
		event.Attributes = append(event.Attributes, types.EventAttribute{
			Key:   "liquidity_" + coin.Denom,
			Value: coin.Amount.String(),
		})
	}
	return nil
}

// ListenEndBlock implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenEndBlock(ctx context.Context, req types.RequestEndBlock, res types.ResponseEndBlock) error {
	// Reset the transaction index id on end block
	defer func() {
		s.txnIndexId = 0
	}()
	// Publish the block data
	err := s.publishBlock(ctx, req)
	if err != nil {
		s.logger.Error("Error publishing block data", "error", err)
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

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

// Listeners implements baseapp.StreamingService.
func (s *indexerStreamingService) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener {
	return s.writeListeners
}

// Stream implements baseapp.StreamingService.
func (s *indexerStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}

// deepCloneEvent deep clones the event.
func deepCloneEvent(event *types.Event) *types.Event {
	clone := *event
	clone.Attributes = make([]types.EventAttribute, len(event.Attributes))
	copy(clone.Attributes, event.Attributes)
	return &clone
}
