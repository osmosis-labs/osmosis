package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

type PublisherMock struct {
}

// PublishPair implements domain.Publisher.
func (p *PublisherMock) PublishPair(ctx context.Context, pair domain.Pair) error {
	panic("unimplemented")
}

// PublishBlock implements domain.Publisher.
func (p *PublisherMock) PublishBlock(ctx context.Context, block domain.Block) error {
	panic("unimplemented")
}

// PublishPool implements domain.Publisher.
func (p *PublisherMock) PublishPool(ctx context.Context, pool domain.Pool) error {
	panic("unimplemented")
}

// PublishPools implements domain.Publisher.
func (p *PublisherMock) PublishPools(ctx context.Context, pools []types.PoolI) error {
	panic("unimplemented")
}

// PublishTokenSupply implements domain.Publisher.
func (p *PublisherMock) PublishTokenSupply(ctx context.Context, tokenSupply domain.TokenSupply) error {
	panic("unimplemented")
}

// PublishTokenSupplyOffset implements domain.Publisher.
func (p *PublisherMock) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset domain.TokenSupplyOffset) error {
	panic("unimplemented")
}

// PublishTransaction implements domain.Publisher.
func (p *PublisherMock) PublishTransaction(ctx context.Context, txn domain.Transaction) error {
	panic("unimplemented")
}

var _ domain.Publisher = &PublisherMock{}
