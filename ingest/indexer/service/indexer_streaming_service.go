package service

import (
	"context"
	"strings"
	"sync"

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

	client indexerdomain.Ingester

	keepers indexerdomain.Keepers
}

// New creates a new sqsStreamingService.
// writeListeners is a map of store keys to write listeners.
// sqsIngester is an ingester that ingests the block data into SQS.
// poolTracker is a tracker that tracks the pools that were changed in the block.
// nodeStatusChecker is a checker that checks if the node is syncing.
func New(writeListeners map[storetypes.StoreKey][]domain.WriteListener, coldStartManager indexerdomain.ColdStartManager, client indexerdomain.Ingester, keepers indexerdomain.Keepers) storetypes.ABCIListener {
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

// ListenFinalizeBlock updates the streaming service with the latest FinalizeBlock messages
func (s *indexerStreamingService) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	return nil
}

// ListenCommit updates the steaming service with the latest Commit messages and state changes
func (s *indexerStreamingService) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	// If did not ingest initial data yet, ingest it now
	if !s.coldStartManager.HasIngestedInitialData() {

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		var err error

		// Ingest the initial data
		s.keepers.BankKeeper.IterateTotalSupplyWithOffsets(sdkCtx, func(coin sdk.Coin) bool {

			// Skip CL pool shares
			if strings.Contains(coin.Denom, "cl/pool") {
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

			return false
		})

		// Mark that the initial data has been ingested
		s.coldStartManager.MarkInitialDataIngested()
	}

	return nil
}

// Listeners implements baseapp.StreamingService.
func (s *indexerStreamingService) Listeners() map[storetypes.StoreKey][]domain.WriteListener {
	return s.writeListeners
}

// Stream implements baseapp.StreamingService.
func (s *indexerStreamingService) Stream(wg *sync.WaitGroup) error {
	return nil
}
