package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
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
	blockPools, createdPoolIDS, err := f.poolExtractor.ExtractCreated(ctx)
	if err != nil {
		return err
	}

	pools := blockPools.GetAll()

	// Publish pool pairs
	if err := f.poolPairPublisher.PublishPoolPairs(ctx, pools, createdPoolIDS); err != nil {
		return err
	}

	return nil
}
