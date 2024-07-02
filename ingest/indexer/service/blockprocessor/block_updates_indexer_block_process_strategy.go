package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
)

type blockUpdatesIndexerBlockProcessStrategy struct {
	client        domain.Publisher
	poolExtracter commondomain.PoolExtracter
}

var _ commondomain.BlockProcessor = &blockUpdatesIndexerBlockProcessStrategy{}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *blockUpdatesIndexerBlockProcessStrategy) ProcessBlock(ctx types.Context) error {
	// Publish supplies
	if err := f.publishChangedPools(ctx); err != nil {
		return err
	}

	return nil
}

// publishChangedPools publishes the pools that were changed in the block.
func (f *blockUpdatesIndexerBlockProcessStrategy) publishChangedPools(ctx types.Context) error {
	// Extract the pools that were changed in the block
	pools, err := f.poolExtracter.ExtractChanged(ctx)
	if err != nil {
		return err
	}

	// Publish the pools
	for _, pool := range pools.GetAll() {
		pool := pool

		if err := f.client.PublishPool(ctx, domain.Pool{
			ChainModel: pool,
		}); err != nil {
			return err
		}
	}

	return nil
}
