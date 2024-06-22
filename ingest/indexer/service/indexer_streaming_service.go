package service

import (
	"context"
	"strings"
	"sync"

<<<<<<< HEAD
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
=======
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

var _ storetypes.ABCIListener = (*indexerStreamingService)(nil)

// ind is a streaming service that processes block data and ingests it into the indexer
type indexerStreamingService struct {
	writeListeners map[storetypes.StoreKey][]domain.WriteListener

	// manages tracking of whether the node is code started
	coldStartManager indexerdomain.ColdStartManager

	storeKeyMap map[string]storetypes.StoreKey

	client indexerdomain.Publisher

	keepers indexerdomain.Keepers
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
<<<<<<< HEAD
func New(writeListeners map[storetypes.StoreKey][]storetypes.WriteListener, coldStartManager domain.ColdStartManager, client domain.Publisher, keepers domain.Keepers) baseapp.StreamingService {
=======
func New(writeListeners map[storetypes.StoreKey][]domain.WriteListener, coldStartManager indexerdomain.ColdStartManager, client indexerdomain.Publisher, storeKeyMap map[string]storetypes.StoreKey, keepers indexerdomain.Keepers) storetypes.ABCIListener {
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	return &indexerStreamingService{

		writeListeners: writeListeners,

		coldStartManager: coldStartManager,

		client: client,

<<<<<<< HEAD
=======
		storeKeyMap: storeKeyMap,

>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
		keepers: keepers,
	}
}

// Close implements baseapp.StreamingService.
func (s *indexerStreamingService) Close() error {
	return nil
}

<<<<<<< HEAD
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
=======
// publishBlock publishes the block data to the indexer backend.
func (s *indexerStreamingService) publishBlock(ctx context.Context, req abci.RequestFinalizeBlock) error {
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := (uint64)(req.GetHeight())
	timeEndBlock := sdkCtx.BlockTime().UTC()
	chainId := sdkCtx.ChainID()
	gasConsumed := sdkCtx.GasMeter().GasConsumed()
<<<<<<< HEAD
	block := domain.Block{
=======
	block := indexerdomain.Block{
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
		ChainId:     chainId,
		Height:      height,
		BlockTime:   timeEndBlock,
		GasConsumed: gasConsumed,
	}
	return s.client.PublishBlock(sdkCtx, block)
}

<<<<<<< HEAD
// ListenEndBlock implements baseapp.StreamingService.
func (s *indexerStreamingService) ListenEndBlock(ctx context.Context, req types.RequestEndBlock, res types.ResponseEndBlock) error {
=======
// ListenFinalizeBlock updates the streaming service with the latest FinalizeBlock messages
func (s *indexerStreamingService) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	// Publish the block data
	err := s.publishBlock(ctx, req)
	if err != nil {
		return err
	}
<<<<<<< HEAD

=======
	return nil
}

// ListenCommit updates the steaming service with the latest Commit messages and state changes
func (s *indexerStreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	// If did not ingest initial data yet, ingest it now
	if !s.coldStartManager.HasIngestedInitialData() {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		var err error

		// Ingest the initial data
		s.keepers.BankKeeper.IterateTotalSupply(sdkCtx, func(coin sdk.Coin) bool {
			// Skip CL pool shares
			if strings.Contains(coin.Denom, "cl/pool") {
				return false
			}

			// Publish the token supply
<<<<<<< HEAD
			err = s.client.PublishTokenSupply(sdkCtx, domain.TokenSupply{
=======
			err = s.client.PublishTokenSupply(sdkCtx, indexerdomain.TokenSupply{
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
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
<<<<<<< HEAD
				err = s.client.PublishTokenSupplyOffset(sdkCtx, domain.TokenSupplyOffset{
=======
				err = s.client.PublishTokenSupplyOffset(sdkCtx, indexerdomain.TokenSupplyOffset{
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
					Denom:        coin.Denom,
					SupplyOffset: supplyOffset,
				})
			}

			return false
		})

		// Mark that the initial data has been ingested
		s.coldStartManager.MarkInitialDataIngested()
<<<<<<< HEAD
=======
	} else {
		for _, kv := range changeSet {
			for _, listener := range s.writeListeners[s.storeKeyMap[kv.StoreKey]] {
				if err := listener.OnWrite(s.storeKeyMap[kv.StoreKey], kv.Key, kv.Value, kv.Delete); err != nil {
					return err
				}
			}
		}
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	}

	return nil
}

// Listeners implements baseapp.StreamingService.
<<<<<<< HEAD
func (s *indexerStreamingService) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener {
=======
func (s *indexerStreamingService) Listeners() map[storetypes.StoreKey][]domain.WriteListener {
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
	return s.writeListeners
}

// Stream implements baseapp.StreamingService.
func (s *indexerStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}
