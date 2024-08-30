package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
)

type blockUpdatesIndexerBlockProcessStrategy struct {
	client            domain.Publisher
	poolExtractor     commondomain.PoolExtractor
	poolPairPublisher domain.PairPublisher
}

var _ commondomain.BlockProcessor = &blockUpdatesIndexerBlockProcessStrategy{}

// IsFullBlockProcessor implements commondomain.BlockProcessor.
func (f *blockUpdatesIndexerBlockProcessStrategy) IsFullBlockProcessor() bool {
	return false
}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *blockUpdatesIndexerBlockProcessStrategy) ProcessBlock(ctx types.Context) error {
	// Publish supplies
	if err := f.publishCreatedPools(ctx); err != nil {
		return err
	}

	return nil
}

// publishCreatedPools publishes the pools that were created in the block.
func (f *blockUpdatesIndexerBlockProcessStrategy) publishCreatedPools(ctx types.Context) error {
	// Extract the pools that were changed in the block
	blockPools, createdPoolIDs, err := f.poolExtractor.ExtractCreated(ctx)
	if err != nil {
		return err
	}

	pools := blockPools.GetAll()

	// Do nothing if no pools were created, or pool metadata is nil
	if len(createdPoolIDs) == 0 || len(pools) == 0 {
		return nil
	}

	// Publish pool pairs
	if err := f.poolPairPublisher.PublishPoolPairs(ctx, pools, createdPoolIDs); err != nil {
		return err
	}

	// Reset the pool tracker
	f.poolExtractor.ResetPoolTracker(ctx)

	return nil
}
