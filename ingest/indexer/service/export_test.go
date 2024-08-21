package service

import (
	"context"

	"github.com/cometbft/cometbft/abci/types"
)

type IndexerStreamingService = indexerStreamingService

func (s *indexerStreamingService) AddTokenLiquidity(ctx context.Context, event *types.Event) error {
	return s.addTokenLiquidity(ctx, event)
}

func (s *indexerStreamingService) AdjustTokenInAmountBySpreadFactor(ctx context.Context, event *types.Event) error {
	return s.adjustTokenInAmountBySpreadFactor(ctx, event)
}
