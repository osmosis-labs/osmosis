package mocks

import (
	"context"

	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

// PublisherMock is a mock for Publisher.
type PublisherMock struct {
	CalledWithPair              indexerdomain.Pair
	CalledWithBlock             indexerdomain.Block
	CalledWithTokenSupply       indexerdomain.TokenSupply
	CalledWithTokenSupplyOffset indexerdomain.TokenSupplyOffset
	CalledWithTransaction       indexerdomain.Transaction
	ForcePairError              error
	ForceBlockError             error
	ForceTokenSupplyError       error
	ForceTokenSupplyOffsetError error
	ForceTransactionError       error
}

// PublishPair implements domain.Publisher.
func (p *PublisherMock) PublishPair(ctx context.Context, pair indexerdomain.Pair) error {
	return p.ForcePairError
}

// PublishBlock implements domain.Publisher.
func (p *PublisherMock) PublishBlock(ctx context.Context, block indexerdomain.Block) error {
	return p.ForceBlockError
}

// PublishTokenSupply implements domain.Publisher.
func (p *PublisherMock) PublishTokenSupply(ctx context.Context, tokenSupply indexerdomain.TokenSupply) error {
	p.CalledWithTokenSupply = tokenSupply
	return p.ForceTokenSupplyError
}

// PublishTokenSupplyOffset implements domain.Publisher.
func (p *PublisherMock) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset indexerdomain.TokenSupplyOffset) error {
	p.CalledWithTokenSupplyOffset = tokenSupplyOffset
	return p.ForceTokenSupplyOffsetError
}

// PublishTransaction implements domain.Publisher.
func (p *PublisherMock) PublishTransaction(ctx context.Context, txn indexerdomain.Transaction) error {
	return p.ForceTransactionError
}

var _ indexerdomain.Publisher = &PublisherMock{}
