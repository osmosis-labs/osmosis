package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

type PublisherMock struct {
	Error error
}

var _ domain.Publisher = &PublisherMock{}

func (g *PublisherMock) PublishBlock(ctx context.Context, block domain.Block) error {
	return g.Error
}

func (g *PublisherMock) PublishTransaction(ctx context.Context, txn domain.Transaction) error {
	return g.Error
}

func (g *PublisherMock) PublishPair(ctx context.Context, pair domain.Pair) error {
	return g.Error
}

func (g *PublisherMock) PublishTokenSupply(ctx context.Context, tokenSupply domain.TokenSupply) error {
	return g.Error
}

func (g *PublisherMock) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset domain.TokenSupplyOffset) error {
	return g.Error
}
