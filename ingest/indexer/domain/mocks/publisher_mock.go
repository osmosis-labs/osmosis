package mocks

import (
	"context"
	"sync"

	indexerdomain "github.com/osmosis-labs/osmosis/v31/ingest/indexer/domain"
)

// PublisherMock is a mock for Publisher.
type PublisherMock struct {
	CalledWithPair                   indexerdomain.Pair
	CalledWithBlock                  indexerdomain.Block
	CalledWithTokenSupply            indexerdomain.TokenSupply
	CalledWithTokenSupplyOffset      indexerdomain.TokenSupplyOffset
	CalledWithTransaction            indexerdomain.Transaction
	NumPublishPairCalls              int
	NumPublishPairCallsWithCreation  int
	PublishPairCallMutex             sync.Mutex
	NumPublishBlockCalls             int
	NumPublishTokenSupplyCalls       int
	NumPublishTokenSupplyOffsetCalls int
	NumPublishTransactionCalls       int
	ForcePairError                   error
	ForceBlockError                  error
	ForceTokenSupplyError            error
	ForceTokenSupplyOffsetError      error
	ForceTransactionError            error
}

// PublishPair implements domain.Publisher.
func (p *PublisherMock) PublishPair(ctx context.Context, pair indexerdomain.Pair) error {
	// Make it thread safe to achieve atomicity in concurrent calls.
	// as PublishPair is called concurrently by multiple goroutines.
	// See also: pair_publisher.go, PublishPoolPairs function.
	p.PublishPairCallMutex.Lock()
	defer p.PublishPairCallMutex.Unlock()
	p.CalledWithPair = pair
	p.NumPublishPairCalls++
	if !pair.PairCreatedAt.IsZero() {
		p.NumPublishPairCallsWithCreation++
	}
	return p.ForcePairError
}

// PublishBlock implements domain.Publisher.
func (p *PublisherMock) PublishBlock(ctx context.Context, block indexerdomain.Block) error {
	p.CalledWithBlock = block
	p.NumPublishBlockCalls++
	return p.ForceBlockError
}

// PublishTokenSupply implements domain.Publisher.
func (p *PublisherMock) PublishTokenSupply(ctx context.Context, tokenSupply indexerdomain.TokenSupply) error {
	p.CalledWithTokenSupply = tokenSupply
	p.NumPublishTokenSupplyCalls++
	return p.ForceTokenSupplyError
}

// PublishTokenSupplyOffset implements domain.Publisher.
func (p *PublisherMock) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset indexerdomain.TokenSupplyOffset) error {
	p.CalledWithTokenSupplyOffset = tokenSupplyOffset
	p.NumPublishTokenSupplyOffsetCalls++
	return p.ForceTokenSupplyOffsetError
}

// PublishTransaction implements domain.Publisher.
func (p *PublisherMock) PublishTransaction(ctx context.Context, txn indexerdomain.Transaction) error {
	p.CalledWithTransaction = txn
	p.NumPublishTransactionCalls++
	return p.ForceTransactionError
}

var _ indexerdomain.Publisher = &PublisherMock{}
