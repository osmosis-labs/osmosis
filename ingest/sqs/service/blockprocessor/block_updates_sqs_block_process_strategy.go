package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v27/ingest/sqs/domain"
)

type blockUpdatesSQSBlockProcessStrategy struct {
	sqsGRPCClient domain.SQSGRPClient

	poolsTransformer domain.PoolsTransformer
	poolExtractor    commondomain.PoolExtractor

	transformAndLoadFunc transformAndLoadFunc

	blockUpdateProcessUtils commondomain.BlockUpdateProcessUtilsI
}

// IsFullBlockProcessor implements commondomain.BlockProcessor.
func (f *blockUpdatesSQSBlockProcessStrategy) IsFullBlockProcessor() bool {
	return false
}

var _ commondomain.BlockProcessor = &blockUpdatesSQSBlockProcessStrategy{}

// ProcessBlock implements commondomain.BlockProcessStrategy.
// ProcessBlock extracts, transforms and loads the pools that were changed in the block.
// Returns an error if any of the steps fail.
func (f *blockUpdatesSQSBlockProcessStrategy) ProcessBlock(ctx types.Context) error {
	// Due to new streaming service design, we need to process the writes in the change set all at once here.
	err := f.blockUpdateProcessUtils.ProcessBlockChangeSet()
	if err != nil {
		return err
	}

	// Extract the pools that were changed in the block
	pools, err := f.poolExtractor.ExtractChanged(ctx)
	if err != nil {
		return err
	}

	// Publish the pools
	err = f.transformAndLoadFunc(ctx, f.poolsTransformer, f.sqsGRPCClient, pools)
	if err != nil {
		return err
	}
	return nil
}
