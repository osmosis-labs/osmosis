package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v26/x/poolmanager/types"
)

type blockUpdatesIndexerBlockProcessStrategy struct {
	client                  domain.Publisher
	poolExtractor           commondomain.PoolExtractor
	poolPairPublisher       domain.PairPublisher
	blockUpdateProcessUtils commondomain.BlockUpdateProcessUtilsI
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

// publishChangedPools publishes the pools that were changed in the block.
func (f *blockUpdatesIndexerBlockProcessStrategy) publishCreatedPools(ctx types.Context) error {
	err := f.blockUpdateProcessUtils.ProcessBlockChangeSet()
	if err != nil {
		return err
	}
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

	// Filter pools to include only those with pool IDs found in createdPoolIDs
	filteredPools := []poolmanagertypes.PoolI{}
	for _, pool := range pools {
		if _, exists := createdPoolIDs[pool.GetId()]; exists {
			filteredPools = append(filteredPools, pool)
		}
	}

	// Do nothing if no pools are left after filtering
	if len(filteredPools) == 0 {
		return nil
	}

	// Publish pool pairs
	if err := f.poolPairPublisher.PublishPoolPairs(ctx, filteredPools, createdPoolIDs); err != nil {
		return err
	}

	return nil
}
