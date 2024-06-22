package indexer

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	service "github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/client"
)

<<<<<<< HEAD
// indexerIngester is an implementation of domain.Ingester.
=======
// indexerIngester is an implementation of domain.Publisher.
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
type indexerPublisher struct {
	pubsubClient service.PubSubClient
}

<<<<<<< HEAD
// NewIndexerIngester creates a new indexer ingester.
=======
// NewIndexerPublisher creates a new IndexerPublisher with the given PubSubClient
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
func NewIndexerPublisher(pubsubClient service.PubSubClient) domain.Publisher {
	return &indexerPublisher{
		pubsubClient: pubsubClient,
	}
}

<<<<<<< HEAD
// PublishBlock implements domain.Ingester.
=======
// PublishBlock implements domain.Publisher.
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
func (i *indexerPublisher) PublishBlock(ctx context.Context, block domain.Block) error {
	err := i.pubsubClient.PublishBlock(ctx, block)
	if err != nil {
		return err
	}
	return nil
}

<<<<<<< HEAD
// PublishTransaction implements domain.Ingester.
=======
// PublishTransaction implements domain.Publisher.
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
func (i *indexerPublisher) PublishTransaction(ctx context.Context, txn domain.Transaction) error {
	err := i.pubsubClient.PublishTransaction(ctx, txn)
	if err != nil {
		return err
	}
	return nil
}

<<<<<<< HEAD
// PublishPool implements domain.Ingester.
=======
// PublishPool implements domain.Publisher.
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
func (i *indexerPublisher) PublishPool(ctx context.Context, pool domain.Pool) error {
	err := i.pubsubClient.PublishPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

<<<<<<< HEAD
// PublishTokenSupply implements domain.Ingester.
=======
// PublishTokenSupply implements domain.Publisher.
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
func (i *indexerPublisher) PublishTokenSupply(ctx context.Context, tokenSupply domain.TokenSupply) error {
	err := i.pubsubClient.PublishTokenSupply(ctx, tokenSupply)
	if err != nil {
		return err
	}
	return nil
}

<<<<<<< HEAD
// PublishTokenSupplyOffset implements domain.Ingester.
=======
// PublishTokenSupplyOffset implements domain.Publisher.
>>>>>>> 84caa891 (feat: DexScreener (main) (#8411))
func (i *indexerPublisher) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset domain.TokenSupplyOffset) error {
	err := i.pubsubClient.PublishTokenSupplyOffset(ctx, tokenSupplyOffset)
	if err != nil {
		return err
	}
	return nil
}
