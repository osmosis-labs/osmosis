package indexer

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	service "github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/client"
)

// indexerIngester is an implementation of domain.Publisher.
type indexerPublisher struct {
	pubsubClient service.PubSubClient
}

// NewIndexerPublisher creates a new IndexerPublisher with the given PubSubClient
func NewIndexerPublisher(pubsubClient service.PubSubClient) domain.Publisher {
	return &indexerPublisher{
		pubsubClient: pubsubClient,
	}
}

// PublishBlock implements domain.Publisher.
func (i *indexerPublisher) PublishBlock(ctx context.Context, block domain.Block) error {
	err := i.pubsubClient.PublishBlock(ctx, block)
	if err != nil {
		return err
	}
	return nil
}

// PublishTransaction implements domain.Publisher.
func (i *indexerPublisher) PublishTransaction(ctx context.Context, txn domain.Transaction) error {
	err := i.pubsubClient.PublishTransaction(ctx, txn)
	if err != nil {
		return err
	}
	return nil
}

// PublishPool implements domain.Publisher.
func (i *indexerPublisher) PublishPool(ctx context.Context, pool domain.Pool) error {
	err := i.pubsubClient.PublishPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

// PublishTokenSupply implements domain.Publisher.
func (i *indexerPublisher) PublishTokenSupply(ctx context.Context, tokenSupply domain.TokenSupply) error {
	err := i.pubsubClient.PublishTokenSupply(ctx, tokenSupply)
	if err != nil {
		return err
	}
	return nil
}

// PublishTokenSupplyOffset implements domain.Publisher.
func (i *indexerPublisher) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset domain.TokenSupplyOffset) error {
	err := i.pubsubClient.PublishTokenSupplyOffset(ctx, tokenSupplyOffset)
	if err != nil {
		return err
	}
	return nil
}
