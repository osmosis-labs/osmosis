package blockprocessor

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v31/ingest/sqs/domain"
)

// transformAndLoad transforms the pools and loads them into the SQS.
func transformAndLoad(ctx sdk.Context, poolsTransformer domain.PoolsTransformer, sqsGRPCClient domain.SQSGRPClient, pools commondomain.BlockPools) error {
	// Transform the pools
	transformedPools, takerFeeMap, err := poolsTransformer.Transform(ctx, pools)
	if err != nil {
		return err
	}

	// load the data
	if err := sqsGRPCClient.PushData(ctx, uint64(ctx.BlockHeight()), transformedPools, takerFeeMap); err != nil {
		return err
	}

	return nil
}
