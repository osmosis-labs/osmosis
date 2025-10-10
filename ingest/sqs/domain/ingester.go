package domain

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ingesttypes "github.com/osmosis-labs/osmosis/v31/ingest/types"

	commondomain "github.com/osmosis-labs/osmosis/v31/ingest/common/domain"
)

// PoolsTransformer is an interface that defines the methods for the pool transformer
type PoolsTransformer interface {
	// Transform processes the pool state, returning pools instrumented with all the necessary chain data.
	// Additionally, returns the taker fee map for every pool denom pair.
	// Returns error if the transformer fails to process pool data.
	Transform(ctx sdk.Context, blockPools commondomain.BlockPools) ([]ingesttypes.PoolI, ingesttypes.TakerFeeMap, error)
}

// SQSGRPClient is an interface that defines the methods for the graceful SQS GRPC client.
// It handles graceful connection management. So that, if a GRPC ingest method returns status.Unavailable,
// the GRPC client will reset the connection and attempt to recreate it before retrying the ingest method.
type SQSGRPClient interface {
	// PushData pushes the height, pools and taker fee data to SQS via GRPC.
	// Returns error if the GRPC client fails to push data.
	// On status.Unavailable, it closes the connection and attempts to re-establish it during the next GRPC call.
	// Note: while there are built-in mechanisms to handle retry such as exponential backoff, they are no suitable for our context.
	// In our context, we would rather continue attempting to repush the data in the next block instead of blocking the system.
	PushData(ctx context.Context, height uint64, pools []ingesttypes.PoolI, takerFeesMap ingesttypes.TakerFeeMap) error
}
