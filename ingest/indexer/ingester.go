package indexer

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

// indexerIngester is an implementation of domain.Ingester.
type indexerIngester struct {
	pubsubClient domain.Ingester
}

// NewIndexerIngester creates a new indexer ingester.
func NewIndexerIngester(pubsubClient domain.Ingester) domain.Ingester {
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
