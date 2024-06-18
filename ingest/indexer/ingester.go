package indexer

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	service "github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/client"
)

// indexerIngester is an implementation of domain.Ingester.
type indexerIngester struct {
	pubsubClient service.PubSubClient
}

// NewIndexerIngester creates a new indexer ingester.
func NewIndexerIngester(pubsubClient service.PubSubClient) domain.Ingester {
	return &indexerIngester{
		pubsubClient: pubsubClient,
	}
}

// PublishBlock implements domain.Ingester.
func (i *indexerIngester) PublishBlock(ctx context.Context, block domain.Block) error {
	err := i.pubsubClient.PublishBlock(ctx, block)
	if err != nil {
		return err
	}
	return nil
}

// PublishTransaction implements domain.Ingester.
func (i *indexerIngester) PublishTransaction(ctx context.Context, txn domain.Transaction) error {
	err := i.pubsubClient.PublishTransaction(ctx, txn)
	if err != nil {
		return err
	}
	return nil
}

// PublishPool implements domain.Ingester.
func (i *indexerIngester) PublishPool(ctx context.Context, pool domain.Pool) error {
	err := i.pubsubClient.PublishPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

// PublishTokenSupply implements domain.Ingester.
func (i *indexerIngester) PublishTokenSupply(ctx context.Context, tokenSupply domain.TokenSupply) error {
	err := i.pubsubClient.PublishTokenSupply(ctx, tokenSupply)
	if err != nil {
		return err
	}
	return nil
}

// PublishTokenSupplyOffset implements domain.Ingester.
func (i *indexerIngester) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset domain.TokenSupplyOffset) error {
	err := i.pubsubClient.PublishTokenSupplyOffset(ctx, tokenSupplyOffset)
	if err != nil {
		return err
	}
	return nil
}
