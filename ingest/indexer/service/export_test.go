package service

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
)

type IndexerStreamingService = indexerStreamingService

func (s *indexerStreamingService) AddTokenLiquidity(ctx context.Context, event *abci.Event) error {
	return s.addTokenLiquidity(ctx, event)
}

func (s *indexerStreamingService) AdjustTokenInAmountBySpreadFactor(ctx context.Context, event *abci.Event) error {
	return s.adjustTokenInAmountBySpreadFactor(ctx, event)
}

func (s *indexerStreamingService) TrackCreatedPoolID(event abci.Event, blockHeight int64, blockTime time.Time, txHash string) error {
	return s.trackCreatedPoolID(event, blockHeight, blockTime, txHash)
}
