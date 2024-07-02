package blockprocessor

import (
	"github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
)

type blockUpdatesSQSBlockProcessStrategy struct {
	sqsGRPCCLient domain.SQSGRPClient

	poolsTransformer domain.PoolsTransformer
	poolExtracter    commondomain.PoolExtracter
}

var _ commondomain.BlockProcessor = &blockUpdatesSQSBlockProcessStrategy{}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *blockUpdatesSQSBlockProcessStrategy) ProcessBlock(ctx types.Context) error {
	// Publish supplies
	if err := f.publishChangedPools(ctx); err != nil {
		return err
	}

	return nil
}

// publishChangedPools publishes the pools that were changed in the block.
func (f *blockUpdatesSQSBlockProcessStrategy) publishChangedPools(ctx types.Context) error {
	// Extract the pools that were changed in the block
	pools, err := f.poolExtracter.ExtractChanged(ctx)
	if err != nil {
		return err
	}

	// Publish the pools
	err = transformAndLoad(ctx, f.poolsTransformer, f.sqsGRPCCLient, pools)
	if err != nil {
		return err
	}

	return nil
}
