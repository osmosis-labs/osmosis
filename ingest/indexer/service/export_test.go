package service

import (
	"context"
	"time"

	"github.com/cometbft/cometbft/abci/types"
)

type IndexerStreamingService = indexerStreamingService

func (s *indexerStreamingService) AddTokenLiquidity(ctx context.Context, event *types.Event) error {
	return s.addTokenLiquidity(ctx, event)
}

func (s *indexerStreamingService) AdjustTokenInAmountBySpreadFactor(ctx context.Context, event *types.Event) error {
	return s.adjustTokenInAmountBySpreadFactor(ctx, event)
}

func (s *indexerStreamingService) TrackCreatedPoolID(event types.Event, blockHeight int64, blockTime time.Time, txHash string) {
	s.trackCreatedPoolID(event, blockHeight, blockTime, txHash)
}
