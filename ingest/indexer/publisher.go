package indexer

import (
	"context"

	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	service "github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/client"
	"github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

// indexerIngester is an implementation of domain.Ingester.
type indexerPublisher struct {
	pubsubClient service.PubSubClient
}

// NewIndexerIngester creates a new indexer ingester.
func NewIndexerPublisher(pubsubClient service.PubSubClient) domain.Publisher {
	return &indexerPublisher{
		pubsubClient: pubsubClient,
	}
}

// PublishBlock implements domain.Ingester.
func (i *indexerPublisher) PublishBlock(ctx context.Context, block domain.Block) error {
	err := i.pubsubClient.PublishBlock(ctx, block)
	if err != nil {
		return err
	}
	return nil
}

// PublishTransaction implements domain.Ingester.
func (i *indexerPublisher) PublishTransaction(ctx context.Context, txn domain.Transaction) error {
	err := i.pubsubClient.PublishTransaction(ctx, txn)
	if err != nil {
		return err
	}
	return nil
}

// PublishPool implements domain.Ingester.
func (i *indexerPublisher) PublishPool(ctx context.Context, pool domain.Pool) error {
	err := i.pubsubClient.PublishPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

// PublishTokenSupply implements domain.Ingester.
func (i *indexerPublisher) PublishTokenSupply(ctx context.Context, tokenSupply domain.TokenSupply) error {
	err := i.pubsubClient.PublishTokenSupply(ctx, tokenSupply)
	if err != nil {
		return err
	}
	return nil
}

// PublishTokenSupplyOffset implements domain.Ingester.
func (i *indexerPublisher) PublishTokenSupplyOffset(ctx context.Context, tokenSupplyOffset domain.TokenSupplyOffset) error {
	err := i.pubsubClient.PublishTokenSupplyOffset(ctx, tokenSupplyOffset)
	if err != nil {
		return err
	}
	return nil
}

// PublishPools implements domain.Publisher.
func (i *indexerPublisher) PublishPools(ctx context.Context, pools []types.PoolI) error {
	// TODO: consider worker pool.

	result := make(chan error, len(pools))

	// Publish all the pools
	for _, pool := range pools {

		go func(pool poolmanagertypes.PoolI) {
			// Publish the pool
			err := i.PublishPool(ctx, domain.Pool{
				ChainModel: pool,
			})

			result <- err
		}(pool)
	}

	// Wait for all the results
	for i := 0; i < len(pools); i++ {
		err := <-result
		if err != nil {
			return err
		}
	}

	return nil
}

// PublishPair implements domain.Publisher.
func (i *indexerPublisher) PublishPair(ctx context.Context, pair domain.Pair) error {
	err := i.pubsubClient.PublishPair(ctx, pair)
	if err != nil {
		return err
	}
	return nil
}
