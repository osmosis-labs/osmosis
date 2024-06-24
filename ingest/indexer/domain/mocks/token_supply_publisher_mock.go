package mocks

import (
	"context"

	indexerdomain "github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

// TokenSupplyPublisherMock is a mock for TokenSupplyPublisher.
type TokenSupplyPublisherMock struct {
	CalledWithTokenSupply indexerdomain.TokenSupply
	ForceTokenSupplyError error

	CalledWithTokenSupplyOffset indexerdomain.TokenSupplyOffset
	ForceTokenSupplyOffsetError error
}

// PublishTokenSupply implements domain.PubSubClientI.
func (p *TokenSupplyPublisherMock) PublishTokenSupply(ctx context.Context, tokenSupply indexerdomain.TokenSupply) error {
	p.CalledWithTokenSupply = tokenSupply
	return p.ForceTokenSupplyError
}

// PublishTokenSupplyOffset implements domain.PubSubClientI.
func (p *TokenSupplyPublisherMock) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset indexerdomain.TokenSupplyOffset) error {
	p.CalledWithTokenSupplyOffset = tokenSupplyOffset
	return p.ForceTokenSupplyOffsetError
}

var (
	_ indexerdomain.TokenSupplyPublisher = (*TokenSupplyPublisherMock)(nil)
)
